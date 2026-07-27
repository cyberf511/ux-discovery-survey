package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"surveyapp/internal/model"
)

func TestWriteCSVStartsWithBOM(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, Data{}); err != nil {
		t.Fatalf("كتابة CSV: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte(bom)) {
		t.Fatalf("الملف يجب أن يبدأ بـ BOM وإلا كسر إكسل العربي")
	}
}

func TestWriteCSVRowsAndColumns(t *testing.T) {
	q1 := model.Question{ID: 1, Text: "صف يومك", Kind: model.KindLongText, Section: "يوم العمل"}
	q2 := model.Question{ID: 2, Text: "رتّب الميزات", Kind: model.KindRanking}

	d := Data{
		Questions: []model.Question{q1, q2},
		Sessions: []model.Session{
			{ID: "s1", Category: model.CatGuard, Name: "فيصل", StartedAt: "2026-07-26T08:00:00Z", FinishedAt: "2026-07-26T08:20:00Z"},
			{ID: "s2", Category: model.CatSupervisor, Name: "", StartedAt: "2026-07-26T09:00:00Z"},
		},
		Answers: map[string]map[int64]json.RawMessage{
			"s1": {
				1: json.RawMessage(`"أبدأ بالتسليم"`),
				2: json.RawMessage(`["الحضور","البلاغات"]`),
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteCSV(&buf, d); err != nil {
		t.Fatalf("كتابة CSV: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "يوم العمل — صف يومك") {
		t.Fatalf("عنوان العمود يجب أن يجمع القسم ونص السؤال:\n%s", out)
	}
	if !strings.Contains(out, "1. الحضور | 2. البلاغات") {
		t.Fatalf("سؤال الترتيب يجب أن يُصدَّر مرقّمًا:\n%s", out)
	}
	if !strings.Contains(out, "حارس") || !strings.Contains(out, "مشرف ميداني") {
		t.Fatalf("أسماء الفئات العربية مفقودة:\n%s", out)
	}
	// الجلسة غير المكتملة تبقى في الملف — نقطة الانسحاب معلومة مفيدة.
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("توقعنا ترويسة وصفّين، وجدنا %d سطرًا:\n%s", len(lines), out)
	}
	if !strings.Contains(lines[2], "لا") {
		t.Fatalf("الجلسة غير المكتملة يجب أن تُعلَّم بـ «لا»:\n%s", lines[2])
	}
}

// الشكل الطويل هو المخرج الصالح للتحليل: صف لكل إجابة، بلا خلايا فارغة.
func TestWriteLongCSVEmitsOneRowPerAnswer(t *testing.T) {
	q1 := model.Question{ID: 1, Text: "صف يومك", Kind: model.KindLongText, Section: "يوم العمل"}
	q2 := model.Question{ID: 2, Text: "رتّب الميزات", Kind: model.KindRanking}
	q3 := model.Question{ID: 3, Text: "سؤال بلا إجابة", Kind: model.KindLongText}

	d := Data{
		Questions: []model.Question{q1, q2, q3},
		Sessions: []model.Session{
			{ID: "s1", Category: model.CatGuard, Name: "فيصل", FinishedAt: "2026-07-27T08:20:00Z"},
			{ID: "s2", Category: model.CatDashboard, Name: "فهد"},
			{ID: "s3", Category: model.CatSupervisor, Name: "بلا إجابات"},
		},
		Answers: map[string]map[int64]json.RawMessage{
			"s1": {
				1: json.RawMessage(`"أبدأ بالتسليم"`),
				2: json.RawMessage(`["الحضور","البلاغات"]`),
			},
			"s2": {1: json.RawMessage(`"أراجع المؤشرات"`)},
		},
		AnsweredAt: map[string]map[int64]string{
			"s1": {1: "2026-07-27T08:01:00Z", 2: "2026-07-27T08:05:00Z"},
			"s2": {1: "2026-07-27T09:00:00Z"},
		},
	}

	var buf bytes.Buffer
	if err := WriteLongCSV(&buf, d); err != nil {
		t.Fatalf("كتابة CSV الطويل: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte(bom)) {
		t.Fatalf("الملف يجب أن يبدأ بـ BOM")
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	// ترويسة + ٣ إجابات فقط: لا صفوف للأسئلة بلا إجابة ولا للجلسات الفارغة.
	if len(lines) != 4 {
		t.Fatalf("توقعنا ترويسة و٣ صفوف، وجدنا %d:\n%s", len(lines), buf.String())
	}
	out := buf.String()
	if strings.Contains(out, "سؤال بلا إجابة") {
		t.Fatalf("السؤال بلا إجابة يجب ألا يظهر")
	}
	if strings.Contains(out, "بلا إجابات") {
		t.Fatalf("الجلسة بلا إجابات يجب ألا تظهر")
	}
	if !strings.Contains(out, "1. الحضور | 2. البلاغات") {
		t.Fatalf("سؤال الترتيب يجب أن يُصدَّر مرقّمًا:\n%s", out)
	}
	if !strings.Contains(out, "داشبورد") {
		t.Fatalf("اسم الدور العربي مفقود:\n%s", out)
	}
	if !strings.Contains(out, "2026-07-27T08:01:00Z") {
		t.Fatalf("وقت الإجابة مفقود:\n%s", out)
	}
	if !strings.Contains(lines[0], "نوع السؤال") {
		t.Fatalf("الترويسة ناقصة: %s", lines[0])
	}
}

func TestFormatValue(t *testing.T) {
	cases := []struct {
		name string
		kind model.Kind
		raw  string
		want string
	}{
		{"نص", model.KindLongText, `"إجابة طويلة"`, "إجابة طويلة"},
		{"صح", model.KindBoolean, `true`, "نعم"},
		{"خطأ", model.KindBoolean, `false`, "لا"},
		{"مقياس", model.KindScale, `4`, "4"},
		{"اختيار متعدد", model.KindMultiChoice, `["أ","ب"]`, "أ | ب"},
		{"ترتيب", model.KindRanking, `["أ","ب"]`, "1. أ | 2. ب"},
		{"فارغ", model.KindLongText, ``, ""},
		{"عدم", model.KindLongText, `null`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := FormatValue(model.Question{Kind: c.kind}, json.RawMessage(c.raw))
			if got != c.want {
				t.Fatalf("توقعنا %q، وجدنا %q", c.want, got)
			}
		})
	}
}

// قيمة مخزَّنة بنوع مخالف لنوع السؤال يجب ألا تسبب انهيارًا.
func TestFormatValueMismatchedType(t *testing.T) {
	got := FormatValue(model.Question{Kind: model.KindRanking}, json.RawMessage(`"نص وليس قائمة"`))
	if got != "نص وليس قائمة" {
		t.Fatalf("المخرج الاحتياطي لم يعمل: %q", got)
	}
}
