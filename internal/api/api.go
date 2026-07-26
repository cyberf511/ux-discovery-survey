// Package api يعرّف معالجات HTTP للمشارك وللوحة الإدارة.
package api

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"surveyapp/internal/model"
	"surveyapp/internal/store"
)

// Server يجمع الاعتماديات المشتركة بين المعالجات.
type Server struct {
	st       *store.Store
	web      fs.FS
	dbPath   string
	authTok  string // قيمة كوكي الأدمن المتوقعة
	password string
}

// New ينشئ الخادم. web هو نظام ملفات الواجهة الثابتة.
func New(st *store.Store, web fs.FS, dbPath, adminPassword string) *Server {
	return &Server{
		st:       st,
		web:      web,
		dbPath:   dbPath,
		authTok:  tokenFor(adminPassword),
		password: adminPassword,
	}
}

// Routes يبني موجّه الطلبات.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// الواجهة الثابتة
	files := http.FileServer(http.FS(s.web))
	mux.Handle("GET /static/", files)
	mux.HandleFunc("GET /", s.page("index.html"))
	mux.HandleFunc("GET /admin", s.page("admin.html"))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// واجهة المشارك
	mux.HandleFunc("GET /api/meta", s.handleMeta)
	mux.HandleFunc("POST /api/sessions", s.handleCreateSession)
	mux.HandleFunc("POST /api/sessions/resume", s.handleResumeSession)
	mux.HandleFunc("POST /api/sessions/leave", s.handleLeaveSession)
	mux.HandleFunc("GET /api/sessions/current", s.handleCurrentSession)
	mux.HandleFunc("GET /api/sessions/{id}", s.handleGetSession)
	mux.HandleFunc("POST /api/sessions/{id}/answers", s.handleSaveAnswer)
	mux.HandleFunc("POST /api/sessions/{id}/finish", s.handleFinishSession)

	// الإدارة
	mux.HandleFunc("POST /api/admin/login", s.handleLogin)
	mux.HandleFunc("POST /api/admin/logout", s.handleLogout)
	mux.Handle("GET /api/admin/state", s.auth(http.HandlerFunc(s.handleAdminState)))
	mux.Handle("POST /api/admin/questions", s.auth(http.HandlerFunc(s.handleCreateQuestion)))
	mux.Handle("PUT /api/admin/questions/{id}", s.auth(http.HandlerFunc(s.handleUpdateQuestion)))
	mux.Handle("DELETE /api/admin/questions/{id}", s.auth(http.HandlerFunc(s.handleDeleteQuestion)))
	mux.Handle("POST /api/admin/questions/reorder", s.auth(http.HandlerFunc(s.handleReorder)))
	mux.Handle("POST /api/admin/sections/toggle", s.auth(http.HandlerFunc(s.handleToggleSection)))
	mux.Handle("POST /api/admin/categories/toggle", s.auth(http.HandlerFunc(s.handleToggleCategory)))
	mux.Handle("POST /api/admin/questions/import", s.auth(http.HandlerFunc(s.handleImport)))
	mux.Handle("POST /api/admin/open", s.auth(http.HandlerFunc(s.handleSetOpen)))
	mux.Handle("GET /api/admin/results", s.auth(http.HandlerFunc(s.handleResults)))
	mux.Handle("DELETE /api/admin/sessions/{id}", s.auth(http.HandlerFunc(s.handleDeleteSession)))
	mux.Handle("GET /api/admin/export.csv", s.auth(http.HandlerFunc(s.handleExportCSV)))
	mux.Handle("GET /api/admin/backup.db", s.auth(http.HandlerFunc(s.handleBackup)))

	return logging(mux)
}

func (s *Server) page(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/admin" {
			http.NotFound(w, r)
			return
		}
		b, err := fs.ReadFile(s.web, name)
		if err != nil {
			http.Error(w, "الصفحة غير موجودة", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(b)
	}
}

// ---------- واجهة المشارك ----------

type categoryInfo struct {
	Value   model.Category `json:"value"`
	Label   string         `json:"label"`
	Enabled bool           `json:"enabled"`
}

