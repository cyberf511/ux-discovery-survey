// Package model يعرّف الأنواع المشتركة بين طبقات التطبيق.
package model

import "encoding/json"

// Kind نوع السؤال.
type Kind string

const (
	KindLongText     Kind = "long_text"
	KindShortText    Kind = "short_text"
	KindBoolean      Kind = "boolean"
	KindSingleChoice Kind = "single_choice"
	KindMultiChoice  Kind = "multi_choice"
	KindRanking      Kind = "ranking"
	KindScale        Kind = "scale_1_5"
)

// ValidKind يتحقق أن النوع مدعوم.
func ValidKind(k Kind) bool {
	switch k {
	case KindLongText, KindShortText, KindBoolean, KindSingleChoice,
		KindMultiChoice, KindRanking, KindScale:
		return true
	}
	return false
}

// NeedsOptions يحدد ما إذا كان النوع يتطلب قائمة خيارات.
func NeedsOptions(k Kind) bool {
	switch k {
	case KindSingleChoice, KindMultiChoice, KindRanking:
		return true
	}
	return false
}

// Category فئة المشارك.
type Category string

const (
	CatGuard          Category = "guard"
	CatSupervisor     Category = "supervisor"
	CatAreaManager    Category = "area_manager"
	CatCompanyManager Category = "company_manager"
)

// AllCategories كل الفئات بالترتيب المعروض.
var AllCategories = []Category{CatGuard, CatSupervisor, CatAreaManager, CatCompanyManager}

// CategoryLabels الأسماء العربية المعروضة للفئات.
var CategoryLabels = map[Category]string{
	CatGuard:          "حارس",
	CatSupervisor:     "مشرف ميداني",
	CatAreaManager:    "مشرف عام",
	CatCompanyManager: "مدير شركة",
}

// ValidCategory يتحقق أن الفئة معروفة.
func ValidCategory(c Category) bool {
	_, ok := CategoryLabels[c]
	return ok
}

// CategoryLabel يعيد الاسم العربي للفئة، أو النص الخام إن كانت غير معروفة.
func CategoryLabel(c Category) string {
	if l, ok := CategoryLabels[c]; ok {
		return l
	}
	return string(c)
}

// Question سؤال في الاستبيان.
type Question struct {
	ID         int64      `json:"id"`
	Text       string     `json:"text"`
	Kind       Kind       `json:"kind"`
	Section    string     `json:"section"`
	Position   int        `json:"position"`
	Categories []Category `json:"categories"`
	Options    []string   `json:"options"`
	Required   bool       `json:"required"`
	Deleted    bool       `json:"deleted"`
}

// AppliesTo يحدد ما إذا كان السؤال موجّهًا لهذه الفئة.
func (q Question) AppliesTo(c Category) bool {
	if len(q.Categories) == 0 {
		return true
	}
	for _, qc := range q.Categories {
		if qc == c {
			return true
		}
	}
	return false
}

// Session جلسة مشارك واحد.
type Session struct {
	ID string `json:"id"`
	// Code كود متابعة قصير يكتبه المشارك ليستأنف من أي جهاز أو متصفح،
	// لأن تخزين المتصفح مربوط بالعنوان ولا ينجو من تغيّره.
	Code       string   `json:"code"`
	Category   Category `json:"category"`
	Name       string   `json:"name"`
	StartedAt  string   `json:"started_at"`
	FinishedAt string   `json:"finished_at"`
}

// Finished يحدد ما إذا كان المشارك أنهى الاستبيان.
func (s Session) Finished() bool { return s.FinishedAt != "" }

// Answer إجابة واحدة على سؤال واحد.
type Answer struct {
	SessionID  string          `json:"session_id"`
	QuestionID int64           `json:"question_id"`
	Value      json.RawMessage `json:"value"`
	AnsweredAt string          `json:"answered_at"`
}
