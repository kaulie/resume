package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
)

const prepColumns = `id, jd_id, content, created_at`

func (a *app) handleCreateInterviewPrep(w http.ResponseWriter, r *http.Request) {
	var body struct {
		JDID string `json:"jdId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.writeErr(w, 400, err.Error())
		return
	}
	jd, err := loadJD(r.Context(), a, body.JDID)
	if errors.Is(err, pgx.ErrNoRows) {
		a.writeErr(w, 404, "JD 不存在")
		return
	}
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}

	settings := loadSettings(r.Context(), a.pool)
	var content InterviewPrepContent
	if settings.LLMEnabled && settings.LLMAPIKey != "" {
		if c, e := prepWithLLM(r.Context(), settings, jd); e == nil {
			content = c
		} else {
			content = buildInterviewPrep(jd)
		}
	} else {
		content = buildInterviewPrep(jd)
	}

	id := newID()
	now := nowISO()
	contentJSON, _ := json.Marshal(content)
	_, err = a.pool.Exec(r.Context(), `INSERT INTO interview_preps (id, jd_id, content, created_at) VALUES ($1, $2, $3, $4)`, id, body.JDID, contentJSON, now)
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	a.writeJSON(w, 201, InterviewPrep{ID: id, JDID: body.JDID, Content: content, CreatedAt: now})
}

func (a *app) handleListInterviewPrep(w http.ResponseWriter, r *http.Request) {
	jdID := r.URL.Query().Get("jdId")
	var rows pgx.Rows
	var err error
	if jdID != "" {
		rows, err = a.pool.Query(r.Context(), `SELECT `+prepColumns+` FROM interview_preps WHERE jd_id = $1 ORDER BY created_at DESC`, jdID)
	} else {
		rows, err = a.pool.Query(r.Context(), `SELECT `+prepColumns+` FROM interview_preps ORDER BY created_at DESC`)
	}
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	defer rows.Close()
	out := []InterviewPrep{}
	for rows.Next() {
		var p InterviewPrep
		var content []byte
		var createdAt time.Time
		if err := rows.Scan(&p.ID, &p.JDID, &content, &createdAt); err != nil {
			a.writeErr(w, 500, err.Error())
			return
		}
		_ = json.Unmarshal(content, &p.Content)
		p.CreatedAt = createdAt.Format(time.RFC3339Nano)
		out = append(out, p)
	}
	a.writeJSON(w, 200, out)
}

func (a *app) handleGetInterviewPrep(w http.ResponseWriter, r *http.Request) {
	var p InterviewPrep
	var content []byte
	var createdAt time.Time
	err := a.pool.QueryRow(r.Context(), `SELECT `+prepColumns+` FROM interview_preps WHERE id = $1`, r.PathValue("id")).Scan(&p.ID, &p.JDID, &content, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		a.writeErr(w, 404, "备战建议不存在")
		return
	}
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	_ = json.Unmarshal(content, &p.Content)
	p.CreatedAt = createdAt.Format(time.RFC3339Nano)
	a.writeJSON(w, 200, p)
}

var unsafeNameRe = regexp.MustCompile(`[\\/:*?"<>|\s]+`)

func sanitizeName(name string) string {
	s := unsafeNameRe.ReplaceAllString(name, "_")
	r := []rune(s)
	if len(r) > 60 {
		s = string(r[:60])
	}
	if s == "" {
		s = "resume"
	}
	return s
}

func (a *app) handleExport(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "md"
	}
	versionID := r.URL.Query().Get("versionId")
	resumeID := r.URL.Query().Get("resumeId")

	var content ResumeContent
	var title string

	if versionID != "" {
		v, err := scanVersion(a.pool.QueryRow(r.Context(), `SELECT `+versionColumns+` FROM resume_versions WHERE id = $1`, versionID))
		if errors.Is(err, pgx.ErrNoRows) {
			a.writeErr(w, 404, "版本不存在")
			return
		}
		if err != nil {
			a.writeErr(w, 500, err.Error())
			return
		}
		content, title = v.Content, v.Label
	} else {
		var raw []byte
		var t string
		var err error
		if resumeID != "" {
			err = a.pool.QueryRow(r.Context(), `SELECT title, content FROM resumes WHERE id = $1`, resumeID).Scan(&t, &raw)
		} else {
			err = a.pool.QueryRow(r.Context(), `SELECT title, content FROM resumes ORDER BY updated_at DESC LIMIT 1`).Scan(&t, &raw)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			a.writeErr(w, 404, "简历不存在")
			return
		}
		if err != nil {
			a.writeErr(w, 500, err.Error())
			return
		}
		_ = json.Unmarshal(raw, &content)
		title = t
	}

	var buf []byte
	var mime, ext string
	var err error
	switch format {
	case "md", "markdown":
		buf, mime, ext = []byte(contentToMarkdown(content)), "text/markdown; charset=utf-8", "md"
	case "txt":
		buf, mime, ext = []byte(contentToText(content)), "text/plain; charset=utf-8", "txt"
	case "docx":
		buf, err = exportDocx(content)
		mime, ext = "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "docx"
	case "pdf":
		settings := loadSettings(r.Context(), a.pool)
		buf, err = exportPDF(content, settings.PDFFontPath)
		mime, ext = "application/pdf", "pdf"
	default:
		a.writeErr(w, 400, "不支持的导出格式")
		return
	}
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", `attachment; filename="`+sanitizeName(title)+"."+ext+`"`)
	_, _ = w.Write(buf)
}

func (a *app) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	a.writeJSON(w, 200, loadSettings(r.Context(), a.pool))
}

func (a *app) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	var s Settings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		a.writeErr(w, 400, err.Error())
		return
	}
	if s.LLMBaseURL == "" {
		s.LLMBaseURL = "https://api.deepseek.com"
	}
	if s.LLMModel == "" {
		s.LLMModel = "deepseek-chat"
	}
	if err := saveSettings(r.Context(), a.pool, s); err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	a.writeJSON(w, 200, s)
}
