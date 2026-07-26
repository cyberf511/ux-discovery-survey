// أداة تُخرج كتالوج الأسئلة إلى ملف Markdown للمراجعة البشرية.
//
//	go run ./cmd/catalog > docs/QUESTIONS.md
package main

import (
	"fmt"
	"os"
	"strings"

	"surveyapp/internal/model"
	"surveyapp/internal/seed"
)

var kindLabels = map[model.Kind]string{
	model.KindLongText:     "نص طويل",
	model.KindShortText:    "نص قصير",
	model.KindBoolean:      "صح/خطأ",
	model.KindSingleChoice: "اختيار واحد",
	model.KindMultiChoice:  "اختيار متعدد",
	model.KindRanking:      "ترتيب",
	model.KindScale:        "مقياس ١–٥",
}

func main() {
	qs := seed.Questions()
	out := &strings.Builder{}

	writeHeader(out, qs)
	writeSections(out, qs)
	writeOverlap(out, qs)

	if _, err := os.Stdout.WriteString(out.String()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func writeHeader(out *strings.Builder, qs []model.Question) {
	fmt.Fprintln(out, "# كتالوج أسئلة استبيان UX Discovery")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "> مولّد من `internal/seed/catalog.go` عبر `go run ./cmd/catalog`.")
	fmt.Fprintln(out, "> عدّل الشيفرة لا هذا الملف، ثم أعد التوليد.")
	fmt.Fprintln(out)

	perCat := map[model.Category]int{}
	sections := []string{}
	seen := map[string]bool{}
	for _, q := range qs {
		if !seen[q.Section] {
			seen[q.Section] = true
			sections = append(sections, q.Section)
		}
		for _, c := range model.AllCategories {
			if q.AppliesTo(c) {
				perCat[c]++
			}
		}
	}

	fmt.Fprintf(out, "**الإجمالي:** %d سؤالًا في %d قسمًا.\n\n", len(qs), len(sections))
	fmt.Fprintln(out, "| الفئة | عدد الأسئلة التي تراها |")
	fmt.Fprintln(out, "|---|---|")
	for _, c := range model.AllCategories {
		fmt.Fprintf(out, "| %s | %d |\n", model.CategoryLabel(c), perCat[c])
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "**رموز الفئات في الجداول:** ح = حارس · مش = مشرف ميداني · عام = مشرف عام · مدير = مدير شركة")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "علامة `●` تعني أن الفئة ترى السؤال. عمود **إلزامي** يعني أن المشارك لا يستطيع تخطّيه.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "---")
	fmt.Fprintln(out)
}

func writeSections(out *strings.Builder, qs []model.Question) {
	order := []string{}
	bySection := map[string][]model.Question{}
	for _, q := range qs {
		if _, ok := bySection[q.Section]; !ok {
			order = append(order, q.Section)
		}
		bySection[q.Section] = append(bySection[q.Section], q)
	}

	for _, sec := range order {
		items := bySection[sec]
		fmt.Fprintf(out, "## %s\n\n", sec)
		fmt.Fprintf(out, "%d سؤالًا.\n\n", len(items))
		fmt.Fprintln(out, "| # | السؤال | النوع | ح | مش | عام | مدير | إلزامي | الخيارات |")
		fmt.Fprintln(out, "|---|---|---|:-:|:-:|:-:|:-:|:-:|---|")
		for i, q := range items {
			fmt.Fprintf(out, "| %d | %s | %s | %s | %s | %s | %s | %s | %s |\n",
				i+1,
				escape(q.Text),
				kindLabels[q.Kind],
				mark(q, model.CatGuard),
				mark(q, model.CatSupervisor),
				mark(q, model.CatAreaManager),
				mark(q, model.CatCompanyManager),
				yesNo(q.Required),
				escape(strings.Join(q.Options, " · ")),
			)
		}
		fmt.Fprintln(out)
	}
	fmt.Fprintln(out, "---")
	fmt.Fprintln(out)
}

// writeOverlap يبرز الأسئلة المعروضة لأكثر من فئة — مادة التنقيح الأولى،
// لأن السؤال الموجّه لمنفّذ العمل لا يناسب من يشرف عليه.
func writeOverlap(out *strings.Builder, qs []model.Question) {
	fmt.Fprintln(out, "## مراجعة التداخل بين الفئات")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "الأسئلة التالية تظهر لأكثر من فئة. راجعها وقرّر أيها يجب أن يخص فئة واحدة:")
	fmt.Fprintln(out)

	counts := map[int][]model.Question{}
	for _, q := range qs {
		n := 0
		for _, c := range model.AllCategories {
			if q.AppliesTo(c) {
				n++
			}
		}
		counts[n] = append(counts[n], q)
	}

	fmt.Fprintln(out, "| عدد الفئات | عدد الأسئلة |")
	fmt.Fprintln(out, "|---|---|")
	for n := 1; n <= len(model.AllCategories); n++ {
		fmt.Fprintf(out, "| %d | %d |\n", n, len(counts[n]))
	}
	fmt.Fprintln(out)

	for n := len(model.AllCategories); n >= 2; n-- {
		if len(counts[n]) == 0 {
			continue
		}
		fmt.Fprintf(out, "### تظهر لـ%d فئات (%d سؤالًا)\n\n", n, len(counts[n]))
		fmt.Fprintln(out, "| القسم | السؤال | ح | مش | عام | مدير |")
		fmt.Fprintln(out, "|---|---|:-:|:-:|:-:|:-:|")
		for _, q := range counts[n] {
			fmt.Fprintf(out, "| %s | %s | %s | %s | %s | %s |\n",
				q.Section, escape(q.Text),
				mark(q, model.CatGuard),
				mark(q, model.CatSupervisor),
				mark(q, model.CatAreaManager),
				mark(q, model.CatCompanyManager))
		}
		fmt.Fprintln(out)
	}
}

func mark(q model.Question, c model.Category) string {
	if q.AppliesTo(c) {
		return "●"
	}
	return "·"
}

func yesNo(b bool) string {
	if b {
		return "نعم"
	}
	return "—"
}

// escape يمنع كسر أعمدة الجدول بعلامة الأنبوب داخل النص.
func escape(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}
