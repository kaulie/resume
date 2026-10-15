package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

func loadResumeContent(ctx context.Context, a *app, id string) (ResumeContent, error) {
	var raw []byte
	err := a.pool.QueryRow(ctx, `SELECT content FROM resumes WHERE id = $1`, id).Scan(&raw)
	if err != nil {
		return ResumeContent{}, err
	}
	var c ResumeContent
	_ = json.Unmarshal(raw, &c)
	return c, nil
}

func loadJD(ctx context.Context, a *app, id string) (JD, error) {
	return scanJD(a.pool.QueryRow(ctx, `SELECT `+jdColumns+` FROM jds WHERE id = $1`, id))
}

func (a *app) handleOptimize(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ResumeID string `json:"resumeId"`
		JDID     string `json:"jdId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.writeErr(w, 400, err.Error())
		return
	}
	resume, err := loadResumeContent(r.Context(), a, body.ResumeID)
	if errors.Is(err, pgx.ErrNoRows) {
		a.writeErr(w, 404, "简历不存在")
		return
	}
	if err != nil {
		a.writeErr(w, 500, err.Error())
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

	var content ResumeContent
	var diff []DiffItem
	settings := loadSettings(r.Context(), a.pool)
	if settings.LLMEnabled && settings.LLMAPIKey != "" {
		if c, d, e := optimizeWithLLM(r.Context(), settings, resume, jd); e == nil {
			content, diff = c, d
		} else {
			content, diff = buildOptimization(resume, jd)
		}
	} else {
		content, diff = buildOptimization(resume, jd)
	}

	id := newID()
	now := nowISO()
	label := "针对「" + jd.Title + "」优化版"
	contentJSON, _ := json.Marshal(content)
	diffJSON, _ := json.Marshal(diff)
	_, err = a.pool.Exec(r.Context(), `INSERT INTO resume_versions (id, resume_id, jd_id, label, content, diff, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, body.ResumeID, body.JDID, label, contentJSON, diffJSON, now)
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	v := ResumeVersion{ID: id, ResumeID: body.ResumeID, JDID: &body.JDID, Label: label, Content: content, Diff: diff, CreatedAt: now}
	a.writeJSON(w, 201, v)
}

const versionColumns = `id, resume_id, jd_id, label, content, diff, created_at`

func scanVersion(row pgx.Row) (ResumeVersion, error) {
	var v ResumeVersion
	var content, diff []byte
	var createdAt time.Time
	var jdID *string
	if err := row.Scan(&v.ID, &v.ResumeID, &jdID, &v.Label, &content, &diff, &createdAt); err != nil {
		return v, err
	}
	v.JDID = jdID
	_ = json.Unmarshal(content, &v.Content)
	_ = json.Unmarshal(diff, &v.Diff)
	v.CreatedAt = createdAt.Format(time.RFC3339Nano)
	return v, nil
}

func (a *app) handleListVersions(w http.ResponseWriter, r *http.Request) {
	resumeID := r.URL.Query().Get("resumeId")
	var rows pgx.Rows
	var err error
	if resumeID != "" {
		rows, err = a.pool.Query(r.Context(), `SELECT `+versionColumns+` FROM resume_versions WHERE resume_id = $1 ORDER BY created_at DESC`, resumeID)
	} else {
		rows, err = a.pool.Query(r.Context(), `SELECT `+versionColumns+` FROM resume_versions ORDER BY created_at DESC`)
	}
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	defer rows.Close()
	out := []ResumeVersion{}
	for rows.Next() {
		var v ResumeVersion
		var content, diff []byte
		var createdAt time.Time
		var jdID *string
		if err := rows.Scan(&v.ID, &v.ResumeID, &jdID, &v.Label, &content, &diff, &createdAt); err != nil {
			a.writeErr(w, 500, err.Error())
			return
		}
		v.JDID = jdID
		_ = json.Unmarshal(content, &v.Content)
		_ = json.Unmarshal(diff, &v.Diff)
		v.CreatedAt = createdAt.Format(time.RFC3339Nano)
		out = append(out, v)
	}
	a.writeJSON(w, 200, out)
}

func (a *app) handleGetVersion(w http.ResponseWriter, r *http.Request) {
	v, err := scanVersion(a.pool.QueryRow(r.Context(), `SELECT `+versionColumns+` FROM resume_versions WHERE id = $1`, r.PathValue("id")))
	if errors.Is(err, pgx.ErrNoRows) {
		a.writeErr(w, 404, "版本不存在")
		return
	}
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	a.writeJSON(w, 200, v)
}

func (a *app) handleDeleteVersion(w http.ResponseWriter, r *http.Request) {
	_, _ = a.pool.Exec(r.Context(), `DELETE FROM resume_versions WHERE id = $1`, r.PathValue("id"))
	a.writeJSON(w, 200, map[string]bool{"ok": true})
}
