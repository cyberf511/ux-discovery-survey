package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"surveyapp/internal/model"
	"surveyapp/internal/store"
)

const testPassword = "سر"

func newTestServer(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("فتح قاعدة الاختبار: %v", err)
	}
	web := fstest.MapFS{
		"index.html": {Data: []byte("<html>مشارك</html>")},
		"admin.html": {Data: []byte("<html>إدارة</html>")},
	}
	srv := httptest.NewServer(New(st, web, dbPath, testPassword).Routes())
	t.Cleanup(func() { srv.Close(); st.Close() })
	return srv, st
}

func do(t *testing.T, c *http.Client, method, url string, body any) (*http.Response, map[string]any) {
	t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatalf("بناء الطلب: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.Do(req)
	if err != nil {
		t.Fatalf("تنفيذ الطلب: %v", err)
	}
	var out map[string]any
	if strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
		json.NewDecoder(res.Body).Decode(&out)
	}
	res.Body.Close()
	return res, out
}

func TestParticipantFlow(t *testing.T) {
	srv, st := newTestServer(t)
	qid, err := st.CreateQuestion(model.Question{Text: "صف يومك", Kind: model.KindLongText})
	if err != nil {
		t.Fatalf("إضافة سؤال: %v", err)
	}
	c := srv.Client()

	res, meta := do(t, c, "GET", srv.URL+"/api/meta", nil)
	if res.StatusCode != 200 || meta["open"] != true {
		t.Fatalf("الميتا خاطئة: %d %v", res.StatusCode, meta)
	}

	res, sess := do(t, c, "POST", srv.URL+"/api/sessions",
		map[string]string{"category": "guard", "name": "فيصل"})
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("إنشاء الجلسة فشل: %d %v", res.StatusCode, sess)
	}
	id, _ := sess["id"].(string)
	if id == "" {
		t.Fatalf("معرّف الجلسة مفقود: %v", sess)
	}

	res, _ = do(t, c, "POST", srv.URL+"/api/sessions/"+id+"/answers",
		map[string]any{"question_id": qid, "value": "أبدأ بالتسليم"})
	if res.StatusCode != 200 {
		t.Fatalf("حفظ الإجابة فشل: %d", res.StatusCode)
	}

	res, got := do(t, c, "GET", srv.URL+"/api/sessions/"+id, nil)
	if res.StatusCode != 200 {
		t.Fatalf("قراءة الجلسة فشلت: %d", res.StatusCode)
	}
	answers, _ := got["answers"].(map[string]any)
	if len(answers) != 1 {
		t.Fatalf("توقعنا إجابة واحدة، وجدنا %v", answers)
	}

	res, _ = do(t, c, "POST", srv.URL+"/api/sessions/"+id+"/finish", map[string]any{})
	if res.StatusCode != 200 {
		t.Fatalf("إنهاء الجلسة فشل: %d", res.StatusCode)
	}
}

// الاستئناف بالكود هو المسار الوحيد المستقل عن تخزين المتصفح،
// لأن localStorage مربوط بالعنوان ويضيع عند تغيّر المضيف أو مسح بيانات الموقع.
func TestResumeByCode(t *testing.T) {
	srv, _ := newTestServer(t)
	c := srv.Client()

	_, created := do(t, c, "POST", srv.URL+"/api/sessions", map[string]string{"category": "guard"})
	code, _ := created["code"].(string)
	if code == "" {
		t.Fatalf("إنشاء الجلسة يجب أن يعيد كود متابعة: %v", created)
	}

	// عميل جديد بلا كوكيز ولا تخزين — يحاكي جهازًا أو متصفحًا آخر.
	fresh := &http.Client{}
	res, resumed := do(t, fresh, "POST", srv.URL+"/api/sessions/resume",
		map[string]string{"code": strings.ToLower(code)})
	if res.StatusCode != 200 {
		t.Fatalf("الاستئناف بالكود فشل: %d %v", res.StatusCode, resumed)
	}
	if resumed["id"] != created["id"] {
		t.Fatalf("الكود أعاد جلسة خاطئة: %v", resumed)
	}

	res, _ = do(t, fresh, "POST", srv.URL+"/api/sessions/resume", map[string]string{"code": "ZZZZZZ"})
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("كود مجهول يجب أن يعيد 404، وجدنا %d", res.StatusCode)
	}
}

