// Package seed يحمّل كتالوج أسئلة المقابلة ويحافظ على ترتيبه.
package seed

import (
	"fmt"

	"surveyapp/internal/model"
	"surveyapp/internal/store"
)

// فئات الاستهداف. الفصل مبني على المنصّة والصلاحية معًا: الميدان جوال فقط
// (حارس ومشرف ميداني)، والمشرف العام يجمع بين الميدان والإشراف على المواقع،
// ومدير الشركة مستخدم ويب مكتبي لا ميداني.
var (
	all     = []model.Category{} // فارغ = يظهر لكل الفئات
	field   = []model.Category{model.CatGuard, model.CatSupervisor}
	staff   = []model.Category{model.CatGuard, model.CatSupervisor, model.CatAreaManager}
	sup     = []model.Category{model.CatSupervisor, model.CatAreaManager}
	mgmt    = []model.Category{model.CatAreaManager, model.CatCompanyManager}
	manager = []model.Category{model.CatCompanyManager}
)

// Apply يضمن وجود كل أسئلة الكتالوج في قاعدة البيانات ويعيد عدد المضاف.
//
// المطابقة بنص السؤال لا بمعرّفه، فالتشغيل متكرر بلا تكرار للأسئلة: قاعدة
// قائمة عليها إجابات تستقبل الأسئلة الجديدة وحدها. والسؤال المحذوف من اللوحة
// لا يعود، لأن البحث يشمل المحذوفين منطقيًا.
func Apply(s *store.Store) (int, error) {
	existing, err := s.Questions(true)
	if err != nil {
		return 0, err
	}
	byText := make(map[string]int64, len(existing))
	for _, q := range existing {
		byText[q.Text] = q.ID
	}

	catalogue := Questions()
	order := make([]int64, 0, len(catalogue))
	added := 0
	for _, q := range catalogue {
		if id, ok := byText[q.Text]; ok {
			order = append(order, id)
			continue
		}
		id, err := s.CreateQuestion(q)
		if err != nil {
			return added, fmt.Errorf("إضافة سؤال %q: %w", q.Text, err)
		}
		byText[q.Text] = id
		order = append(order, id)
		added++
	}

	// الأسئلة المضافة من اللوحة تبقى بعد أسئلة الكتالوج بترتيبها الحالي.
	inCatalogue := make(map[int64]bool, len(order))
	for _, id := range order {
		inCatalogue[id] = true
	}
	for _, q := range existing {
		if !inCatalogue[q.ID] {
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

// Questions كتالوج أسئلة المقابلة بالترتيب المعروض: السياق العام أولًا،
// ثم رحلات العمل بترتيب حدوثها، ثم أسئلة الواجهة، ثم أدوار الإشراف والإدارة.
func Questions() []model.Question {
	groups := [][]model.Question{
		dayQuestions(),
		deviceQuestions(),
		physicalQuestions(),
		authQuestions(),
		briefQuestions(),
		attendQuestions(),
		queueQuestions(),
		patrolQuestions(),
		checkQuestions(),
		reportQuestions(),
		sosQuestions(),
		custodyQuestions(),
		requestQuestions(),
		directiveQuestions(),
		commQuestions(),
		notifyQuestions(),
		offlineQuestions(),
		homeQuestions(),
		navQuestions(),
		inputQuestions(),
		langQuestions(),
		errorQuestions(),
		perfQuestions(),
		a11yQuestions(),
		trustQuestions(),
		trainQuestions(),
		teamQuestions(),
		reviewQuestions(),
		gsQuestions(),
		webDailyQuestions(),
		webDataQuestions(),
		financeQuestions(),
		hrQuestions(),
		clientQuestions(),
		hiddenQuestions(),
		priorityQuestions(),
		finalQuestions(),
	}
	out := []model.Question{}
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}
