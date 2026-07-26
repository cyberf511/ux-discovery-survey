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
