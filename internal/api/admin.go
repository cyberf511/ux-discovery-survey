package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"surveyapp/internal/export"
	"surveyapp/internal/model"
)

const adminCookie = "survey_admin"

// tokenFor يشتق قيمة الكوكي من كلمة المرور. تغيير كلمة المرور يُبطل الجلسات القائمة.
func tokenFor(password string) string {
	mac := hmac.New(sha256.New, []byte(password))
	mac.Write([]byte("survey-admin-v1"))
	return hex.EncodeToString(mac.Sum(nil))
}

// auth يحمي معالجات الإدارة بكوكي موقّع.
func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(adminCookie)
		if err != nil || !hmac.Equal([]byte(c.Value), []byte(s.authTok)) {
			writeError(w, http.StatusUnauthorized, "يلزم تسجيل الدخول")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	// مقارنة ثابتة الزمن تمنع استنتاج كلمة المرور من فروق التوقيت.
	if !hmac.Equal([]byte(tokenFor(body.Password)), []byte(s.authTok)) {
		writeError(w, http.StatusUnauthorized, "كلمة المرور غير صحيحة")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookie,
		Value:    s.authTok,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   14 * 24 * 3600,
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: adminCookie, Value: "", Path: "/", MaxAge: -1})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleAdminState(w http.ResponseWriter, r *http.Request) {
	questions, err := s.st.Questions(false)
	if err != nil {
		fail(w, err)
		return
	}
	sections, err := s.st.Sections()
	if err != nil {
		fail(w, err)
		return
	}
	stats, err := s.st.CategoryStats()
	if err != nil {
		fail(w, err)
		return
	}
	open, err := s.st.IsOpen()
	if err != nil {
		fail(w, err)
		return
	}
	cats, err := s.categories(true)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"questions":  questions,
		"sections":   sections,
		"stats":      stats,
		"open":       open,
		"categories": cats,
	})
}

// questionInput الحمولة المقبولة لإنشاء أو تعديل سؤال.
type questionInput struct {
	Text       string           `json:"text"`
	Kind       model.Kind       `json:"kind"`
	Section    string           `json:"section"`
	Categories []model.Category `json:"categories"`
	Options    []string         `json:"options"`
	Required   bool             `json:"required"`
}

func (in questionInput) toQuestion() (model.Question, error) {
	q := model.Question{
		Text:       strings.TrimSpace(in.Text),
		Kind:       in.Kind,
		Section:    strings.TrimSpace(in.Section),
		Categories: in.Categories,
		Options:    cleanOptions(in.Options),
		Required:   in.Required,
	}
	if q.Text == "" {
		return q, fmt.Errorf("نص السؤال مطلوب")
	}
	if !model.ValidKind(q.Kind) {
		return q, fmt.Errorf("نوع السؤال غير مدعوم")
	}
	if model.NeedsOptions(q.Kind) && len(q.Options) < 2 {
		return q, fmt.Errorf("هذا النوع يحتاج خيارين على الأقل")
	}
	for _, c := range q.Categories {
		if !model.ValidCategory(c) {
			return q, fmt.Errorf("فئة غير معروفة: %s", c)
		}
	}
	return q, nil
}

func cleanOptions(opts []string) []string {
	out := []string{}
	for _, o := range opts {
		if t := strings.TrimSpace(o); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func (s *Server) handleCreateQuestion(w http.ResponseWriter, r *http.Request) {
	var in questionInput
	if !readJSON(w, r, &in) {
		return
	}
	q, err := in.toQuestion()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, err := s.st.CreateQuestion(q)
	if err != nil {
		fail(w, err)
		return
	}
	q.ID = id
	writeJSON(w, http.StatusCreated, q)
}

func (s *Server) handleUpdateQuestion(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "معرّف غير صالح")
		return
	}
	var in questionInput
	if !readJSON(w, r, &in) {
		return
	}
	q, err := in.toQuestion()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	q.ID = id
	if err := s.st.UpdateQuestion(q); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, q)
}

func (s *Server) handleDeleteQuestion(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "معرّف غير صالح")
		return
	}
	if err := s.st.DeleteQuestion(id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *Server) handleReorder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []int64 `json:"ids"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	if err := s.st.ReorderQuestions(body.IDs); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// handleToggleCategory يفعّل دورًا أو يعطّله. الدور المعطّل يختفي من شاشة
// اختيار الفئة، وتختفي أسئلته معه لأنها موسومة به وحده.
func (s *Server) handleToggleCategory(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Category model.Category `json:"category"`
		Enabled  bool           `json:"enabled"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	if !model.ValidCategory(body.Category) {
		writeError(w, http.StatusBadRequest, "فئة غير معروفة")
		return
	}
	if err := s.st.SetCategoryEnabled(body.Category, body.Enabled); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"category": body.Category, "enabled": body.Enabled})
}

// handleToggleSection يفعّل قسمًا كاملًا أو يعطّله لتشغيل الاستبيان على موجات.
func (s *Server) handleToggleSection(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Section string `json:"section"`
		Active  bool   `json:"active"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.Section) == "" {
		writeError(w, http.StatusBadRequest, "اسم القسم مطلوب")
		return
	}
	n, err := s.st.SetSectionActive(body.Section, body.Active)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"affected": n})
}

// handleImport يستورد دفعة أسئلة نصية: سطر يبدأ بـ ## يصبح عنوان قسم، وكل سطر آخر سؤال نصّي طويل.
func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text       string           `json:"text"`
		Categories []model.Category `json:"categories"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	for _, c := range body.Categories {
		if !model.ValidCategory(c) {
			writeError(w, http.StatusBadRequest, "فئة غير معروفة: "+string(c))
			return
		}
	}
	section := ""
	added := 0
	for _, line := range strings.Split(body.Text, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "*"))
		line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "##") {
			section = strings.TrimSpace(strings.TrimPrefix(line, "##"))
			continue
		}
		q := model.Question{
			Text:       line,
			Kind:       model.KindLongText,
			Section:    section,
			Categories: body.Categories,
			Options:    []string{},
		}
		if _, err := s.st.CreateQuestion(q); err != nil {
			fail(w, err)
			return
		}
		added++
	}
	writeJSON(w, http.StatusOK, map[string]any{"added": added})
}