// كوكي الجلسة يعيد المشارك تلقائيًا حتى لو مُسح تخزين المتصفح.
func TestCurrentSessionFromCookie(t *testing.T) {
	srv, _ := newTestServer(t)
	c := &http.Client{Jar: newJar()}

	_, created := do(t, c, "POST", srv.URL+"/api/sessions", map[string]string{"category": "guard"})

	res, current := do(t, c, "GET", srv.URL+"/api/sessions/current", nil)
	if res.StatusCode != 200 {
		t.Fatalf("استرجاع الجلسة من الكوكي فشل: %d", res.StatusCode)
	}
	if current["id"] != created["id"] {
		t.Fatalf("الكوكي أعاد جلسة خاطئة: %v", current)
	}

	// عميل بلا كوكي يجب أن يحصل على 404 لا على جلسة شخص آخر.
	res, _ = do(t, &http.Client{}, "GET", srv.URL+"/api/sessions/current", nil)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("بلا كوكي توقعنا 404، وجدنا %d", res.StatusCode)
	}
}

// «ابدأ من جديد» ينسى هوية الجهاز فقط: الإجابات السابقة تبقى محفوظة
// حتى يقدر الشخص نفسه يجيب بصفة ثانية بلا أن يفقد ما أجابه أولًا.
func TestLeaveSessionKeepsAnswers(t *testing.T) {
	srv, st := newTestServer(t)
	qid, _ := st.CreateQuestion(model.Question{Text: "سؤال", Kind: model.KindLongText})
	c := &http.Client{Jar: newJar()}

	_, created := do(t, c, "POST", srv.URL+"/api/sessions", map[string]string{"category": "guard"})
	id, _ := created["id"].(string)
	code, _ := created["code"].(string)
	do(t, c, "POST", srv.URL+"/api/sessions/"+id+"/answers",
		map[string]any{"question_id": qid, "value": "إجابة الحارس"})

	res, _ := do(t, c, "POST", srv.URL+"/api/sessions/leave", map[string]any{})
	if res.StatusCode != 200 {
		t.Fatalf("مغادرة الجلسة فشلت: %d", res.StatusCode)
	}

	// الكوكي ما عاد يعيد الجلسة، فالمشارك يرى شاشة اختيار الفئة.
	res, _ = do(t, c, "GET", srv.URL+"/api/sessions/current", nil)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("بعد المغادرة توقعنا 404، وجدنا %d", res.StatusCode)
	}

	// لكن الجلسة وإجاباتها باقية، ويمكن استرجاعها بالكود.
	answers, err := st.SessionAnswers(id)
	if err != nil || len(answers) != 1 {
		t.Fatalf("الإجابات السابقة يجب أن تبقى: %v %v", answers, err)
	}
	res, resumed := do(t, c, "POST", srv.URL+"/api/sessions/resume", map[string]string{"code": code})
	if res.StatusCode != 200 || resumed["id"] != id {
		t.Fatalf("الاستئناف بالكود بعد المغادرة فشل: %d %v", res.StatusCode, resumed)
	}
}

