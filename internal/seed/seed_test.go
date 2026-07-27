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

// القاعدة الحاكمة: كل سؤال يخص دورًا واحدًا، إلا الأسئلة المشتركة.
// السؤال الموجّه لدورين يعني أننا نسأل المشرف عمّا ينفّذه الحارس.
func TestEveryQuestionBelongsToExactlyOneRole(t *testing.T) {
	for _, q := range Questions() {
		if q.Section == secCommon {
			// المشترك للأدوار الوظيفية وحدها؛ «داشبورد» مسار تصميم
			// لا يُسأل صاحبه عن يوم عمله في الميدان.
			if len(q.Categories) != len(model.JobRoles) {
				t.Errorf("السؤال المشترك يجب أن يكون للأدوار الوظيفية الأربعة: %s", q.Text)
			}
			if q.AppliesTo(model.CatDashboard) {
				t.Errorf("دور الداشبورد لا يجيب على الأسئلة المشتركة: %s", q.Text)
			}
			continue
		}
		if len(q.Categories) != 1 {
			t.Errorf("سؤال موجّه لـ%d أدوار بدل واحد (%s): %s",
				len(q.Categories), q.Section, q.Text)
		}
	}
}

// المشترك ثمانية أسئلة بالضبط — أي زيادة تُميّع المقارنة بين الأدوار.
func TestCommonQuestionsAreExactlyEight(t *testing.T) {
	n := 0
	for _, q := range Questions() {
		if q.Section == secCommon {
			n++
		}
	}
	if n != 8 {
		t.Errorf("الأسئلة المشتركة %d، والمتفق عليه ٨", n)
	}
}

// كل قسم يخص دورًا واحدًا: لا يجوز أن يحوي قسم أسئلة لأدوار مختلفة.
func TestEachSectionBelongsToOneRole(t *testing.T) {
	owner := map[string]model.Category{}
	for _, q := range Questions() {
		if q.Section == secCommon || len(q.Categories) != 1 {
			continue
		}
		c := q.Categories[0]
		if prev, ok := owner[q.Section]; ok && prev != c {
			t.Errorf("قسم %q يخلط بين %s و%s",
				q.Section, model.CategoryLabel(prev), model.CategoryLabel(c))
		}
		owner[q.Section] = c
	}
}

// أسئلة التنفيذ الميداني للحارس وحده: لا يُسأل المشرف عن QR ولا NFC،
// ولا يُسأل المشرف العام عن المشي في الجولة.
func TestFieldExecutionIsGuardOnly(t *testing.T) {
	guardOnly := map[string]bool{
		secShiftStart: true, secPatrol: true, secCheck: true,
		secSOS: true, secTasks: true, secMyRequests: true,
	}
	for _, q := range Questions() {
		if !guardOnly[q.Section] {
			continue
		}
		for _, c := range q.Categories {
			if c != model.CatGuard {
				t.Errorf("قسم التنفيذ %q معروض على %s: %s",
					q.Section, model.CategoryLabel(c), q.Text)
			}
		}
	}
}

// مدير الشركة لا يُسأل عن العمل الميداني اليومي.
func TestCompanyManagerHasNoFieldQuestions(t *testing.T) {
	fieldSections := map[string]bool{
		secShiftStart: true, secPatrol: true, secCheck: true, secReport: true,
		secSOS: true, secTasks: true, secNotify: true, secNetwork: true,
		secUsability: true, secMyRequests: true,
	}
	for _, q := range Questions() {
		if fieldSections[q.Section] && q.AppliesTo(model.CatCompanyManager) {
			t.Errorf("مدير الشركة يرى سؤالًا ميدانيًا (%s): %s", q.Section, q.Text)
		}
	}
}

func TestRequiredQuestionsStayFew(t *testing.T) {
	required := 0
	for _, q := range Questions() {
		if q.Required {
			required++
		}
	}
	if required > 15 {
		t.Errorf("عدد الأسئلة الإلزامية %d — أكثر مما يحتمله الاستبيان", required)
	}
}

// الإلزامي يُقاس لكل دور لا إجماليًا: ما يهم هو ما يواجهه المشارك الواحد.
func TestRequiredPerRoleStaysFew(t *testing.T) {
	for _, c := range model.AllCategories {
		n := 0
		for _, q := range Questions() {
			if q.Required && q.AppliesTo(c) {
				n++
			}
		}
		if n > 6 {
			t.Errorf("%s يواجه %d سؤالًا إلزاميًا — كثير", model.CategoryLabel(c), n)
		}
	}
}

