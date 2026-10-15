package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const evalColumns = `id, resume_id, jd_id, overall_score, dimensions, suggestions, created_at`

func scanEvaluation(row pgx.Row) (Evaluation, error) {
	var e Evaluation
	var dimensions, suggestions []byte
	var createdAt time.Time
	if err := row.Scan(&e.ID, &e.ResumeID, &e.JDID, &e.OverallScore, &dimensions, &suggestions, &createdAt); err != nil {
		return e, err
	}
	_ = json.Unmarshal(dimensions, &e.Dimensions)
	_ = json.Unmarshal(suggestions, &e.Suggestions)
	e.CreatedAt = createdAt.Format(time.RFC3339Nano)
	return e, nil
}

func (a *app) handleCreateEvaluation(w http.ResponseWriter, r *http.Request) {
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

	settings := loadSettings(r.Context(), a.pool)
	var overall int
	var dimensions []Dimension
	var suggestions []string
	if settings.LLMEnabled && settings.LLMAPIKey != "" {
		if o, d, s, e := evaluateWithLLM(r.Context(), settings, resume, jd); e == nil {
			overall, dimensions, suggestions = o, d, s
		} else {
			overall, dimensions, suggestions = buildEvaluation(resume, jd)
		}
	} else {
		overall, dimensions, suggestions = buildEvaluation(resume, jd)
	}

	id := newID()
	now := nowISO()
	dimJSON, _ := json.Marshal(dimensions)
	sugJSON, _ := json.Marshal(suggestions)
	_, err = a.pool.Exec(r.Context(), `INSERT INTO evaluations (id, resume_id, jd_id, overall_score, dimensions, suggestions, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, body.ResumeID, body.JDID, overall, dimJSON, sugJSON, now)
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	a.writeJSON(w, 201, Evaluation{ID: id, ResumeID: body.ResumeID, JDID: body.JDID, OverallScore: overall, Dimensions: dimensions, Suggestions: suggestions, CreatedAt: now})
}

func (a *app) handleListEvaluations(w http.ResponseWriter, r *http.Request) {
	resumeID := r.URL.Query().Get("resumeId")
	jdID := r.URL.Query().Get("jdId")
	q := `SELECT ` + evalColumns + ` FROM evaluations`
	var args []any
	var conds []string
	if resumeID != "" {
		args = append(args, resumeID)
		conds = append(conds, "resume_id = $"+itoa(len(args)))
	}
	if jdID != "" {
		args = append(args, jdID)
		conds = append(conds, "jd_id = $"+itoa(len(args)))
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY created_at DESC"

	rows, err := a.pool.Query(r.Context(), q, args...)
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	defer rows.Close()
	out := []Evaluation{}
	for rows.Next() {
		var e Evaluation
		var dimensions, suggestions []byte
		var createdAt time.Time
		if err := rows.Scan(&e.ID, &e.ResumeID, &e.JDID, &e.OverallScore, &dimensions, &suggestions, &createdAt); err != nil {
			a.writeErr(w, 500, err.Error())
			return
		}
		_ = json.Unmarshal(dimensions, &e.Dimensions)
		_ = json.Unmarshal(suggestions, &e.Suggestions)
		e.CreatedAt = createdAt.Format(time.RFC3339Nano)
		out = append(out, e)
	}
	a.writeJSON(w, 200, out)
}

func (a *app) handleGetEvaluation(w http.ResponseWriter, r *http.Request) {
	e, err := scanEvaluation(a.pool.QueryRow(r.Context(), `SELECT `+evalColumns+` FROM evaluations WHERE id = $1`, r.PathValue("id")))
	if errors.Is(err, pgx.ErrNoRows) {
		a.writeErr(w, 404, "评价不存在")
		return
	}
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	a.writeJSON(w, 200, e)
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
