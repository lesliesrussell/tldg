// Package render produces tldg output: terminal text, Markdown, and the stable
// versioned JSON contract (spec §8).
package render

import (
	"github.com/lesliesrussell/tldg/internal/evidence"
	"github.com/lesliesrussell/tldg/internal/repo"
)

// tldg-5xh

// SchemaVersion is the JSON output contract version (spec §8.2).
const SchemaVersion = "1.0"

// Repository identifies the analyzed repository in output.
type Repository struct {
	Host  string `json:"host,omitempty"`
	Owner string `json:"owner,omitempty"`
	Name  string `json:"name,omitempty"`
	Ref   string `json:"ref,omitempty"`
}

// TargetInfo describes the resolved target in output.
type TargetInfo struct {
	Kind       string      `json:"kind"`
	Path       string      `json:"path,omitempty"`
	Repository *Repository `json:"repository,omitempty"`
}

// Answer is the top-level synthesized answer/summary.
type Answer struct {
	Text                 string                `json:"text"`
	Confidence           evidence.Confidence   `json:"confidence"`
	ExternalResearchUsed bool                  `json:"external_research_used"`
}

// Claim is a single supported statement with its evidence references.
type Claim struct {
	Text        string              `json:"text"`
	Confidence  evidence.Confidence `json:"confidence"`
	EvidenceIDs []string            `json:"evidence_ids"`
}

// Timing reports per-phase durations in milliseconds (spec §8.2).
type Timing struct {
	IndexMS     int64 `json:"index_ms"`
	RetrievalMS int64 `json:"retrieval_ms"`
	ModelMS     int64 `json:"model_ms"`
}

// Result is the full output envelope (spec §8.2).
type Result struct {
	SchemaVersion string              `json:"schema_version"`
	Command       string              `json:"command"`
	Target        TargetInfo          `json:"target"`
	Answer        Answer              `json:"answer"`
	Claims        []Claim             `json:"claims,omitempty"`
	Evidence      []evidence.Evidence `json:"evidence,omitempty"`
	Warnings      []string            `json:"warnings"`
	Timing        Timing              `json:"timing"`
	// Profile is attached to summary results for structured consumers.
	Profile *repo.Profile `json:"profile,omitempty"`
}

// NewResult constructs a Result with the schema version set and non-nil slices.
func NewResult(command string) *Result {
	return &Result{
		SchemaVersion: SchemaVersion,
		Command:       command,
		Warnings:      []string{},
	}
}
