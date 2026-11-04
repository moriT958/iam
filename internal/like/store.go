// Package like implements the "like" button API: a per-post counter stored in SQLite.
package like

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // registers the "sqlite" driver (pure Go, no CGO)
)

// Store persists like counters keyed by post slug.
type Store interface {
	// Counts returns the counters of every post that has at least one like.
	Counts(ctx context.Context) (map[string]int64, error)
	// Count returns the counter of a post (0 when it has never been liked).
	Count(ctx context.Context, slug string) (int64, error)
	// Increment adds one like and returns the new counter.
	Increment(ctx context.Context, slug string) (int64, error)
	// Decrement removes one like (never below zero) and returns the new counter.
	Decrement(ctx context.Context, slug string) (int64, error)
}

// SQLiteStore is a Store backed by a SQLite database file.
type SQLiteStore struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS likes (
	slug  TEXT    PRIMARY KEY,
	count INTEGER NOT NULL DEFAULT 0 CHECK (count >= 0)
) STRICT;
`

// OpenSQLite opens (creating it if needed) the database at path.
func OpenSQLite(path string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	// WAL keeps readers from blocking the writer; busy_timeout absorbs short lock waits.
	dsn := "file:" + path +
		"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// A single process owns the file and the workload is tiny, so one connection
	// serialises writes and avoids SQLITE_BUSY altogether.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

// Close closes the database.
func (s *SQLiteStore) Close() error { return s.db.Close() }

func (s *SQLiteStore) Counts(ctx context.Context) (map[string]int64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT slug, count FROM likes`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[string]int64)
	for rows.Next() {
		var slug string
		var n int64
		if err := rows.Scan(&slug, &n); err != nil {
			return nil, err
		}
		counts[slug] = n
	}
	return counts, rows.Err()
}

func (s *SQLiteStore) Count(ctx context.Context, slug string) (int64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, `SELECT count FROM likes WHERE slug = ?`, slug).Scan(&n)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return n, err
}

func (s *SQLiteStore) Increment(ctx context.Context, slug string) (int64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO likes (slug, count) VALUES (?, 1)
		ON CONFLICT (slug) DO UPDATE SET count = count + 1
		RETURNING count`, slug).Scan(&n)
	return n, err
}

func (s *SQLiteStore) Decrement(ctx context.Context, slug string) (int64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, `
		UPDATE likes SET count = MAX(count - 1, 0)
		WHERE slug = ?
		RETURNING count`, slug).Scan(&n)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return n, err
}
