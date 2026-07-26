// Package seed يحمّل كتالوج أسئلة المقابلة ويحافظ على ترتيبه.
package seed

import (
	"fmt"

	"surveyapp/internal/model"
	"surveyapp/internal/store"
)

// فئات الاستهداف. لا وجود لفئات مركّبة عمدًا: كل قسم يخص دورًا واحدًا،
// و all محصورة في الأسئلة الثمانية المشتركة المصممة للمقارنة بين الأدوار.
var (
	all        = []model.Category{} // فارغ = يظهر لكل الفئات
	guard      = []model.Category{model.CatGuard}
	supervisor = []model.Category{model.CatSupervisor}
	area       = []model.Category{model.CatAreaManager}
	manager    = []model.Category{model.CatCompanyManager}
)

// Apply يوحّد قاعدة البيانات مع الكتالوج ويعيد عدد الأسئلة المضافة.
//
// المطابقة بنص السؤال: القاعدة القائمة عليها إجابات تستقبل الجديد وحده،
// والأسئلة التي خرجت من الكتالوج تُعطَّل تلقائيًا مع بقاء إجاباتها، ولا
// يمسّ ذلك الأسئلة المضافة يدويًا من اللوحة.
func Apply(s *store.Store) (int, error) {
	existing, err := s.Questions(true)
	if err != nil {
		return 0, err
	}
	byText := make(map[string]model.Question, len(existing))
	for _, q := range existing {
		byText[q.Text] = q
	}

	catalogue := Questions()
	order := make([]int64, 0, len(catalogue))
	keep := make(map[int64]bool, len(catalogue))
	added := 0

	for _, q := range catalogue {
		q.FromCatalog = true
		if prev, ok := byText[q.Text]; ok {
			// السؤال موجود: نوحّده مع الكتالوج لأن الكتالوج هو المصدر الموثوق.
			if err := s.SyncCatalogQuestion(prev.ID, q); err != nil {
				return added, fmt.Errorf("تحديث سؤال %q: %w", q.Text, err)
			}
			order = append(order, prev.ID)
			keep[prev.ID] = true
			continue
		}
		id, err := s.CreateQuestion(q)
		if err != nil {
			return added, fmt.Errorf("إضافة سؤال %q: %w", q.Text, err)
		}
		order = append(order, id)
		keep[id] = true
		added++
	}

	if _, err := s.RetireCatalogQuestions(keep); err != nil {
		return added, fmt.Errorf("تعطيل الأسئلة الخارجة من الكتالوج: %w", err)
	}

	// الأسئلة المضافة من اللوحة تبقى بعد أسئلة الكتالوج بترتيبها الحالي.
	for _, q := range existing {
		if !keep[q.ID] {
			order = append(order, q.ID)
		}
	}
	if err := s.ReorderQuestions(order); err != nil {
		return added, fmt.Errorf("ترتيب الأسئلة: %w", err)
	}
	return added, s.MarkSeeded()
}

func q(section, text string, kind model.Kind, cats []model.Category, required bool, opts ...string) model.Question {
	if opts == nil {
		opts = []string{}
	}
	return model.Question{
		Text:       text,
		Kind:       kind,
		Section:    section,
		Categories: cats,
		Options:    opts,
		Required:   required,
	}
}

// Questions كتالوج أسئلة المقابلة: المشترك أولًا، ثم أقسام كل دور مجمّعة.
func Questions() []model.Question {
	groups := [][]model.Question{
		commonQuestions(),

		// الحارس — التنفيذ
		shiftStartQuestions(),
		taskQuestions(),
		patrolQuestions(),
		checkQuestions(),
		reportQuestions(),
		sosQuestions(),
		myRequestQuestions(),
		notifyQuestions(),
		networkQuestions(),
		usabilityQuestions(),

		// المشرف الميداني — إدارة الحراس داخل المواقع
		attendWatchQuestions(),
		assignQuestions(),
		patrolWatchQuestions(),
		reportWatchQuestions(),
		approveQuestions(),
		opsRoomQuestions(),
		siteVisitQuestions(),
		shiftMapQuestions(),

		// المشرف العام — عدة مشرفين وعدة مواقع
		sitesQuestions(),
		mapQuestions(),
		kpiQuestions(),
		criticalQuestions(),
		resourceQuestions(),
		decisionQuestions(),

		// مدير الشركة — معطّل حاليًا
		profitQuestions(),
		contractQuestions(),
		billingQuestions(),
		payrollQuestions(),
		qualityQuestions(),
		execQuestions(),
	}
	out := []model.Question{}
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}
