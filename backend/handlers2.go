package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (a *app) handleCreateResume(w http.ResponseWriter, r *http.Request) {
	id := newID()
	now := nowISO()
	title := ""
	originalName := ""
	mime := ""
	sourceText := ""
	var content ResumeContent

	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			a.writeErr(w, 400, err.Error())
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			a.writeErr(w, 400, "缺少文件")
			return
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, 20<<20))
		if err != nil {
			a.writeErr(w, 500, err.Error())
			return
		}
		originalName = header.Filename
		mime = detectMime(originalName, header.Header.Get("Content-Type"))
		text, err := parseFileToText(data, mime)
		if err != nil {
			a.writeErr(w, 500, "文件解析失败："+err.Error())
			return
		}
		sourceText = text
		content = textToContent(text)
		title = content.Name
		if title == "" || title == "未命名简历" {
			title = strings.TrimSuffix(originalName, "."+extOf(originalName))
		}
	} else {
		var body struct {
			Title   string        `json:"title"`
			Content ResumeContent `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			a.writeErr(w, 400, err.Error())
			return
		}
		content = body.Content
		title = body.Title
		if title == "" {
			title = content.Name
		}
		if title == "" {
			title = "未命名简历"
		}
	}

	contentJSON, _ := json.Marshal(content)
	_, err := a.pool.Exec(r.Context(), `INSERT INTO resumes (id, title, original_name, mime, source_text, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`, id, title, originalName, mime, sourceText, contentJSON, now)
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	res, err := scanResume(a.pool.QueryRow(r.Context(), `SELECT `+resumeColumns+` FROM resumes WHERE id = $1`, id))
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	a.writeJSON(w, 201, res)
}

func extOf(name string) string {
	i := strings.LastIndex(name, ".")
	if i < 0 {
		return ""
	}
	return name[i+1:]
}

func (a *app) handleUpdateResume(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var existing struct {
		Title   string
		Content []byte
	}
	err := a.pool.QueryRow(r.Context(), `SELECT title, content FROM resumes WHERE id = $1`, id).Scan(&existing.Title, &existing.Content)
	if errors.Is(err, pgx.ErrNoRows) {
		a.writeErr(w, 404, "简历不存在")
		return
	}
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	var body struct {
		Title   *string        `json:"title"`
		Content *ResumeContent `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		a.writeErr(w, 400, err.Error())
		return
	}
	nextTitle := existing.Title
	if body.Title != nil {
		nextTitle = *body.Title
	}
	nextContent := existing.Content
	if body.Content != nil {
		nextContent, _ = json.Marshal(*body.Content)
	}
	_, err = a.pool.Exec(r.Context(), `UPDATE resumes SET title = $1, content = $2, updated_at = $3 WHERE id = $4`, nextTitle, nextContent, nowISO(), id)
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	res, err := scanResume(a.pool.QueryRow(r.Context(), `SELECT `+resumeColumns+` FROM resumes WHERE id = $1`, id))
	if err != nil {
		a.writeErr(w, 500, err.Error())
		return
	}
	a.writeJSON(w, 200, res)
}

func (a *app) handleDeleteResume(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, _ = a.pool.Exec(r.Context(), `DELETE FROM resume_versions WHERE resume_id = $1`, id)
	_, _ = a.pool.Exec(r.Context(), `DELETE FROM evaluations WHERE resume_id = $1`, id)
	_, _ = a.pool.Exec(r.Context(), `DELETE FROM resumes WHERE id = $1`, id)
	a.writeJSON(w, 200, map[string]bool{"ok": true})
}