// categories يبني قائمة الأدوار. includeDisabled للوحة الإدارة فقط؛
// المشارك لا يرى الأدوار المعطّلة أصلًا.
func (s *Server) categories(includeDisabled bool) ([]categoryInfo, error) {
	disabled, err := s.st.DisabledCategories()
	if err != nil {
		return nil, err
	}
	out := make([]categoryInfo, 0, len(model.AllCategories))
	for _, c := range model.AllCategories {
		if disabled[c] && !includeDisabled {
			continue
		}
		out = append(out, categoryInfo{
			Value:   c,
			Label:   model.CategoryLabel(c),
			Enabled: !disabled[c],
		})
	}
	return out, nil
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	open, err := s.st.IsOpen()
	if err != nil {
		fail(w, err)
		return
	}
	cats, err := s.categories(false)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"open": open, "categories": cats})
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Category model.Category `json:"category"`
		Name     string         `json:"name"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	if !model.ValidCategory(body.Category) {
		writeError(w, http.StatusBadRequest, "فئة غير معروفة")
		return
	}
	disabled, err := s.st.DisabledCategories()
	if err != nil {
		fail(w, err)
		return
	}
	if disabled[body.Category] {
		writeError(w, http.StatusConflict, "هذا الدور غير مفعّل في الاستبيان حاليًا")
		return
	}
	sess, err := s.st.CreateSession(body.Category, strings.TrimSpace(body.Name))
	if err != nil {
		fail(w, err)
		return
	}
	setSessionCookie(w, sess.ID)
	writeJSON(w, http.StatusCreated, sess)
}

// handleResumeSession يستأنف جلسة بكود المتابعة — المسار المستقل عن تخزين المتصفح.
func (s *Server) handleResumeSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	sess, err := s.st.SessionByCode(body.Code)
	if err != nil {
		fail(w, err)
		return
	}
	setSessionCookie(w, sess.ID)
	writeJSON(w, http.StatusOK, sess)
}

// handleLeaveSession ينسى هوية الجهاز ليبدأ الشخص نفسه استبيانًا بفئة أخرى.
// لا يحذف الجلسة ولا إجاباتها — الكود يبقى صالحًا للرجوع إليها.
func (s *Server) handleLeaveSession(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// handleCurrentSession يعيد الجلسة المرتبطة بالكوكي، لتنجو من مسح التخزين المحلي.
func (s *Server) handleCurrentSession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		writeError(w, http.StatusNotFound, "لا توجد جلسة محفوظة")
		return
	}
	sess, err := s.st.Session(c.Value)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

// sessionCookie اسم كوكي جلسة المشارك.
const sessionCookie = "survey_sid"

// setSessionCookie يثبّت هوية المشارك لستة أشهر — أطول بكثير من أسبوع الجمع.
func setSessionCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   180 * 24 * 3600,
	})
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	sess, err := s.st.Session(r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	questions, err := s.st.QuestionsFor(sess.Category)
	if err != nil {
		fail(w, err)
		return
	}
	answers, err := s.st.SessionAnswers(sess.ID)
	if err != nil {
		fail(w, err)
		return
	}
	open, err := s.st.IsOpen()
	if err != nil {
		fail(w, err)
		return
	}
	// مفاتيح JSON نصّية، لذا نحوّل معرّفات الأسئلة إلى نصوص.
	ans := make(map[string]json.RawMessage, len(answers))
	for id, v := range answers {
		ans[itoa(id)] = v
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"session":        sess,
		"category_label": model.CategoryLabel(sess.Category),
		"questions":      questions,
		"answers":        ans,
		"open":           open,
	})
}

func (s *Server) handleSaveAnswer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		QuestionID int64           `json:"question_id"`
		Value      json.RawMessage `json:"value"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	if err := s.st.SaveAnswer(r.PathValue("id"), body.QuestionID, body.Value); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}

func (s *Server) handleFinishSession(w http.ResponseWriter, r *http.Request) {
	if err := s.st.FinishSession(r.PathValue("id")); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"finished": true})
}

// ---------- مساعدات الاستجابة ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("تشفير الاستجابة: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// fail يترجم أخطاء الطبقة السفلية إلى رموز HTTP مناسبة.
func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "غير موجود")
	case errors.Is(err, store.ErrSurveyClosed):
		writeError(w, http.StatusConflict, "الاستبيان مقفل ولا يستقبل إجابات جديدة")
	default:
		log.Printf("خطأ داخلي: %v", err)
		writeError(w, http.StatusInternalServerError, "حدث خطأ في الخادم")
	}
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "طلب غير صالح")
		return false
	}
	return true
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
