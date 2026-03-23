package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	DB *sql.DB
}

func NewSQLiteStore(dataDir string) (*SQLiteStore, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	dbPath := filepath.Join(dataDir, "app.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	store := &SQLiteStore{DB: db}
	if err := store.migrate(context.Background()); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error { return s.DB.Close() }

func (s *SQLiteStore) migrate(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT NOT NULL, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS lights_cache (id TEXT PRIMARY KEY, name TEXT NOT NULL, room_id TEXT, room_name TEXT, on_state INTEGER NOT NULL, brightness INTEGER NOT NULL, color_hex TEXT NOT NULL, reachable INTEGER NOT NULL, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS rooms (id TEXT PRIMARY KEY, name TEXT NOT NULL, type TEXT NOT NULL, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS scenes (id TEXT PRIMARY KEY, name TEXT NOT NULL, group_id TEXT, group_name TEXT, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS rules (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, trigger_type TEXT NOT NULL, enabled INTEGER NOT NULL DEFAULT 1, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS rule_actions (id INTEGER PRIMARY KEY AUTOINCREMENT, rule_id INTEGER NOT NULL, action_type TEXT NOT NULL, target TEXT NOT NULL, value TEXT NOT NULL, FOREIGN KEY(rule_id) REFERENCES rules(id) ON DELETE CASCADE);`,
		`CREATE TABLE IF NOT EXISTS rule_conditions (id INTEGER PRIMARY KEY AUTOINCREMENT, rule_id INTEGER NOT NULL, condition_type TEXT NOT NULL, key_name TEXT NOT NULL, value TEXT NOT NULL, negate INTEGER NOT NULL DEFAULT 0, FOREIGN KEY(rule_id) REFERENCES rules(id) ON DELETE CASCADE);`,
		`CREATE TABLE IF NOT EXISTS devices (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, ip TEXT NOT NULL, present INTEGER NOT NULL DEFAULT 0, failure_count INTEGER NOT NULL DEFAULT 0, consecutive_oks INTEGER NOT NULL DEFAULT 0, last_seen_at DATETIME, last_checked_at DATETIME);`,
		`CREATE TABLE IF NOT EXISTS logs (id INTEGER PRIMARY KEY AUTOINCREMENT, level TEXT NOT NULL, source TEXT NOT NULL, message TEXT NOT NULL, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
	}
	for _, query := range queries {
		if _, err := s.DB.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("migrate query failed: %w", err)
		}
	}
	return nil
}

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}
