package history

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "modernc.org/sqlite"
)

type Record struct {
	ID          int64
	Command     string
	Cwd         string
	ProjectName string
	ExitCode    int
	Stdout      string
	Stderr      string
	StartedAt   string
	DurationMS  int64
	Shell       string
	OS          string
}

type Store struct {
	db *sql.DB
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".term", "term.db"), nil
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS command_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  command TEXT NOT NULL,
  cwd TEXT,
  project_name TEXT,
  exit_code INTEGER,
  stdout TEXT,
  stderr TEXT,
  started_at TEXT,
  duration_ms INTEGER,
  shell TEXT,
  os TEXT
)`)
	return err
}

func NewRecord(command string, exitCode int, stdout, stderr string, durationMS int64, cwd, project string) Record {
	return Record{
		Command:     command,
		Cwd:         cwd,
		ProjectName: project,
		ExitCode:    exitCode,
		Stdout:      stdout,
		Stderr:      stderr,
		StartedAt:   time.Now().UTC().Format(time.RFC3339),
		DurationMS:  durationMS,
		Shell:       shellName(),
		OS:          runtime.GOOS,
	}
}

func (s *Store) Insert(r Record) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO command_history
(command, cwd, project_name, exit_code, stdout, stderr, started_at, duration_ms, shell, os)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.Command, r.Cwd, r.ProjectName, r.ExitCode, r.Stdout, r.Stderr, r.StartedAt, r.DurationMS, r.Shell, r.OS)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) Search(query string, limit int) ([]Record, error) {
	like := "%" + query + "%"
	rows, err := s.db.Query(`SELECT id, command, cwd, project_name, exit_code, stdout, stderr, started_at, duration_ms, shell, os
FROM command_history
WHERE command LIKE ? OR cwd LIKE ? OR project_name LIKE ? OR stdout LIKE ? OR stderr LIKE ?
ORDER BY id DESC LIMIT ?`, like, like, like, like, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (s *Store) Get(id int64) (*Record, error) {
	row := s.db.QueryRow(`SELECT id, command, cwd, project_name, exit_code, stdout, stderr, started_at, duration_ms, shell, os
FROM command_history WHERE id = ?`, id)
	return scanRow(row)
}

func (s *Store) LatestFailed() (*Record, error) {
	row := s.db.QueryRow(`SELECT id, command, cwd, project_name, exit_code, stdout, stderr, started_at, duration_ms, shell, os
FROM command_history WHERE exit_code != 0 AND stderr != '' ORDER BY id DESC LIMIT 1`)
	return scanRow(row)
}

func (s *Store) Clear() error {
	_, err := s.db.Exec(`DELETE FROM command_history`)
	return err
}

func scanRows(rows *sql.Rows) ([]Record, error) {
	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.ID, &r.Command, &r.Cwd, &r.ProjectName, &r.ExitCode, &r.Stdout, &r.Stderr, &r.StartedAt, &r.DurationMS, &r.Shell, &r.OS); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRow(row rowScanner) (*Record, error) {
	var r Record
	err := row.Scan(&r.ID, &r.Command, &r.Cwd, &r.ProjectName, &r.ExitCode, &r.Stdout, &r.Stderr, &r.StartedAt, &r.DurationMS, &r.Shell, &r.OS)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &r, err
}

func shellName() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("ComSpec")
	}
	return os.Getenv("SHELL")
}