// حجم الاستبيان لكل دور يجب أن يبقى قابلًا للإنجاز في جلسة واحدة.
// الحارس أضيق سقفًا لأنه يجيب واقفًا وسط وردية، لا جالسًا على مكتب.
func TestPerRoleLoadStaysReasonable(t *testing.T) {
	limits := map[model.Category]int{
		model.CatGuard:          60,
		model.CatSupervisor:     80,
		model.CatAreaManager:    80,
		model.CatCompanyManager: 80,
		model.CatDashboard:      40,
	}
	for _, c := range model.AllCategories {
		n := 0
		for _, q := range Questions() {
			if q.AppliesTo(c) {
				n++
			}
		}
		if n < 30 {
			t.Errorf("%s يرى %d سؤالًا فقط — قليل جدًا", model.CategoryLabel(c), n)
		}
		if n > limits[c] {
			t.Errorf("%s يرى %d سؤالًا والحد %d — أطول من جلسة واحدة",
				model.CategoryLabel(c), n, limits[c])
		}
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
		t.Fatalf("التحميل المتكرر أضاف %d سؤالًا", second)
	}

	stored, _ := s.Questions(false)
	if len(stored) != len(Questions()) {
		t.Fatalf("عدد الأسئلة المخزّنة %d لا يطابق الكتالوج %d", len(stored), len(Questions()))
	}
}

// السؤال الذي خرج من الكتالوج يُعطَّل، وإجابته تبقى.
func TestApplyRetiresQuestionsDroppedFromCatalogue(t *testing.T) {
	s := newStore(t)

	old := model.Question{
		Text: "سؤال قديم خرج من الكتالوج", Kind: model.KindLongText,
		Section: "قسم ملغى", FromCatalog: true,
	}
	id, err := s.CreateQuestion(old)
	if err != nil {
		t.Fatalf("إضافة سؤال قديم: %v", err)
	}
	sess, _ := s.CreateSession(model.CatGuard, "فيصل")
	if err := s.SaveAnswer(sess.ID, id, []byte(`"إجابة قديمة"`)); err != nil {
		t.Fatalf("حفظ إجابة: %v", err)
	}

	if _, err := Apply(s); err != nil {
		t.Fatalf("الترقية: %v", err)
	}

	got, err := s.Question(id)
	if err != nil {
		t.Fatalf("قراءة السؤال القديم: %v", err)
	}
	if !got.Deleted {
		t.Fatalf("السؤال الخارج من الكتالوج يجب أن يُعطَّل")
	}
	answers, _ := s.SessionAnswers(sess.ID)
	if len(answers) != 1 {
		t.Fatalf("إجابة السؤال المعطّل ضاعت")
	}
}

// الأسئلة المضافة يدويًا من اللوحة لا تُعطَّل مع الترقية.
func TestApplyKeepsManuallyAddedQuestions(t *testing.T) {
	s := newStore(t)
	// تحميل أول حتى ينتهي ترحيل توسيم الأسئلة السابقة قبل الإضافة اليدوية.
	if _, err := Apply(s); err != nil {
		t.Fatalf("التحميل الأول: %v", err)
	}
	id, err := s.CreateQuestion(model.Question{
		Text: "سؤال أضافه الأدمن", Kind: model.KindLongText, Section: "إضافات",
	})
	if err != nil {
		t.Fatalf("إضافة سؤال يدوي: %v", err)
	}

	if _, err := Apply(s); err != nil {
		t.Fatalf("الترقية: %v", err)
	}

	got, _ := s.Question(id)
	if got.Deleted {
		t.Fatalf("السؤال المضاف يدويًا عُطِّل بالخطأ")
	}
}

// تعديل فئة سؤال في الكتالوج يجب أن يصل إلى القاعدة القائمة.
func TestApplySyncsChangedCategories(t *testing.T) {
	s := newStore(t)
	first := Questions()[len(commonQuestions())] // أول سؤال خاص بالحارس
	stale := first
	stale.Categories = []model.Category{model.CatCompanyManager}
	id, err := s.CreateQuestion(stale)
	if err != nil {
		t.Fatalf("إضافة سؤال بفئة قديمة: %v", err)
	}

	if _, err := Apply(s); err != nil {
		t.Fatalf("الترقية: %v", err)
	}

	got, _ := s.Question(id)
	if len(got.Categories) != 1 || got.Categories[0] != first.Categories[0] {
		t.Fatalf("الفئة لم تُوحَّد مع الكتالوج: %v", got.Categories)
	}
}

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

// تعطيل الأقسام قرار إداري يجب أن ينجو من إعادة النشر.
func TestApplyPreservesDisabledSections(t *testing.T) {
	s := newStore(t)
	if _, err := Apply(s); err != nil {
		t.Fatalf("التحميل الأول: %v", err)
	}
	if _, err := s.SetSectionActive(secPatrol, false); err != nil {
		t.Fatalf("تعطيل قسم: %v", err)
	}
	before, _ := s.Questions(false)

	if _, err := Apply(s); err != nil {
		t.Fatalf("إعادة النشر: %v", err)
	}
	after, _ := s.Questions(false)
	if len(after) != len(before) {
		t.Fatalf("إعادة النشر أعادت تفعيل قسم معطّل: %d بدل %d", len(after), len(before))
	}
}

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
