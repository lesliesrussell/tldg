// Package evidence defines the normalized evidence record, citation
// identifiers, and citation validation used across tldg (spec §8.3, §11.2).
package evidence

import (
	"fmt"
	"regexp"
	"time"
)

// tldg-5xh

// Kind enumerates evidence origins (spec §11.3). M0/M1 implements a subset.
type Kind string

const (
	KindLocalFile        Kind = "local_file"
	KindGitCommit        Kind = "git_commit"
	KindGitTag           Kind = "git_tag"
	KindGeneratedAnalysis Kind = "generated_analysis"
)

// TrustClass classifies evidence provenance (spec §13.4).
type TrustClass string

const (
	TrustPrimary     TrustClass = "primary"
	TrustOperational TrustClass = "operational"
	TrustSecondary   TrustClass = "secondary"
	TrustCommunity   TrustClass = "community"
	TrustUnverified  TrustClass = "unverified"
)

// Visibility marks whether evidence originates from public or local/private data.
type Visibility string

const (
	VisibilityLocal  Visibility = "local"
	VisibilityPublic Visibility = "public"
)

// Confidence labels an answer or claim (spec §4.2).
type Confidence string

const (
	ConfidenceVerified  Confidence = "verified"
	ConfidenceStrong    Confidence = "strong_inference"
	ConfidenceTentative Confidence = "tentative_inference"
	ConfidenceUnknown   Confidence = "unknown"
)

// Evidence is the normalized record for any extracted or retrieved item.
type Evidence struct {
	ID          string            `json:"id"`
	Kind        Kind              `json:"kind"`
	Source      string            `json:"source,omitempty"`
	URI         string            `json:"uri"`
	Ref         string            `json:"ref,omitempty"`
	Title       string            `json:"title,omitempty"`
	Content     string            `json:"-"`
	ContentHash string            `json:"content_hash,omitempty"`
	LineStart   *int              `json:"line_start,omitempty"`
	LineEnd     *int              `json:"line_end,omitempty"`
	RetrievedAt time.Time         `json:"retrieved_at"`
	TrustClass  TrustClass        `json:"trust_class,omitempty"`
	Visibility  Visibility        `json:"visibility,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	// Citation is the human-facing identifier, e.g. "local:README.md:10-44".
	Citation string `json:"citation,omitempty"`
}

// LocalFile constructs a local_file evidence record with a citation identifier.
// If lineStart/lineEnd are both > 0 they are included in the citation.
func LocalFile(id, relPath, ref string, lineStart, lineEnd int, now time.Time) Evidence {
	e := Evidence{
		ID:          id,
		Kind:        KindLocalFile,
		Source:      "local",
		URI:         "file:" + relPath,
		Ref:         ref,
		RetrievedAt: now,
		TrustClass:  TrustPrimary,
		Visibility:  VisibilityLocal,
	}
	if lineStart > 0 {
		ls := lineStart
		e.LineStart = &ls
	}
	if lineEnd > 0 {
		le := lineEnd
		e.LineEnd = &le
	}
	e.Citation = Citation("local", relPath, lineStart, lineEnd)
	return e
}

// Citation builds a human-facing citation identifier body (without brackets).
// Examples: "local:README.md:10-44", "local:go.mod", "git:abc1234".
func Citation(scheme, ref string, lineStart, lineEnd int) string {
	switch {
	case lineStart > 0 && lineEnd > 0 && lineEnd != lineStart:
		return fmt.Sprintf("%s:%s:%d-%d", scheme, ref, lineStart, lineEnd)
	case lineStart > 0:
		return fmt.Sprintf("%s:%s:%d", scheme, ref, lineStart)
	default:
		return fmt.Sprintf("%s:%s", scheme, ref)
	}
}

// citationRefRe extracts the body of bracketed citations like "[local:README.md:1-4]".
var citationRefRe = regexp.MustCompile(`\[([a-z]+:[^\]]+)\]`)

// ExtractCitations returns the unique citation bodies referenced in text.
func ExtractCitations(text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range citationRefRe.FindAllStringSubmatch(text, -1) {
		body := m[1]
		if !seen[body] {
			seen[body] = true
			out = append(out, body)
		}
	}
	return out
}

// Validate checks that every bracketed citation in text resolves to a known
// evidence citation. It returns the list of unknown (unresolvable) citations.
func Validate(text string, bundle []Evidence) []string {
	known := map[string]bool{}
	for _, e := range bundle {
		if e.Citation != "" {
			known[e.Citation] = true
		}
	}
	var unknown []string
	for _, c := range ExtractCitations(text) {
		if !known[c] {
			unknown = append(unknown, c)
		}
	}
	return unknown
}
