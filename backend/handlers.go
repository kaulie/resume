package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type app struct {
	pool *pgxpool.Pool
}

func (a *app) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (a *app) writeErr(w http.ResponseWriter, status int, msg string) {
	a.writeJSON(w, status, map[string]string{"error": msg})
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", a.handleHealth)
	mux.HandleFunc("GET /api/meta", a.handleMeta)

	mux.HandleFunc("GET /api/resumes", a.handleListResumes)
	mux.HandleFunc("POST /api/resumes", a.handleCreateResume)
	mux.HandleFunc("GET /api/resumes/{id}", a.handleGetResume)
	mux.HandleFunc("PUT /api/resumes/{id}", a.handleUpdateResume)
	mux.HandleFunc("DELETE /api/resumes/{id}", a.handleDeleteResume)

	mux.HandleFunc("GET /api/jds", a.handleListJDs)
	mux.HandleFunc("POST /api/jds", a.handleCreateJD)
	mux.HandleFunc("GET /api/jds/{id}", a.handleGetJD)
	mux.HandleFunc("PUT /api/jds/{id}", a.handleUpdateJD)
	mux.HandleFunc("DELETE /api/jds/{id}", a.handleDeleteJD)

	mux.HandleFunc("POST /api/optimize", a.handleOptimize)
	mux.HandleFunc("GET /api/versions", a.handleListVersions)
	mux.HandleFunc("GET /api/versions/{id}", a.handleGetVersion)
	mux.HandleFunc("DELETE /api/versions/{id}", a.handleDeleteVersion)

	mux.HandleFunc("POST /api/evaluations", a.handleCreateEvaluation)
	mux.HandleFunc("GET /api/evaluations", a.handleListEvaluations)
	mux.HandleFunc("GET /api/evaluations/{id}", a.handleGetEvaluation)

	mux.HandleFunc("POST /api/interview-prep", a.handleCreateInterviewPrep)
	mux.HandleFunc("GET /api/interview-prep", a.handleListInterviewPrep)
	mux.HandleFunc("GET /api/interview-prep/{id}", a.handleGetInterviewPrep)

	mux.HandleFunc("GET /api/export", a.handleExport)
	mux.HandleFunc("GET /api/settings", a.handleGetSettings)
	mux.HandleFunc("PUT /api/settings", a.handlePutSettings)

	return a.cors(mux)
}

func (a *app) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	a.writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *app) handleMeta(w http.ResponseWriter, r *http.Request) {
	engine := "规则引擎"
	s := loadSettings(r.Context(), a.pool)
	if s.LLMEnabled && s.LLMAPIKey != "" {
		engine = "LLM"
	}
	a.writeJSON(w, http.StatusOK, map[string]string{"app": "个人简历动态更新系统", "engine": engine, "version": "1.0.0"})
}

const resumeColumns = `id, title, original_name, mime, source_text, content, created_at, updated_at`

func scanResume(row pgx.Row) (Resume, error) {
	var r Resume
	var content []byte
	var createdAt, updatedAt time.Time
	if err := row.Scan(&r.ID, &r.Title, &r.OriginalName, &r.Mime, &r.SourceText, &content, &createdAt, &updatedAt); err != nil {
		return r, err
	}
	_ = json.Unmarshal(content, &r.Content)
	r.CreatedAt = createdAt.Format(time.RFC3339Nano)
	r.UpdatedAt = updatedAt.Format(time.RFC3339Nano)
	return r, nil
}

func (a *app) handleListResumes(w http.ResponseWriter, r *http.Request) {
	rows, err := a.pool.Query(r.Context(), `SELECT `+resumeColumns+` FROM resumes ORDER BY updated_at DESC`)
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	defer rows.Close()
	out := []Resume{}
	for rows.Next() {
		var r Resume
		var content []byte
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&r.ID, &r.Title, &r.OriginalName, &r.Mime, &r.SourceText, &content, &createdAt, &updatedAt); err != nil {
			a.writeErr(w, 500, err.Error())
			return
		}
		_ = json.Unmarshal(content, &r.Content)
		r.CreatedAt = createdAt.Format(time.RFC3339Nano)
		r.UpdatedAt = updatedAt.Format(time.RFC3339Nano)
		out = append(out, r)
	}
	a.writeJSON(w, 200, out)
}

func (a *app) handleGetResume(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	res, err := scanResume(a.pool.QueryRow(r.Context(), `SELECT `+resumeColumns+` FROM resumes WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		a.writeErr(w, 404, "简历不存在")
		return
	}
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	a.writeJSON(w, 200, res)
}
