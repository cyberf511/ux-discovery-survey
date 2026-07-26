package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"surveyapp/internal/model"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("فتح قاعدة الاختبار: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func addQuestion(t *testing.T, s *Store, text string, cats []model.Category) int64 {
	t.Helper()
	id, err := s.CreateQuestion(model.Question{
		Text:       text,
		Kind:       model.KindLongText,
		Categories: cats,
	})
	if err != nil {
		t.Fatalf("إضافة سؤال: %v", err)
	}
	return id
}

func TestSessionAndAnswerRoundTrip(t *testing.T) {
	s := newTestStore(t)
	qid := addQuestion(t, s, "صف يوم عملك", nil)

	sess, err := s.CreateSession(model.CatGuard, "فيصل")
	if err != nil {
		t.Fatalf("إنشاء جلسة: %v", err)
	}
	if sess.ID == "" || sess.StartedAt == "" {
		t.Fatalf("جلسة ناقصة: %+v", sess)
	}

	if err := s.SaveAnswer(sess.ID, qid, json.RawMessage(`"أبدأ بالتسليم"`)); err != nil {
		t.Fatalf("حفظ إجابة: %v", err)
	}
	// الحفظ مرة ثانية يحدّث القيمة ولا ينشئ صفًا جديدًا.
	if err := s.SaveAnswer(sess.ID, qid, json.RawMessage(`"أبدأ بالجولة"`)); err != nil {
		t.Fatalf("تحديث إجابة: %v", err)
	}

	answers, err := s.SessionAnswers(sess.ID)
	if err != nil {
		t.Fatalf("قراءة الإجابات: %v", err)
	}
	if len(answers) != 1 {
		t.Fatalf("توقعنا إجابة واحدة، وجدنا %d", len(answers))
	}
	if got := string(answers[qid]); got != `"أبدأ بالجولة"` {
		t.Fatalf("القيمة لم تُحدَّث: %s", got)
	}
}

// المشارك لازم يقدر يستأنف جلسته بكود قصير، بلا اعتماد على تخزين المتصفح.
func TestResumeSessionByCode(t *testing.T) {
	s := newTestStore(t)
	sess, err := s.CreateSession(model.CatGuard, "فيصل")
	if err != nil {
		t.Fatalf("إنشاء جلسة: %v", err)
	}
	if sess.Code == "" {
		t.Fatalf("الجلسة يجب أن تحمل كود متابعة")
	}

	got, err := s.SessionByCode(sess.Code)
	if err != nil {
		t.Fatalf("استئناف بالكود: %v", err)
	}
	if got.ID != sess.ID {
		t.Fatalf("الكود أعاد جلسة خاطئة: %s بدل %s", got.ID, sess.ID)
	}

	// الكود غير حسّاس لحالة الأحرف ولا للمسافات، لأن المشارك يكتبه بيده.
	if got, err = s.SessionByCode("  " + strings.ToLower(sess.Code) + " "); err != nil || got.ID != sess.ID {
		t.Fatalf("الكود يجب أن يُقبل بأي حالة أحرف: %v", err)
	}

	if _, err := s.SessionByCode("ZZZZZZ"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("كود مجهول يجب أن يعيد ErrNotFound، وجدنا %v", err)
	}
}

// قاعدة بيانات أُنشئت قبل إضافة أكواد المتابعة يجب أن تُرقَّى بلا انهيار،
// وأن تحصل جلساتها القائمة على أكواد.
func TestMigratesLegacyDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")

	legacy, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("فتح قاعدة قديمة: %v", err)
	}
	_, err = legacy.Exec(`
		CREATE TABLE sessions (
		  id          TEXT PRIMARY KEY,
		  category    TEXT NOT NULL,
		  name        TEXT NOT NULL DEFAULT '',
		  started_at  TEXT NOT NULL,
		  finished_at TEXT NOT NULL DEFAULT ''
		);
		INSERT INTO sessions(id, category, name, started_at) VALUES('old1', 'guard', 'فيصل', '2026-07-20T08:00:00Z');`)
	if err != nil {
		t.Fatalf("تهيئة القاعدة القديمة: %v", err)
	}
	legacy.Close()

	s, err := Open(path)
	if err != nil {
		t.Fatalf("ترقية القاعدة القديمة فشلت: %v", err)
	}
	defer s.Close()

	sess, err := s.Session("old1")
	if err != nil {
		t.Fatalf("قراءة جلسة قديمة: %v", err)
	}
	if sess.Name != "فيصل" {
		t.Fatalf("بيانات الجلسة القديمة ضاعت: %+v", sess)
	}
	if sess.Code == "" {
		t.Fatalf("الجلسة القديمة يجب أن تحصل على كود متابعة")
	}
	if got, err := s.SessionByCode(sess.Code); err != nil || got.ID != "old1" {
		t.Fatalf("الاستئناف بكود الجلسة القديمة فشل: %v", err)
	}
}

func TestSessionCodesAreUnique(t *testing.T) {
	s := newTestStore(t)
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		sess, err := s.CreateSession(model.CatGuard, "")
		if err != nil {
			t.Fatalf("إنشاء جلسة %d: %v", i, err)
		}
		if seen[sess.Code] {
			t.Fatalf("تكرّر كود المتابعة: %s", sess.Code)
		}
		seen[sess.Code] = true
	}
}

// الأكواد تُقرأ وتُكتب يدويًا، فلا تحتمل حروفًا متشابهة الشكل.
func TestSessionCodeAvoidsAmbiguousCharacters(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 30; i++ {
		sess, _ := s.CreateSession(model.CatGuard, "")
		if strings.ContainsAny(sess.Code, "OI01L") {
			t.Fatalf("الكود يحوي حرفًا ملتبسًا: %s", sess.Code)
		}
	}
}

