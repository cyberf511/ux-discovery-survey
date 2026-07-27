// Package export يحوّل الإجابات المجموعة إلى CSV صالح للفتح في إكسل.
package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"surveyapp/internal/model"
)

// bom علامة ترتيب البايت لـ UTF-8. بدونها يعرض إكسل العربي كرموز مكسورة.
const bom = "\xEF\xBB\xBF"

// Data كل ما يلزم لبناء ملف التصدير.
type Data struct {
	Questions []model.Question
	Sessions  []model.Session
	Answers   map[string]map[int64]json.RawMessage
	// AnsweredAt أوقات الإجابات، تُستخدم في التصدير الطويل وحده.
	AnsweredAt map[string]map[int64]string
}

// WriteLongCSV يكتب صفًا لكل إجابة بدل عمود لكل سؤال.
//
// الشكل العريض (صف لكل مشارك) يصبح عديم الفائدة مع مئات الأسئلة: معظم
// خلاياه فارغة لأن كل دور يرى جزءًا من الكتالوج، وصفوفه أطول من أن تُقرأ.
// الشكل الطويل يُفهرس ويُصفّى ويُجمَّع مباشرة، بشرًا كان القارئ أو نموذجًا.
func WriteLongCSV(w io.Writer, d Data) error {
	if _, err := io.WriteString(w, bom); err != nil {
		return err
	}
	cw := csv.NewWriter(w)

	header := []string{
		"معرّف المشارك", "الدور", "الاسم", "أنهى الاستبيان؟",
		"القسم", "رقم السؤال", "السؤال", "نوع السؤال", "الإجابة", "وقت الإجابة",
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, s := range d.Sessions {
		answers := d.Answers[s.ID]
		if len(answers) == 0 {
			continue
		}
		for _, q := range d.Questions {
			raw, ok := answers[q.ID]
			if !ok {
				continue
			}
			text := FormatValue(q, raw)
			if text == "" {
				continue
			}
			row := []string{
				s.ID,
				model.CategoryLabel(s.Category),
				s.Name,
				yesNo(s.Finished()),
				q.Section,
				fmt.Sprintf("%d", q.ID),
				q.Text,
				KindLabel(q.Kind),
				text,
				d.AnsweredAt[s.ID][q.ID],
			}
			if err := cw.Write(row); err != nil {
				return err
			}
		}
	}

	cw.Flush()
	return cw.Error()
}

// KindLabel الاسم العربي لنوع السؤال.
func KindLabel(k model.Kind) string {
	switch k {
	case model.KindLongText:
		return "نص طويل"
	case model.KindShortText:
		return "نص قصير"
	case model.KindBoolean:
		return "صح/خطأ"
	case model.KindSingleChoice:
		return "اختيار واحد"
	case model.KindMultiChoice:
		return "اختيار متعدد"
	case model.KindRanking:
		return "ترتيب"
	case model.KindScale:
		return "مقياس ١–٥"
	}
	return string(k)
}

// WriteCSV يكتب ملف CSV بترميز UTF-8 مع BOM: صف لكل مشارك وعمود لكل سؤال.
func WriteCSV(w io.Writer, d Data) error {
	if _, err := io.WriteString(w, bom); err != nil {
		return err
	}
	cw := csv.NewWriter(w)

	header := []string{"المعرّف", "الفئة", "الاسم", "وقت البداية", "وقت الانتهاء", "مكتملة؟", "عدد الإجابات"}
	for _, q := range d.Questions {
		header = append(header, columnTitle(q))
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, s := range d.Sessions {
		answers := d.Answers[s.ID]
		row := []string{
			s.ID,
			model.CategoryLabel(s.Category),
			s.Name,
			s.StartedAt,
			s.FinishedAt,
			yesNo(s.Finished()),
			fmt.Sprintf("%d", len(answers)),
		}
		for _, q := range d.Questions {
			row = append(row, FormatValue(q, answers[q.ID]))
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}

// columnTitle عنوان عمود السؤال: القسم ثم نص السؤال.
func columnTitle(q model.Question) string {
	if q.Section == "" {
		return q.Text
	}
	return q.Section + " — " + q.Text
}

// FormatValue يحوّل قيمة الإجابة المخزَّنة إلى نص مقروء حسب نوع السؤال.
func FormatValue(q model.Question, raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	switch q.Kind {
	case model.KindBoolean:
		var b bool
		if err := json.Unmarshal(raw, &b); err == nil {
			return yesNo(b)
		}
	case model.KindScale:
		var n json.Number
		if err := json.Unmarshal(raw, &n); err == nil {
			return n.String()
		}
	case model.KindRanking:
		var items []string
		if err := json.Unmarshal(raw, &items); err == nil {
			parts := make([]string, 0, len(items))
			for i, it := range items {
				parts = append(parts, fmt.Sprintf("%d. %s", i+1, it))
			}
			return strings.Join(parts, " | ")
		}
	case model.KindMultiChoice:
		var items []string
		if err := json.Unmarshal(raw, &items); err == nil {
			return strings.Join(items, " | ")
		}
	}
	// النصوص والاختيار الواحد، ومَخرج احتياطي لأي قيمة لم تطابق نوعها المتوقع.
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

func yesNo(b bool) string {
	if b {
		return "نعم"
	}
	return "لا"
}
