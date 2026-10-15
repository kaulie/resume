package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

const jdColumns = `id, title, company, description, responsibilities, requirements, plus_points, created_at, updated_at`

func scanJD(row pgx.Row) (JD, error) {
	var j JD
	var createdAt, updatedAt time.Time
	if err := row.Scan(&j.ID, &j.Title, &j.Company, &j.Description, &j.Responsibilities, &j.Requirements, &j.PlusPoints, &createdAt, &updatedAt); err != nil {
		return j, err
	}
	j.CreatedAt = createdAt.Format(time.RFC3339Nano)
	j.UpdatedAt = updatedAt.Format(time.RFC3339Nano)
	return j, nil
}

func (a *app) handleListJDs(w http.ResponseWriter, r *http.Request) {
	rows, err := a.pool.Query(r.Context(), `SELECT `+jdColumns+` FROM jds ORDER BY updated_at DESC`)
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	defer rows.Close()
	out := []JD{}
	for rows.Next() {
		var j JD
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&j.ID, &j.Title, &j.Company, &j.Description, &j.Responsibilities, &j.Requirements, &j.PlusPoints, &createdAt, &updatedAt); err != nil {
			a.writeErr(w, 500, err.Error())
			return
		}
		j.CreatedAt = createdAt.Format(time.RFC3339Nano)
		j.UpdatedAt = updatedAt.Format(time.RFC3339Nano)
		out = append(out, j)
	}
	a.writeJSON(w, 200, out)
}

func (a *app) handleGetJD(w http.ResponseWriter, r *http.Request) {
	j, err := scanJD(a.pool.QueryRow(r.Context(), `SELECT `+jdColumns+` FROM jds WHERE id = $1`, r.PathValue("id")))
	if errors.Is(err, pgx.ErrNoRows) {
		a.writeErr(w, 404, "JD 不存在")
		return
	}
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	a.writeJSON(w, 200, j)
}

func (a *app) handleCreateJD(w http.ResponseWriter, r *http.Request) {
	var b JD
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		a.writeErr(w, 400, err.Error())
		return
	}
	if b.Title == "" {
		a.writeErr(w, 400, "职位标题不能为空")
		return
	}
	b.ID = newID()
	now := nowISO()
	_, err := a.pool.Exec(r.Context(), `INSERT INTO jds (id, title, company, description, responsibilities, requirements, plus_points, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)`, b.ID, b.Title, b.Company, b.Description, b.Responsibilities, b.Requirements, b.PlusPoints, now)
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	j, _ := scanJD(a.pool.QueryRow(r.Context(), `SELECT `+jdColumns+` FROM jds WHERE id = $1`, b.ID))
	a.writeJSON(w, 201, j)
}

func (a *app) handleUpdateJD(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var exists int
	err := a.pool.QueryRow(r.Context(), `SELECT 1 FROM jds WHERE id = $1`, id).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		a.writeErr(w, 404, "JD 不存在")
		return
	}
	var b JD
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		a.writeErr(w, 400, err.Error())
		return
	}
	if b.Title == "" {
		a.writeErr(w, 400, "职位标题不能为空")
		return
	}
	_, err = a.pool.Exec(r.Context(), `UPDATE jds SET title=$1, company=$2, description=$3, responsibilities=$4, requirements=$5, plus_points=$6, updated_at=$7 WHERE id=$8`,
		b.Title, b.Company, b.Description, b.Responsibilities, b.Requirements, b.PlusPoints, nowISO(), id)
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	j, _ := scanJD(a.pool.QueryRow(r.Context(), `SELECT `+jdColumns+` FROM jds WHERE id = $1`, id))
	a.writeJSON(w, 200, j)
}

func (a *app) handleDeleteJD(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, _ = a.pool.Exec(r.Context(), `DELETE FROM resume_versions WHERE jd_id = $1`, id)
	_, _ = a.pool.Exec(r.Context(), `DELETE FROM evaluations WHERE jd_id = $1`, id)
	_, _ = a.pool.Exec(r.Context(), `DELETE FROM interview_preps WHERE jd_id = $1`, id)
	_, _ = a.pool.Exec(r.Context(), `DELETE FROM jds WHERE id = $1`, id)
	a.writeJSON(w, 200, map[string]bool{"ok": true})
}