// تعطيل قسم يخفي أسئلته عن المشاركين ويُبقي الإجابات المجموعة عليها.
func TestToggleSectionHidesQuestionsButKeepsAnswers(t *testing.T) {
	srv, st := newTestServer(t)
	qid, _ := st.CreateQuestion(model.Question{
		Text: "سؤال مالي", Kind: model.KindLongText, Section: "المالية والفوترة والعقود",
	})
	st.CreateQuestion(model.Question{Text: "سؤال ميداني", Kind: model.KindLongText, Section: "الدوريات"})
	sess, _ := st.CreateSession(model.CatGuard, "فيصل")
	st.SaveAnswer(sess.ID, qid, json.RawMessage(`"إجابة سابقة"`))

	admin := loginClient(t, srv)
	res, out := do(t, admin, "POST", srv.URL+"/api/admin/sections/toggle",
		map[string]any{"section": "المالية والفوترة والعقود", "active": false})
	if res.StatusCode != 200 {
		t.Fatalf("تعطيل القسم فشل: %d %v", res.StatusCode, out)
	}
	if out["affected"] != float64(1) {
		t.Fatalf("توقعنا تعطيل سؤال واحد، تأثّر %v", out["affected"])
	}

	// المشارك ما عاد يرى السؤال المعطّل.
	_, view := do(t, srv.Client(), "GET", srv.URL+"/api/sessions/"+sess.ID, nil)
	questions, _ := view["questions"].([]any)
	for _, q := range questions {
		if m, ok := q.(map[string]any); ok && m["section"] == "المالية والفوترة والعقود" {
			t.Fatalf("السؤال المعطّل ما زال يظهر للمشارك")
		}
	}
	if len(questions) != 1 {
		t.Fatalf("توقعنا سؤالًا واحدًا فعّالًا، وجدنا %d", len(questions))
	}

	// والإجابة السابقة باقية وتظهر في التصدير.
	answers, _ := st.SessionAnswers(sess.ID)
	if len(answers) != 1 {
		t.Fatalf("الإجابة على السؤال المعطّل ضاعت")
	}
	csv, err := admin.Get(srv.URL + "/api/admin/export.csv")
	if err != nil {
		t.Fatalf("التصدير: %v", err)
	}
	defer csv.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(csv.Body)
	if !strings.Contains(buf.String(), "إجابة سابقة") {
		t.Fatalf("إجابة القسم المعطّل يجب أن تبقى في التصدير")
	}

	// وإعادة التشغيل تعيد الأسئلة كما كانت.
	do(t, admin, "POST", srv.URL+"/api/admin/sections/toggle",
		map[string]any{"section": "المالية والفوترة والعقود", "active": true})
	_, view = do(t, srv.Client(), "GET", srv.URL+"/api/sessions/"+sess.ID, nil)
	if q, _ := view["questions"].([]any); len(q) != 2 {
		t.Fatalf("إعادة تشغيل القسم لم تُعد أسئلته: %d", len(q))
	}
}

// الدور المعطّل يختفي من شاشة اختيار الفئة ولا تُقبل جلسات جديدة له.
func TestDisabledCategoryIsHiddenAndRejected(t *testing.T) {
	srv, _ := newTestServer(t)
	admin := loginClient(t, srv)

	res, out := do(t, admin, "POST", srv.URL+"/api/admin/categories/toggle",
		map[string]any{"category": "company_manager", "enabled": false})
	if res.StatusCode != 200 {
		t.Fatalf("تعطيل الدور فشل: %d %v", res.StatusCode, out)
	}

	_, meta := do(t, srv.Client(), "GET", srv.URL+"/api/meta", nil)
	cats, _ := meta["categories"].([]any)
	if len(cats) != len(model.AllCategories)-1 {
		t.Fatalf("توقعنا %d أدوار للمشارك، وجدنا %d", len(model.AllCategories)-1, len(cats))
	}
	for _, c := range cats {
		if m, ok := c.(map[string]any); ok && m["value"] == "company_manager" {
			t.Fatalf("الدور المعطّل ما زال يظهر للمشارك")
		}
	}

	res, _ = do(t, srv.Client(), "POST", srv.URL+"/api/sessions",
		map[string]string{"category": "company_manager"})
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("توقعنا 409 لدور معطّل، وجدنا %d", res.StatusCode)
	}

	// اللوحة ترى كل الأدوار مع حالتها لتستطيع إعادة التشغيل.
	_, state := do(t, admin, "GET", srv.URL+"/api/admin/state", nil)
	adminCats, _ := state["categories"].([]any)
	if len(adminCats) != len(model.AllCategories) {
		t.Fatalf("اللوحة يجب أن ترى كل الأدوار: %d", len(adminCats))
	}

	do(t, admin, "POST", srv.URL+"/api/admin/categories/toggle",
		map[string]any{"category": "company_manager", "enabled": true})
	res, _ = do(t, srv.Client(), "POST", srv.URL+"/api/sessions",
		map[string]string{"category": "company_manager"})
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("إعادة تشغيل الدور لم تعمل: %d", res.StatusCode)
	}
}

func TestUnknownCategoryRejected(t *testing.T) {
	srv, _ := newTestServer(t)
	res, _ := do(t, srv.Client(), "POST", srv.URL+"/api/sessions",
		map[string]string{"category": "ملك"})
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("توقعنا 400 لفئة مجهولة، وجدنا %d", res.StatusCode)
	}
}

func TestAdminRequiresLogin(t *testing.T) {
	srv, _ := newTestServer(t)
	res, _ := do(t, srv.Client(), "GET", srv.URL+"/api/admin/state", nil)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("توقعنا 401 بدون تسجيل دخول، وجدنا %d", res.StatusCode)
	}

	res, _ = do(t, srv.Client(), "POST", srv.URL+"/api/admin/login",
		map[string]string{"password": "خطأ"})
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("توقعنا 401 لكلمة مرور خاطئة، وجدنا %d", res.StatusCode)
	}
}

