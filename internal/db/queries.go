package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type SessionStatus string

const (
	StatusQueued    SessionStatus = "queued"
	StatusRunning   SessionStatus = "running"
	StatusCompleted SessionStatus = "completed"
	StatusFailed    SessionStatus = "failed"
	StatusCancelled SessionStatus = "cancelled"
)

type SessionRow struct {
	ID           string        `json:"id"`
	Workspace    string        `json:"workspace"`
	App          string        `json:"app"`
	Task         string        `json:"task"`
	Cwd          string        `json:"cwd"`
	PID          *int64        `json:"pid"`
	Status       SessionStatus `json:"status"`
	CreatedAt    int64         `json:"created_at"`
	StartedAt    *int64        `json:"started_at"`
	EndedAt      *int64        `json:"ended_at"`
	ExitCode     *int64        `json:"exit_code"`
	JsonlPath    *string       `json:"jsonl_path"`
	StderrPath   *string       `json:"stderr_path"`
	StdoutPath   *string       `json:"stdout_path"`
	ErrorMessage *string       `json:"error_message"`
	MetadataJSON *string       `json:"metadata_json"`
}

type SessionSummary struct {
	ID        string        `json:"id"`
	Workspace string        `json:"workspace"`
	App       string        `json:"app"`
	Task      string        `json:"task"`
	Status    SessionStatus `json:"status"`
	CreatedAt int64         `json:"created_at"`
	StartedAt *int64        `json:"started_at"`
	EndedAt   *int64        `json:"ended_at"`
}

type InsertSessionInput struct {
	ID         string
	Workspace  string
	App        string
	Task       string
	Cwd        string
	Status     SessionStatus
	JsonlPath  string
	StdoutPath string
	StderrPath string
	CreatedAt  int64
}

func InsertSession(db *sql.DB, s InsertSessionInput) error {
	_, err := db.Exec(
		`INSERT INTO sessions
		 (id, workspace, app, task, cwd, status, jsonl_path, stdout_path, stderr_path, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Workspace, s.App, s.Task, s.Cwd, s.Status,
		s.JsonlPath, s.StdoutPath, s.StderrPath, s.CreatedAt,
	)
	return err
}

func GetSession(db *sql.DB, id string) (*SessionRow, error) {
	row := db.QueryRow(`SELECT id, workspace, app, task, cwd, pid, status, created_at,
		started_at, ended_at, exit_code, jsonl_path, stderr_path, stdout_path, error_message, metadata_json
		FROM sessions WHERE id = ?`, id)
	return scanSession(row)
}

type scanner interface{ Scan(dest ...any) error }

func scanSession(s scanner) (*SessionRow, error) {
	var r SessionRow
	if err := s.Scan(
		&r.ID, &r.Workspace, &r.App, &r.Task, &r.Cwd,
		&r.PID, &r.Status, &r.CreatedAt,
		&r.StartedAt, &r.EndedAt, &r.ExitCode,
		&r.JsonlPath, &r.StderrPath, &r.StdoutPath,
		&r.ErrorMessage, &r.MetadataJSON,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

type ListOpts struct {
	Workspace string
	Status    SessionStatus
	Limit     int
}

func ListSessions(db *sql.DB, opts ListOpts) ([]SessionSummary, error) {
	where := []string{}
	args := []any{}
	if opts.Workspace != "" {
		where = append(where, "workspace = ?")
		args = append(args, opts.Workspace)
	}
	if opts.Status != "" {
		where = append(where, "status = ?")
		args = append(args, string(opts.Status))
	}
	if opts.Limit <= 0 {
		opts.Limit = 50
	}
	args = append(args, opts.Limit)

	q := `SELECT id, workspace, app, task, status, created_at, started_at, ended_at
	      FROM sessions`
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY created_at DESC LIMIT ?"

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SessionSummary
	for rows.Next() {
		var s SessionSummary
		if err := rows.Scan(&s.ID, &s.Workspace, &s.App, &s.Task, &s.Status,
			&s.CreatedAt, &s.StartedAt, &s.EndedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func UpdateSessionSpawned(db *sql.DB, id string, pid int64, startedAt int64) error {
	_, err := db.Exec(
		`UPDATE sessions SET pid = ?, started_at = ?, status = 'running' WHERE id = ?`,
		pid, startedAt, id,
	)
	return err
}

func UpdateSessionExited(db *sql.DB, id string, status SessionStatus,
	exitCode *int64, endedAt int64, errMsg *string) error {
	_, err := db.Exec(
		`UPDATE sessions SET status = ?, exit_code = ?, ended_at = ?, error_message = ? WHERE id = ?`,
		string(status), exitCode, endedAt, errMsg, id,
	)
	return err
}

func UpdateSessionFailed(db *sql.DB, id string, errMsg string, endedAt int64) error {
	_, err := db.Exec(
		`UPDATE sessions SET status = 'failed', error_message = ?, ended_at = ? WHERE id = ?`,
		errMsg, endedAt, id,
	)
	return err
}

func SelectRunning(db *sql.DB) ([]SessionRow, error) {
	rows, err := db.Query(`SELECT id, workspace, app, task, cwd, pid, status, created_at,
		started_at, ended_at, exit_code, jsonl_path, stderr_path, stdout_path, error_message, metadata_json
		FROM sessions WHERE status = 'running'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SessionRow
	for rows.Next() {
		r, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		if r != nil {
			out = append(out, *r)
		}
	}
	return out, rows.Err()
}

func InsertEvent(db *sql.DB, sessionID, kind string, payload any) error {
	var payloadJSON *string
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal event payload: %w", err)
		}
		s := string(b)
		payloadJSON = &s
	}
	_, err := db.Exec(
		`INSERT INTO events (session_id, ts, kind, payload_json) VALUES (?, ?, ?, ?)`,
		sessionID, time.Now().UnixMilli(), kind, payloadJSON,
	)
	return err
}
