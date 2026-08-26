// Package index provides SQLite-backed metadata and a lexical FTS5 index
// (spec §11.1). It uses the pure-Go modernc.org/sqlite driver so tldg ships as
// a single static binary. The FTS layer is added in milestone 1; milestone 0
// uses Open/IntegrityCheck for the doctor probe.
package index

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// tldg-5xh

// Store wraps a SQLite database connection.
type Store struct {
	db   *sql.DB
	path string
}

// Open opens (creating if needed) a SQLite database at path. Use ":memory:"
// for an ephemeral probe.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite %s: %w", path, err)
	}
	return &Store{db: db, path: path}, nil
}

// IntegrityCheck runs SQLite's integrity_check pragma.
func (s *Store) IntegrityCheck() error {
	var result string
	if err := s.db.QueryRow("PRAGMA integrity_check").Scan(&result); err != nil {
		return fmt.Errorf("integrity_check: %w", err)
	}
	if result != "ok" {
		return fmt.Errorf("integrity_check reported: %s", result)
	}
	return nil
}

// DB exposes the underlying handle for the FTS layer (milestone 1).
func (s *Store) DB() *sql.DB { return s.db }

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }
