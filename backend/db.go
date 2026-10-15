package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func connectDB(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://gaolei@localhost:5432/resume?sslmode=disable"
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}

const schema = `
CREATE TABLE IF NOT EXISTS resumes (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  original_name TEXT NOT NULL DEFAULT '',
  mime TEXT NOT NULL DEFAULT '',
  source_text TEXT NOT NULL DEFAULT '',
  content JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS jds (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  company TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  responsibilities TEXT NOT NULL DEFAULT '',
  requirements TEXT NOT NULL DEFAULT '',
  plus_points TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS resume_versions (
  id TEXT PRIMARY KEY,
  resume_id TEXT NOT NULL,
  jd_id TEXT,
  label TEXT NOT NULL,
  content JSONB NOT NULL DEFAULT '{}',
  diff JSONB NOT NULL DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS evaluations (
  id TEXT PRIMARY KEY,
  resume_id TEXT NOT NULL,
  jd_id TEXT NOT NULL,
  overall_score INT NOT NULL DEFAULT 0,
  dimensions JSONB NOT NULL DEFAULT '[]',
  suggestions JSONB NOT NULL DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS interview_preps (
  id TEXT PRIMARY KEY,
  jd_id TEXT NOT NULL,
  content JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
`

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schema)
	return err
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func defaultSettings() Settings {
	return Settings{
		LLMEnabled:  false,
		LLMBaseURL:  "https://api.deepseek.com",
		LLMModel:    "deepseek-chat",
		LLMAPIKey:   "",
		PDFFontPath: "",
	}
}

func loadSettings(ctx context.Context, pool *pgxpool.Pool) Settings {
	s := defaultSettings()
	rows, err := pool.Query(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return s
	}
	defer rows.Close()
	for rows.Next() {
		var k, v string
		if rows.Scan(&k, &v) == nil {
			switch k {
			case "llm_enabled":
				s.LLMEnabled = v == "true"
			case "llm_base_url":
				s.LLMBaseURL = v
			case "llm_model":
				s.LLMModel = v
			case "llm_api_key":
				s.LLMAPIKey = v
			case "pdf_font_path":
				s.PDFFontPath = v
			}
		}
	}
	return s
}

func saveSettings(ctx context.Context, pool *pgxpool.Pool, s Settings) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	upsert := func(k, v string) error {
		_, err := tx.Exec(ctx, `INSERT INTO settings (key, value) VALUES ($1, $2)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`, k, v)
		return err
	}
	enabled := "false"
	if s.LLMEnabled {
		enabled = "true"
	}
	if err := upsert("llm_enabled", enabled); err != nil {
		return err
	}
	if err := upsert("llm_base_url", s.LLMBaseURL); err != nil {
		return err
	}
	if err := upsert("llm_model", s.LLMModel); err != nil {
		return err
	}
	if err := upsert("llm_api_key", s.LLMAPIKey); err != nil {
		return err
	}
	if err := upsert("pdf_font_path", s.PDFFontPath); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