func loginClient(t *testing.T, srv *httptest.Server) *http.Client {
	t.Helper()
	jar := newJar()
	c := &http.Client{Jar: jar}
	res, _ := do(t, c, "POST", srv.URL+"/api/admin/login", map[string]string{"password": testPassword})
	if res.StatusCode != 200 {
		t.Fatalf("تسجيل الدخول فشل: %d", res.StatusCode)
	}
	return c
}

func TestAdminCanLockSurveyAndBlockAnswers(t *testing.T) {
	srv, st := newTestServer(t)
	qid, _ := st.CreateQuestion(model.Question{Text: "سؤال", Kind: model.KindLongText})
	sess, _ := st.CreateSession(model.CatGuard, "")

	admin := loginClient(t, srv)
	res, _ := do(t, admin, "POST", srv.URL+"/api/admin/open", map[string]bool{"open": false})
	if res.StatusCode != 200 {
		t.Fatalf("قفل الاستبيان فشل: %d", res.StatusCode)
	}

	res, _ = do(t, srv.Client(), "POST", srv.URL+"/api/sessions/"+sess.ID+"/answers",
		map[string]any{"question_id": qid, "value": "متأخر"})
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("توقعنا 409 بعد القفل، وجدنا %d", res.StatusCode)
	}
}

func TestExportCSVHasBOMAndAnswers(t *testing.T) {
	srv, st := newTestServer(t)
	qid, _ := st.CreateQuestion(model.Question{Text: "صف يومك", Kind: model.KindLongText})
	sess, _ := st.CreateSession(model.CatGuard, "فيصل")
	st.SaveAnswer(sess.ID, qid, json.RawMessage(`"أبدأ بالتسليم"`))

	admin := loginClient(t, srv)
	res, err := admin.Get(srv.URL + "/api/admin/export.csv")
	if err != nil {
		t.Fatalf("طلب التصدير: %v", err)
	}
	defer res.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)

	if !bytes.HasPrefix(buf.Bytes(), []byte("\xEF\xBB\xBF")) {
		t.Fatalf("ملف CSV يجب أن يبدأ بـ BOM")
	}
	if !strings.Contains(buf.String(), "أبدأ بالتسليم") || !strings.Contains(buf.String(), "فيصل") {
		t.Fatalf("محتوى التصدير ناقص:\n%s", buf.String())
	}
}

func TestBackupDownloadsDatabase(t *testing.T) {
	srv, _ := newTestServer(t)
	admin := loginClient(t, srv)
	res, err := admin.Get(srv.URL + "/api/admin/backup.db")
	if err != nil {
		t.Fatalf("طلب النسخة الاحتياطية: %v", err)
	}
	defer res.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	if buf.Len() == 0 {
		t.Fatalf("النسخة الاحتياطية فارغة")
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("SQLite format 3")) {
		t.Fatalf("الملف المنزَّل ليس قاعدة SQLite")
	}
}

func TestImportQuestionsParsesSections(t *testing.T) {
	srv, st := newTestServer(t)
	admin := loginClient(t, srv)

	res, out := do(t, admin, "POST", srv.URL+"/api/admin/questions/import", map[string]any{
		"text":       "## الحضور\n- كيف تسجل حضورك؟\n\n* ما المشاكل التي تواجهها؟",
		"categories": []string{"guard"},
	})
	if res.StatusCode != 200 {
		t.Fatalf("الاستيراد فشل: %d %v", res.StatusCode, out)
	}
	if out["added"] != float64(2) {
		t.Fatalf("توقعنا سؤالين مضافين، وجدنا %v", out["added"])
	}
	qs, _ := st.Questions(false)
	if len(qs) != 2 || qs[0].Section != "الحضور" {
		t.Fatalf("الأقسام لم تُفسَّر: %+v", qs)
	}
}

func TestServesIndexPage(t *testing.T) {
	srv, _ := newTestServer(t)
	res, err := srv.Client().Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("طلب الصفحة: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("توقعنا 200، وجدنا %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("نوع المحتوى خاطئ: %s", ct)
	}
}

// newJar يوفّر مخزن كوكيز لاختبارات الإدارة حتى تُحفظ جلسة الدخول.
func newJar() http.CookieJar {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return jar
}
