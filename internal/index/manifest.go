package index

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// tldg-eca

// Manifest records index reproducibility metadata (spec §11.7).
type Manifest struct {
	RepositoryID   string    `json:"repository_id"`
	Revision       string    `json:"revision"`
	WorkingTreeState string  `json:"working_tree_state"` // clean|dirty|none
	TLDGVersion    string    `json:"tldg_version"`
	EmbeddingModel string    `json:"embedding_model,omitempty"`
	ChunkConfigHash string   `json:"chunk_config_hash"`
	ChunkCount     int       `json:"chunk_count"`
	Timestamp      time.Time `json:"timestamp"`
}

// Write persists the manifest as manifest.json under dir.
func (m *Manifest) Write(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), b, 0o644)
}

// HashConfig produces a stable hash of chunking configuration inputs.
func HashConfig(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}
