// Package store هو المنفذ الوحيد للوصول إلى قاعدة البيانات. لا يوجد SQL خارج هذه الحزمة.
package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"surveyapp/internal/model"
)

// ErrSurveyClosed يُعاد عند محاولة الكتابة والاستبيان مقفل.
var ErrSurveyClosed = errors.New("الاستبيان مقفل")

// ErrNotFound يُعاد عند طلب سجل غير موجود.
var ErrNotFound = errors.New("غير موجود")

const (
	keySurveyOpen = "survey_open"
	keySeeded     = "seeded"
)

// Store يغلّف اتصال SQLite.
type Store struct {
	db *sql.DB
}

// Open يفتح قاعدة البيانات وينشئ الجداول إن لم تكن موجودة.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("فتح قاعدة البيانات: %w", err)
	}
	// SQLite لا يستفيد من اتصالات كتابة متوازية؛ اتصال واحد يمنع أخطاء القفل.
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close يغلق الاتصال.
func (s *Store) Close() error { return s.db.Close() }

const schema = `
CREATE TABLE IF NOT EXISTS questions (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  text       TEXT    NOT NULL,
  kind       TEXT    NOT NULL,
  section    TEXT    NOT NULL DEFAULT '',
  position   INTEGER NOT NULL DEFAULT 0,
  categories TEXT    NOT NULL DEFAULT '[]',
  options    TEXT    NOT NULL DEFAULT '[]',
  required   INTEGER NOT NULL DEFAULT 0,
  deleted    INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sessions (
  id          TEXT PRIMARY KEY,
  code        TEXT NOT NULL DEFAULT '',
  category    TEXT NOT NULL,
  name        TEXT NOT NULL DEFAULT '',
  started_at  TEXT NOT NULL,
  finished_at TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS answers (
  session_id  TEXT    NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  question_id INTEGER NOT NULL,
  value       TEXT    NOT NULL,
  answered_at TEXT    NOT NULL,
  PRIMARY KEY (session_id, question_id)
);

CREATE TABLE IF NOT EXISTS settings (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

`

// الفهارس منفصلة عن الجداول لأنها قد تشير إلى أعمدة تضيفها الترقية،
// فلا بد أن تُنفَّذ بعدها لا قبلها.
const indexes = `
CREATE INDEX IF NOT EXISTS idx_answers_question ON answers(question_id);
CREATE INDEX IF NOT EXISTS idx_sessions_category ON sessions(category);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_code ON sessions(code) WHERE code <> '';
`

// الأعمدة المضافة بعد الإصدار الأول. SQLite لا يدعم ADD COLUMN IF NOT EXISTS،
// فنتجاهل خطأ العمود المكرّر عند التشغيل على قاعدة مُرقّاة أصلًا.
var addedColumns = []string{
	`ALTER TABLE sessions ADD COLUMN code TEXT NOT NULL DEFAULT ''`,
}

func (s *Store) migrate() error {
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("إنشاء الجداول: %w", err)
	}
	for _, stmt := range addedColumns {
		if _, err := s.db.Exec(stmt); err != nil && !strings.Contains(err.Error(), "duplicate column") {
			return fmt.Errorf("ترقية الأعمدة: %w", err)
		}
	}
	if _, err := s.db.Exec(indexes); err != nil {
		return fmt.Errorf("إنشاء الفهارس: %w", err)
	}
	return s.backfillCodes()
}

// backfillCodes يمنح الجلسات القديمة أكواد متابعة حتى تعمل لها ميزة الاستئناف.
func (s *Store) backfillCodes() error {
	rows, err := s.db.Query(`SELECT id FROM sessions WHERE code = ''`)
	if err != nil {
		return err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, id := range ids {
		code, err := s.uniqueCode()
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(`UPDATE sessions SET code = ? WHERE id = ?`, code, id); err != nil {
			return err
		}
	}
	return nil
}

// ---------- الإعدادات ----------

// Setting يقرأ قيمة إعداد؛ يعيد def إن لم يوجد.
func (s *Store) Setting(key, def string) (string, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return def, nil
	}
	if err != nil {
		return "", err
	}
	return v, nil
}