func TestSaveAnswerRejectedWhenClosed(t *testing.T) {
	s := newTestStore(t)
	qid := addQuestion(t, s, "سؤال", nil)
	sess, _ := s.CreateSession(model.CatGuard, "")

	if err := s.SetOpen(false); err != nil {
		t.Fatalf("قفل الاستبيان: %v", err)
	}
	err := s.SaveAnswer(sess.ID, qid, json.RawMessage(`"متأخر"`))
	if !errors.Is(err, ErrSurveyClosed) {
		t.Fatalf("توقعنا ErrSurveyClosed، وجدنا %v", err)
	}
	if _, err := s.CreateSession(model.CatGuard, ""); !errors.Is(err, ErrSurveyClosed) {
		t.Fatalf("الجلسات يجب أن تُرفض أيضًا بعد القفل، وجدنا %v", err)
	}
}

func TestSaveAnswerUnknownSessionOrQuestion(t *testing.T) {
	s := newTestStore(t)
	qid := addQuestion(t, s, "سؤال", nil)
	sess, _ := s.CreateSession(model.CatGuard, "")

	if err := s.SaveAnswer("لا-يوجد", qid, json.RawMessage(`"x"`)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("توقعنا ErrNotFound لجلسة مجهولة، وجدنا %v", err)
	}
	if err := s.SaveAnswer(sess.ID, 9999, json.RawMessage(`"x"`)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("توقعنا ErrNotFound لسؤال مجهول، وجدنا %v", err)
	}
}

func TestQuestionsForFiltersByCategory(t *testing.T) {
	s := newTestStore(t)
	addQuestion(t, s, "سؤال للجميع", nil)
	addQuestion(t, s, "سؤال ميداني", []model.Category{model.CatGuard})

	guard, err := s.QuestionsFor(model.CatGuard)
	if err != nil {
		t.Fatalf("أسئلة الحارس: %v", err)
	}
	if len(guard) != 2 {
		t.Fatalf("الحارس يجب أن يرى سؤالين، رأى %d", len(guard))
	}

	manager, err := s.QuestionsFor(model.CatCompanyManager)
	if err != nil {
		t.Fatalf("أسئلة المدير: %v", err)
	}
	if len(manager) != 1 {
		t.Fatalf("المدير يجب أن يرى سؤالًا واحدًا، رأى %d", len(manager))
	}
}

func TestDeleteQuestionIsSoftAndKeepsAnswers(t *testing.T) {
	s := newTestStore(t)
	qid := addQuestion(t, s, "سؤال قديم", nil)
	sess, _ := s.CreateSession(model.CatGuard, "")
	if err := s.SaveAnswer(sess.ID, qid, json.RawMessage(`"إجابة"`)); err != nil {
		t.Fatalf("حفظ إجابة: %v", err)
	}
	if err := s.DeleteQuestion(qid); err != nil {
		t.Fatalf("حذف سؤال: %v", err)
	}

	visible, _ := s.Questions(false)
	if len(visible) != 0 {
		t.Fatalf("السؤال المحذوف يجب ألا يظهر، ظهر %d", len(visible))
	}
	withDeleted, _ := s.Questions(true)
	if len(withDeleted) != 1 || !withDeleted[0].Deleted {
		t.Fatalf("السؤال المحذوف يجب أن يبقى مخزّنًا: %+v", withDeleted)
	}
	answers, _ := s.SessionAnswers(sess.ID)
	if len(answers) != 1 {
		t.Fatalf("الإجابة يجب أن تبقى بعد حذف السؤال")
	}
}

func TestReorderQuestions(t *testing.T) {
	s := newTestStore(t)
	a := addQuestion(t, s, "الأول", nil)
	b := addQuestion(t, s, "الثاني", nil)

	if err := s.ReorderQuestions([]int64{b, a}); err != nil {
		t.Fatalf("إعادة الترتيب: %v", err)
	}
	qs, _ := s.Questions(false)
	if qs[0].ID != b || qs[1].ID != a {
		t.Fatalf("الترتيب لم يتغيّر: %d ثم %d", qs[0].ID, qs[1].ID)
	}
}

func TestCategoryStatsCountsFinished(t *testing.T) {
	s := newTestStore(t)
	done, _ := s.CreateSession(model.CatGuard, "أ")
	s.CreateSession(model.CatGuard, "ب")
	if err := s.FinishSession(done.ID); err != nil {
		t.Fatalf("إنهاء الجلسة: %v", err)
	}

	stats, err := s.CategoryStats()
	if err != nil {
		t.Fatalf("الإحصاءات: %v", err)
	}
	if len(stats) != len(model.AllCategories) {
		t.Fatalf("توقعنا صفًا لكل فئة، وجدنا %d", len(stats))
	}
	for _, st := range stats {
		if st.Category != model.CatGuard {
			continue
		}
		if st.Started != 2 || st.Finished != 1 {
			t.Fatalf("إحصاء الحارس خاطئ: %+v", st)
		}
	}
}

func TestSessionsFilteredByCategory(t *testing.T) {
	s := newTestStore(t)
	s.CreateSession(model.CatGuard, "أ")
	s.CreateSession(model.CatCompanyManager, "ب")

	all, _ := s.Sessions("")
	if len(all) != 2 {
		t.Fatalf("توقعنا جلستين، وجدنا %d", len(all))
	}
	guards, _ := s.Sessions(model.CatGuard)
	if len(guards) != 1 || guards[0].Name != "أ" {
		t.Fatalf("تصفية الفئة لم تعمل: %+v", guards)
	}
}
