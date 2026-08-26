// Package repo defines the structured repository profile assembled by the
// analysis pipeline (spec §9, §7.3).
package repo

import (
	"time"

	"github.com/leslierussell/tldg/internal/git"
)

// tldg-5xh

// Language is a detected language/ecosystem with a rough weight.
type Language struct {
	Name      string `json:"name"`
	Files     int    `json:"files"`
	Manifests []string `json:"manifests,omitempty"`
	Primary   bool   `json:"primary"`
}

// Dependency is a direct manifest dependency (spec §9.8, manifest-level).
type Dependency struct {
	Name     string `json:"name"`
	Version  string `json:"version,omitempty"`
	Ecosystem string `json:"ecosystem"`
	Manifest string `json:"manifest"`
}

// DocFile is an extracted, classified documentation/manifest file.
type DocFile struct {
	Path     string `json:"path"`
	Category string `json:"category"` // readme, license, changelog, manifest, ci, docs, adr, ...
	Bytes    int    `json:"bytes"`
}

// EntryPoint is a candidate execution entry point (spec §9.7).
type EntryPoint struct {
	Path       string `json:"path"`
	Kind       string `json:"kind"`       // main, script, dockerfile, ci, ...
	Detail     string `json:"detail,omitempty"`
	Inferred   bool   `json:"inferred"`   // framework-convention vs static certainty
}

// Profile is the structured repository model (spec §7.3 summary sections).
type Profile struct {
	Path         string        `json:"path"`
	Identity     git.Identity  `json:"identity"`
	Reduced      bool          `json:"reduced"` // non-Git directory
	Languages    []Language    `json:"languages"`
	Dependencies []Dependency  `json:"dependencies"`
	Docs         []DocFile     `json:"docs"`
	EntryPoints  []EntryPoint  `json:"entry_points"`
	FileCount    int           `json:"file_count"`
	SelectedFiles int          `json:"selected_files"`
	HasTests     bool          `json:"has_tests"`
	HasCI        bool          `json:"has_ci"`
	GeneratedAt  time.Time     `json:"generated_at"`
	// Notes carries non-fatal observations (e.g. skipped large files).
	Notes []string `json:"notes,omitempty"`
}

// PrimaryLanguage returns the primary language name, or "" if unknown.
func (p *Profile) PrimaryLanguage() string {
	for _, l := range p.Languages {
		if l.Primary {
			return l.Name
		}
	}
	if len(p.Languages) > 0 {
		return p.Languages[0].Name
	}
	return ""
}