func (s *Server) handleSetOpen(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Open bool `json:"open"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	if err := s.st.SetOpen(body.Open); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"open": body.Open})
}

func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	if err := s.st.DeleteSession(r.PathValue("id")); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

// exportData يجمع الأسئلة والجلسات والإجابات لفئة معيّنة (أو للكل عند category فارغة).
func (s *Server) exportData(category model.Category) (export.Data, error) {
	all, err := s.st.Questions(true)
	if err != nil {
		return export.Data{}, err
	}
	answered, err := s.st.AnsweredQuestionIDs()
	if err != nil {
		return export.Data{}, err
	}
	questions := []model.Question{}
	for _, q := range all {
		// السؤال المحذوف يبقى في التصدير ما دامت عليه إجابات مجموعة.
		if q.Deleted && !answered[q.ID] {
			continue
		}
		if category != "" && !q.AppliesTo(category) {
			continue
		}
		questions = append(questions, q)
	}
	sessions, err := s.st.Sessions(category)
	if err != nil {
		return export.Data{}, err
	}
	answers, err := s.st.AllAnswers()
	if err != nil {
		return export.Data{}, err
	}
	times, err := s.st.AnswerTimes()
	if err != nil {
		return export.Data{}, err
	}
	return export.Data{
		Questions:  questions,
		Sessions:   sessions,
		Answers:    answers,
		AnsweredAt: times,
	}, nil
}

func (s *Server) handleResults(w http.ResponseWriter, r *http.Request) {
	category := model.Category(r.URL.Query().Get("category"))
	if category != "" && !model.ValidCategory(category) {
		writeError(w, http.StatusBadRequest, "فئة غير معروفة")
		return
	}
	d, err := s.exportData(category)
	if err != nil {
		fail(w, err)
		return
	}
	type row struct {
		Question model.Question `json:"question"`
		Answers  []answerView   `json:"answers"`
	}
	out := make([]row, 0, len(d.Questions))
	for _, q := range d.Questions {
		views := []answerView{}
		for _, sess := range d.Sessions {
			raw := d.Answers[sess.ID][q.ID]
			text := export.FormatValue(q, raw)
			if text == "" {
				continue
			}
			views = append(views, answerView{
				SessionID: sess.ID,
				Name:      sess.Name,
				Category:  model.CategoryLabel(sess.Category),
				Text:      text,
			})
		}
		out = append(out, row{Question: q, Answers: views})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sessions": sessionViews(d.Sessions, d.Answers),
		"rows":     out,
	})
}

type answerView struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Text      string `json:"text"`
}

type sessionView struct {
	model.Session
	Label       string `json:"label"`
	AnswerCount int    `json:"answer_count"`
	Done        bool   `json:"done"`
}

func sessionViews(sessions []model.Session, answers map[string]map[int64]json.RawMessage) []sessionView {
	out := make([]sessionView, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, sessionView{
			Session:     s,
			Label:       model.CategoryLabel(s.Category),
			AnswerCount: len(answers[s.ID]),
			Done:        s.Finished(),
		})
	}
	return out
}

func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	s.exportCSV(w, r, "wide", export.WriteCSV)
}

// handleExportLongCSV يصدّر صفًا لكل إجابة — الشكل الصالح للتحليل والفهرسة.
func (s *Server) handleExportLongCSV(w http.ResponseWriter, r *http.Request) {
	s.exportCSV(w, r, "long", export.WriteLongCSV)
}

func (s *Server) exportCSV(w http.ResponseWriter, r *http.Request, shape string,
	write func(io.Writer, export.Data) error) {
	category := model.Category(r.URL.Query().Get("category"))
	if category != "" && !model.ValidCategory(category) {
		writeError(w, http.StatusBadRequest, "فئة غير معروفة")
		return
	}
	d, err := s.exportData(category)
	if err != nil {
		fail(w, err)
		return
	}
	name := "all"
	if category != "" {
		name = model.CategoryLabel(category)
	}
	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("survey-%s-%s-%s.csv", shape, name, date)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	// filename* بترميز RFC 5987 حتى تصل الأسماء العربية سليمة إلى المتصفح.
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="survey-%s-%s.csv"; filename*=UTF-8''%s`,
			shape, date, urlEncode(filename)))
	if err := write(w, d); err != nil {
		// الترويسات أُرسلت بالفعل، فلا يمكن إرجاع رمز خطأ.
		fmt.Fprintf(w, "\n#خطأ أثناء التصدير: %v\n", err)
	}
}

func (s *Server) handleBackup(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(s.dbPath)
	if err != nil {
		fail(w, err)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="survey-backup-%s.db"`, time.Now().Format("2006-01-02")))
	if _, err := io.Copy(w, f); err != nil {
		fmt.Fprintf(w, "\n#خطأ أثناء النسخ: %v\n", err)
	}
}

func urlEncode(s string) string {
	const hexDigits = "0123456789ABCDEF"
	var b strings.Builder
	for _, c := range []byte(s) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hexDigits[c>>4])
		b.WriteByte(hexDigits[c&0x0F])
	}
	return b.String()
}