// SetSetting يكتب قيمة إعداد.
func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO settings(key, value) VALUES(?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

// IsOpen يحدد ما إذا كان الاستبيان يستقبل إجابات. الافتراضي مفتوح.
func (s *Store) IsOpen() (bool, error) {
	v, err := s.Setting(keySurveyOpen, "1")
	return v == "1", err
}

// SetOpen يفتح أو يقفل استقبال الإجابات.
func (s *Store) SetOpen(open bool) error {
	v := "0"
	if open {
		v = "1"
	}
	return s.SetSetting(keySurveyOpen, v)
}

// Seeded يحدد ما إذا كانت الأسئلة الأولية حُمِّلت من قبل.
func (s *Store) Seeded() (bool, error) {
	v, err := s.Setting(keySeeded, "0")
	return v == "1", err
}

// MarkSeeded يسجّل أن الأسئلة الأولية حُمِّلت.
func (s *Store) MarkSeeded() error { return s.SetSetting(keySeeded, "1") }

// ---------- الأسئلة ----------

func scanQuestion(sc interface{ Scan(...any) error }) (model.Question, error) {
	var (
		q            model.Question
		catsJSON     string
		optsJSON     string
		req, deleted int
	)
	if err := sc.Scan(&q.ID, &q.Text, &q.Kind, &q.Section, &q.Position, &catsJSON, &optsJSON, &req, &deleted); err != nil {
		return q, err
	}
	q.Required = req == 1
	q.Deleted = deleted == 1
	if err := json.Unmarshal([]byte(catsJSON), &q.Categories); err != nil {
		q.Categories = nil
	}
	if err := json.Unmarshal([]byte(optsJSON), &q.Options); err != nil {
		q.Options = nil
	}
	if q.Categories == nil {
		q.Categories = []model.Category{}
	}
	if q.Options == nil {
		q.Options = []string{}
	}
	return q, nil
}

const questionCols = `id, text, kind, section, position, categories, options, required, deleted`

// Questions يعيد كل الأسئلة مرتّبة. includeDeleted يشمل المحذوفة منطقيًا.
func (s *Store) Questions(includeDeleted bool) ([]model.Question, error) {
	q := `SELECT ` + questionCols + ` FROM questions`
	if !includeDeleted {
		q += ` WHERE deleted = 0`
	}
	q += ` ORDER BY position, id`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Question{}
	for rows.Next() {
		item, err := scanQuestion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// Question يعيد سؤالًا واحدًا بالمعرّف.
func (s *Store) Question(id int64) (model.Question, error) {
	row := s.db.QueryRow(`SELECT `+questionCols+` FROM questions WHERE id = ?`, id)
	q, err := scanQuestion(row)
	if errors.Is(err, sql.ErrNoRows) {
		return q, ErrNotFound
	}
	return q, err
}

// QuestionsFor يعيد أسئلة فئة معيّنة مرتّبة.
func (s *Store) QuestionsFor(c model.Category) ([]model.Question, error) {
	all, err := s.Questions(false)
	if err != nil {
		return nil, err
	}
	out := []model.Question{}
	for _, q := range all {
		if q.AppliesTo(c) {
			out = append(out, q)
		}
	}
	return out, nil
}

// CreateQuestion يضيف سؤالًا في نهاية القائمة ويعيد معرّفه.
func (s *Store) CreateQuestion(q model.Question) (int64, error) {
	cats, _ := json.Marshal(nonNilCats(q.Categories))
	opts, _ := json.Marshal(nonNilStrs(q.Options))
	if q.Position == 0 {
		var maxPos sql.NullInt64
		if err := s.db.QueryRow(`SELECT MAX(position) FROM questions`).Scan(&maxPos); err != nil {
			return 0, err
		}
		q.Position = int(maxPos.Int64) + 1
	}
	res, err := s.db.Exec(
		`INSERT INTO questions(text, kind, section, position, categories, options, required, deleted)
		 VALUES(?, ?, ?, ?, ?, ?, ?, 0)`,
		q.Text, string(q.Kind), q.Section, q.Position, string(cats), string(opts), boolInt(q.Required))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateQuestion يعدّل سؤالًا قائمًا.
func (s *Store) UpdateQuestion(q model.Question) error {
	cats, _ := json.Marshal(nonNilCats(q.Categories))
	opts, _ := json.Marshal(nonNilStrs(q.Options))
	res, err := s.db.Exec(
		`UPDATE questions SET text = ?, kind = ?, section = ?, categories = ?, options = ?, required = ?
		 WHERE id = ?`,
		q.Text, string(q.Kind), q.Section, string(cats), string(opts), boolInt(q.Required), q.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteQuestion يحذف السؤال منطقيًا حتى لا تضيع الإجابات المجموعة عليه.
func (s *Store) DeleteQuestion(id int64) error {
	res, err := s.db.Exec(`UPDATE questions SET deleted = 1 WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// SetSectionActive يفعّل قسمًا كاملًا أو يعطّله، ليُشغَّل الاستبيان الطويل
// على موجات قصيرة بدل عرض كل الأقسام على المشارك دفعة واحدة.
// التعطيل حذف منطقي، فالإجابات المجموعة تبقى في النتائج والتصدير.
func (s *Store) SetSectionActive(section string, active bool) (int64, error) {
	res, err := s.db.Exec(`UPDATE questions SET deleted = ? WHERE section = ?`,
		boolInt(!active), section)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Sections يعيد أقسام الأسئلة بترتيب ظهورها مع عدد أسئلة كل قسم وحالته.
func (s *Store) Sections() ([]Section, error) {
	rows, err := s.db.Query(
		`SELECT section, COUNT(*), SUM(CASE WHEN deleted = 0 THEN 1 ELSE 0 END), MIN(position)
		 FROM questions GROUP BY section ORDER BY MIN(position)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Section{}
	for rows.Next() {
		var (
			sec           Section
			total, active int
			pos           int
		)
		if err := rows.Scan(&sec.Name, &total, &active, &pos); err != nil {
			return nil, err
		}
		sec.Total = total
		sec.Active = active
		out = append(out, sec)
	}
	return out, rows.Err()
}

// Section ملخّص قسم أسئلة.
type Section struct {
	Name   string `json:"name"`
	Total  int    `json:"total"`
	Active int    `json:"active"`
}

// ReorderQuestions يعيد ترتيب الأسئلة حسب تسلسل المعرّفات المعطى.
func (s *Store) ReorderQuestions(ids []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		if _, err := tx.Exec(`UPDATE questions SET position = ? WHERE id = ?`, i+1, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ---------- الجلسات ----------

// CreateSession ينشئ جلسة مشارك جديدة.
func (s *Store) CreateSession(c model.Category, name string) (model.Session, error) {
	open, err := s.IsOpen()
	if err != nil {
		return model.Session{}, err
	}
	if !open {
		return model.Session{}, ErrSurveyClosed
	}
	code, err := s.uniqueCode()
	if err != nil {
		return model.Session{}, err
	}
	sess := model.Session{
		ID:        newID(),
		Code:      code,
		Category:  c,
		Name:      name,
		StartedAt: now(),
	}
	_, err = s.db.Exec(
		`INSERT INTO sessions(id, code, category, name, started_at, finished_at) VALUES(?, ?, ?, ?, ?, '')`,
		sess.ID, sess.Code, string(sess.Category), sess.Name, sess.StartedAt)
	if err != nil {
		return model.Session{}, err
	}
	return sess, nil
}

const sessionCols = `id, code, category, name, started_at, finished_at`

// Session يعيد جلسة بالمعرّف.
func (s *Store) Session(id string) (model.Session, error) {
	return s.sessionWhere(`id = ?`, id)
}

// SessionByCode يعيد جلسة بكود المتابعة. الكود غير حسّاس لحالة الأحرف أو المسافات
// لأن المشارك يكتبه بيده من ورقة أو رسالة.
func (s *Store) SessionByCode(code string) (model.Session, error) {
	code = NormalizeCode(code)
	if code == "" {
		return model.Session{}, ErrNotFound
	}
	return s.sessionWhere(`code = ?`, code)
}

func (s *Store) sessionWhere(cond string, arg any) (model.Session, error) {
	var sess model.Session
	err := s.db.QueryRow(`SELECT `+sessionCols+` FROM sessions WHERE `+cond, arg).
		Scan(&sess.ID, &sess.Code, &sess.Category, &sess.Name, &sess.StartedAt, &sess.FinishedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return sess, ErrNotFound
	}
	return sess, err
}

// NormalizeCode يوحّد شكل الكود المُدخل قبل مطابقته.
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// FinishSession يسجّل انتهاء الجلسة.
func (s *Store) FinishSession(id string) error {
	res, err := s.db.Exec(`UPDATE sessions SET finished_at = ? WHERE id = ? AND finished_at = ''`, now(), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		// إمّا الجلسة غير موجودة أو منتهية أصلًا؛ الأخيرة ليست خطأً.
		if _, err := s.Session(id); err != nil {
			return err
		}
	}
	return nil
}

// Sessions يعيد الجلسات، ويُفلتر بالفئة إن لم تكن فارغة.
func (s *Store) Sessions(c model.Category) ([]model.Session, error) {
	var (
		rows *sql.Rows
		err  error
	)
	base := `SELECT ` + sessionCols + ` FROM sessions`
	if c == "" {
		rows, err = s.db.Query(base + ` ORDER BY started_at`)
	} else {
		rows, err = s.db.Query(base+` WHERE category = ? ORDER BY started_at`, string(c))
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Session{}
	for rows.Next() {
		var sess model.Session
		if err := rows.Scan(&sess.ID, &sess.Code, &sess.Category, &sess.Name, &sess.StartedAt, &sess.FinishedAt); err != nil {
			return nil, err
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

// DeleteSession يحذف جلسة وإجاباتها.
func (s *Store) DeleteSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM answers WHERE session_id = ?`, id)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- الإجابات ----------

// SaveAnswer يحفظ أو يحدّث إجابة. يرفض الحفظ إن كان الاستبيان مقفلًا.
func (s *Store) SaveAnswer(sessionID string, questionID int64, value json.RawMessage) error {
	open, err := s.IsOpen()
	if err != nil {
		return err
	}
	if !open {
		return ErrSurveyClosed
	}
	if _, err := s.Session(sessionID); err != nil {
		return err
	}
	if _, err := s.Question(questionID); err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO answers(session_id, question_id, value, answered_at) VALUES(?, ?, ?, ?)
		 ON CONFLICT(session_id, question_id) DO UPDATE SET value = excluded.value, answered_at = excluded.answered_at`,
		sessionID, questionID, string(value), now())
	return err
}

// SessionAnswers يعيد إجابات جلسة مفهرسة بمعرّف السؤال.
func (s *Store) SessionAnswers(sessionID string) (map[int64]json.RawMessage, error) {
	rows, err := s.db.Query(`SELECT question_id, value FROM answers WHERE session_id = ?`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]json.RawMessage{}
	for rows.Next() {
		var (
			qid int64
			val string
		)
		if err := rows.Scan(&qid, &val); err != nil {
			return nil, err
		}
		out[qid] = json.RawMessage(val)
	}
	return out, rows.Err()
}

// AllAnswers يعيد كل الإجابات مفهرسة بمعرّف الجلسة ثم معرّف السؤال.
func (s *Store) AllAnswers() (map[string]map[int64]json.RawMessage, error) {
	rows, err := s.db.Query(`SELECT session_id, question_id, value FROM answers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]map[int64]json.RawMessage{}
	for rows.Next() {
		var (
			sid string
			qid int64
			val string
		)
		if err := rows.Scan(&sid, &qid, &val); err != nil {
			return nil, err
		}
		if out[sid] == nil {
			out[sid] = map[int64]json.RawMessage{}
		}
		out[sid][qid] = json.RawMessage(val)
	}
	return out, rows.Err()
}

// AnsweredQuestionIDs يعيد معرّفات الأسئلة التي عليها إجابة واحدة على الأقل.
func (s *Store) AnsweredQuestionIDs() (map[int64]bool, error) {
	rows, err := s.db.Query(`SELECT DISTINCT question_id FROM answers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// Stats إحصاءات مختصرة لكل فئة.
type Stats struct {
	Category model.Category `json:"category"`
	Label    string         `json:"label"`
	Started  int            `json:"started"`
	Finished int            `json:"finished"`
}

// CategoryStats يعيد عدد الجلسات المبدوءة والمنتهية لكل فئة.
func (s *Store) CategoryStats() ([]Stats, error) {
	rows, err := s.db.Query(
		`SELECT category, COUNT(*), SUM(CASE WHEN finished_at <> '' THEN 1 ELSE 0 END)
		 FROM sessions GROUP BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byCat := map[model.Category]Stats{}
	for rows.Next() {
		var (
			cat         string
			total, done int
		)
		if err := rows.Scan(&cat, &total, &done); err != nil {
			return nil, err
		}
		byCat[model.Category(cat)] = Stats{Started: total, Finished: done}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]Stats, 0, len(model.AllCategories))
	for _, c := range model.AllCategories {
		st := byCat[c]
		st.Category = c
		st.Label = model.CategoryLabel(c)
		out = append(out, st)
	}
	return out, nil
}

// ---------- مساعدات ----------

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nonNilCats(c []model.Category) []model.Category {
	if c == nil {
		return []model.Category{}
	}
	return c
}

func nonNilStrs(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

// codeAlphabet يستبعد الأحرف الملتبسة (O/0، I/1/L) لأن الكود يُنسخ يدويًا.
const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

const codeLength = 6

// uniqueCode يولّد كود متابعة غير مستخدم. الاصطدام نادر جدًا، لكن المحاولات
// المحدودة تمنع أي احتمال لحلقة لا نهائية.
func (s *Store) uniqueCode() (string, error) {
	for attempt := 0; attempt < 20; attempt++ {
		code, err := randomCode()
		if err != nil {
			return "", err
		}
		var exists int
		err = s.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE code = ?`, code).Scan(&exists)
		if err != nil {
			return "", err
		}
		if exists == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("تعذّر توليد كود متابعة غير مكرّر")
}

func randomCode() (string, error) {
	b := make([]byte, codeLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("توليد كود المتابعة: %w", err)
	}
	out := make([]byte, codeLength)
	for i, v := range b {
		out[i] = codeAlphabet[int(v)%len(codeAlphabet)]
	}
	return string(out), nil
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand لا يفشل عمليًا؛ السقوط على الوقت أفضل من الانهيار.
		return fmt.Sprintf("t%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
