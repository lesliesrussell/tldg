package index

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// tldg-eca

// Chunk is a stored unit of indexed content with source location.
type Chunk struct {
	RelPath     string
	LineStart   int
	LineEnd     int
	Category    string
	Content     string
	ContentHash string
}

// Hit is a lexical search result.
type Hit struct {
	RelPath   string
	LineStart int
	LineEnd   int
	Category  string
	Snippet   string
}

// InitSchema creates the metadata and FTS5 tables (idempotent).
func (s *Store) InitSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS chunks (
			id INTEGER PRIMARY KEY,
			rel_path TEXT NOT NULL,
			line_start INTEGER,
			line_end INTEGER,
			category TEXT,
			content_hash TEXT
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
			content,
			rel_path UNINDEXED,
			tokenize='porter'
		)`,
	}
	for _, st := range stmts {
		if _, err := s.db.Exec(st); err != nil {
			return fmt.Errorf("init schema: %w", err)
		}
	}
	return nil
}

// Reset drops and recreates the index tables for a fresh build. Dropping (vs
// DELETE) also upgrades any table left by an older schema version.
func (s *Store) Reset() error {
	for _, st := range []string{`DROP TABLE IF EXISTS chunks_fts`, `DROP TABLE IF EXISTS chunks`} {
		if _, err := s.db.Exec(st); err != nil {
			return fmt.Errorf("reset drop: %w", err)
		}
	}
	return s.InitSchema()
}

// AddChunk indexes a single chunk, returning its content hash.
func (s *Store) AddChunk(c Chunk) error {
	if c.ContentHash == "" {
		sum := sha256.Sum256([]byte(c.Content))
		c.ContentHash = "sha256:" + hex.EncodeToString(sum[:])
	}
	res, err := s.db.Exec(
		`INSERT INTO chunks(rel_path, line_start, line_end, category, content_hash)
		 VALUES(?,?,?,?,?)`,
		c.RelPath, c.LineStart, c.LineEnd, c.Category, c.ContentHash,
	)
	if err != nil {
		return fmt.Errorf("insert chunk: %w", err)
	}
	id, _ := res.LastInsertId()
	if _, err := s.db.Exec(
		`INSERT INTO chunks_fts(rowid, content, rel_path) VALUES(?,?,?)`,
		id, c.Content, c.RelPath,
	); err != nil {
		return fmt.Errorf("insert fts: %w", err)
	}
	return nil
}

// Count returns the number of indexed chunks.
func (s *Store) Count() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM chunks`).Scan(&n)
	return n, err
}

// Search runs a lexical FTS5 query and returns up to limit hits.
func (s *Store) Search(query string, limit int) ([]Hit, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Query(
		`SELECT c.rel_path, c.line_start, c.line_end, c.category,
		        snippet(chunks_fts, 0, '', '', ' … ', 12)
		 FROM chunks_fts
		 JOIN chunks c ON c.id = chunks_fts.rowid
		 WHERE chunks_fts MATCH ?
		 ORDER BY rank
		 LIMIT ?`,
		sanitizeQuery(query), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("fts query: %w", err)
	}
	defer rows.Close()
	var hits []Hit
	for rows.Next() {
		var h Hit
		if err := rows.Scan(&h.RelPath, &h.LineStart, &h.LineEnd, &h.Category, &h.Snippet); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	return hits, rows.Err()
}

// sanitizeQuery makes a user string safe for FTS5 MATCH by quoting terms.
func sanitizeQuery(q string) string {
	fields := strings.Fields(q)
	quoted := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.ReplaceAll(f, `"`, "")
		if f != "" {
			quoted = append(quoted, `"`+f+`"`)
		}
	}
	return strings.Join(quoted, " OR ")
}
