package seed

import (
	"path/filepath"
	"strings"
	"testing"

	"surveyapp/internal/model"
	"surveyapp/internal/store"
)

func newStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "seed.db"))
	if err != nil {
		t.Fatalf("فتح قاعدة الاختبار: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// نص السؤال هو مفتاح المطابقة عند الترقية، فتكراره يعني إضافة مزدوجة.
func TestNoDuplicateQuestionText(t *testing.T) {
	seen := map[string]string{}
	for _, q := range Questions() {
		if prev, ok := seen[q.Text]; ok {
			t.Errorf("سؤال مكرر بين %q و%q: %s", prev, q.Section, q.Text)
		}
		seen[q.Text] = q.Section
	}
}

func TestEveryQuestionIsWellFormed(t *testing.T) {
	for _, q := range Questions() {
		if strings.TrimSpace(q.Text) == "" {
			t.Errorf("سؤال بلا نص في قسم %q", q.Section)
		}
		if q.Section == "" {
			t.Errorf("سؤال بلا قسم: %s", q.Text)
		}
		if !model.ValidKind(q.Kind) {
			t.Errorf("نوع غير مدعوم %q في: %s", q.Kind, q.Text)
		}
		if model.NeedsOptions(q.Kind) && len(q.Options) < 2 {
			t.Errorf("نوع %q يحتاج خيارين على الأقل: %s", q.Kind, q.Text)
		}
		if !model.NeedsOptions(q.Kind) && len(q.Options) > 0 {
			t.Errorf("نوع %q لا يأخذ خيارات: %s", q.Kind, q.Text)
		}
		for _, c := range q.Categories {
			if !model.ValidCategory(c) {
				t.Errorf("فئة غير معروفة %q في: %s", c, q.Text)
			}
		}
	}
}

// كل فئة يجب أن ترى أسئلة، ومدير الشركة مستخدم ويب فلا تُعرض له
// أقسام ميدانية مثل الدوريات ونقاط التفتيش والعهد.
func TestEachCategoryGetsRelevantQuestions(t *testing.T) {
	fieldOnly := map[string]bool{
		secPatrol: true, secCheck: true, secCustody: true,
		secSOS: true, secBrief: true, secAttend: true,
	}
	for _, c := range model.AllCategories {
		count := 0
		for _, q := range Questions() {
			if !q.AppliesTo(c) {
				continue
			}
			count++
			if c == model.CatCompanyManager && fieldOnly[q.Section] {
				t.Errorf("مدير الشركة لا يجب أن يرى قسمًا ميدانيًا (%s): %s", q.Section, q.Text)
			}
		}
		if count < 30 {
			t.Errorf("فئة %s ترى %d سؤالًا فقط", model.CategoryLabel(c), count)
		}
	}
}

// الأسئلة الإلزامية قليلة عمدًا: الاستبيان طويل، وكثرة الإلزام تزيد الانسحاب.
func TestRequiredQuestionsStayFew(t *testing.T) {
	required := 0
	for _, q := range Questions() {
		if q.Required {
			required++
		}
	}
	if required > 10 {
		t.Errorf("عدد الأسئلة الإلزامية %d — أكثر مما يحتمله استبيان بهذا الطول", required)
	}
}

func TestApplyIsIdempotent(t *testing.T) {
	s := newStore(t)

	first, err := Apply(s)
	if err != nil {
		t.Fatalf("التحميل الأول: %v", err)
	}
	if first != len(Questions()) {
		t.Fatalf("توقعنا %d سؤالًا، حُمّل %d", len(Questions()), first)
	}

	second, err := Apply(s)
	if err != nil {
		t.Fatalf("التحميل الثاني: %v", err)
	}
	if second != 0 {
		t.Fatalf("التحميل المتكرر أضاف %d سؤالًا — كان يجب ألا يضيف شيئًا", second)
	}

	stored, _ := s.Questions(false)
	if len(stored) != len(Questions()) {
		t.Fatalf("عدد الأسئلة المخزّنة %d لا يطابق الكتالوج %d", len(stored), len(Questions()))
	}
}

// الترقية على قاعدة قائمة تضيف الجديد وحده وتحافظ على الإجابات المجموعة.
func TestApplyAddsOnlyNewQuestionsToExistingDatabase(t *testing.T) {
	s := newStore(t)

	old := Questions()[0]
	id, err := s.CreateQuestion(old)
	if err != nil {
		t.Fatalf("إضافة سؤال قديم: %v", err)
	}
	sess, _ := s.CreateSession(model.CatGuard, "فيصل")
	if err := s.SaveAnswer(sess.ID, id, []byte(`"إجابة قديمة"`)); err != nil {
		t.Fatalf("حفظ إجابة: %v", err)
	}

	added, err := Apply(s)
	if err != nil {
		t.Fatalf("الترقية: %v", err)
	}
	if added != len(Questions())-1 {
		t.Fatalf("توقعنا إضافة %d، أُضيف %d", len(Questions())-1, added)
	}

	answers, _ := s.SessionAnswers(sess.ID)
	if len(answers) != 1 || string(answers[id]) != `"إجابة قديمة"` {
		t.Fatalf("الإجابة القديمة تغيّرت أو ضاعت: %v", answers)
	}
}

// السؤال المحذوف من اللوحة يجب ألا يعود مع أي ترقية لاحقة.
func TestApplyDoesNotResurrectDeletedQuestions(t *testing.T) {
	s := newStore(t)
	if _, err := Apply(s); err != nil {
		t.Fatalf("التحميل الأول: %v", err)
	}
	qs, _ := s.Questions(false)
	if err := s.DeleteQuestion(qs[0].ID); err != nil {
		t.Fatalf("حذف سؤال: %v", err)
	}

	if _, err := Apply(s); err != nil {
		t.Fatalf("الترقية بعد الحذف: %v", err)
	}
	after, _ := s.Questions(false)
	if len(after) != len(qs)-1 {
		t.Fatalf("السؤال المحذوف عاد: %d بدل %d", len(after), len(qs)-1)
	}
}

// الترتيب المعروض يجب أن يطابق ترتيب الكتالوج حتى تبقى الأقسام متجاورة.
func TestApplyKeepsCatalogueOrder(t *testing.T) {
	s := newStore(t)
	if _, err := Apply(s); err != nil {
		t.Fatalf("التحميل: %v", err)
	}
	stored, _ := s.Questions(false)
	for i, q := range Questions() {
		if stored[i].Text != q.Text {
			t.Fatalf("الترتيب اختلف عند الموضع %d: %q بدل %q", i, stored[i].Text, q.Text)
		}
	}
}
