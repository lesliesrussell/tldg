package app

import (
	"context"
	"fmt"
	"time"

	"github.com/leslierussell/tldg/internal/config"
	"github.com/leslierussell/tldg/internal/evidence"
	"github.com/leslierussell/tldg/internal/git"
	"github.com/leslierussell/tldg/internal/render"
	"github.com/leslierussell/tldg/internal/repo"
	"github.com/leslierussell/tldg/internal/target"
)

// tldg-5xh

// SummaryOptions controls the summary command.
type SummaryOptions struct {
	Target       string
	Offline      bool
	NoIndex      bool
	Depth        string // brief|standard|architecture|exhaustive
	ModelOverride string
}

// RunSummary resolves the target and assembles a repository profile. In
// milestone 0 this produces repository identity and a structured profile; the
// scan/extraction/synthesis pipeline is layered on in milestone 1.
func RunSummary(ctx context.Context, cfg *config.Config, opts SummaryOptions) (*render.Result, error) {
	backend := git.NewCLI()
	t, err := target.Resolve(ctx, backend, opts.Target)
	if err != nil {
		return nil, err
	}

	res := render.NewResult("summary")
	res.Target = targetInfo(t)

	prof := &repo.Profile{
		Path:        t.Path,
		Identity:    t.Identity,
		Reduced:     t.Reduced,
		GeneratedAt: time.Now(),
	}

	// Build the pipeline profile (scan/docs/language) — implemented in M1.
	if err := analyzeProfile(ctx, cfg, opts, t, prof, res); err != nil {
		return nil, err
	}
	res.Profile = prof

	// Offline (or no model available): render the structured profile without
	// LLM synthesis (spec §20 degraded operation).
	if opts.Offline {
		res.Answer = render.Answer{
			Text:       profileText(prof),
			Confidence: evidence.ConfidenceVerified,
		}
		res.Warnings = append(res.Warnings, "model synthesis skipped (offline)")
		return res, nil
	}

	// Online synthesis — implemented in M1.
	if err := synthesizeSummary(ctx, cfg, opts, prof, res); err != nil {
		return nil, err
	}
	return res, nil
}

// targetInfo maps a resolved target to render.TargetInfo.
func targetInfo(t *target.Target) render.TargetInfo {
	ti := render.TargetInfo{Kind: string(t.Kind), Path: t.Path}
	if t.Identity.IsGitRepo {
		repoInfo := &render.Repository{Ref: t.Identity.HeadCommit}
		if t.Identity.Branch != "" {
			repoInfo.Ref = t.Identity.Branch
		}
		if len(t.Identity.Remotes) > 0 {
			repoInfo.Name = t.Identity.Remotes[0].NormalizedURL
		}
		ti.Repository = repoInfo
	}
	return ti
}

// profileText renders a compact textual summary of the structured profile,
// used for offline output.
func profileText(p *repo.Profile) string {
	var s string
	s += "Repository profile\n\n"
	if p.Reduced {
		s += fmt.Sprintf("Path: %s (not a Git worktree — reduced mode)\n", p.Path)
	} else {
		s += fmt.Sprintf("Path: %s\n", p.Path)
		if p.Identity.Branch != "" {
			s += fmt.Sprintf("Branch: %s\n", p.Identity.Branch)
		}
		if p.Identity.HeadCommit != "" {
			s += fmt.Sprintf("HEAD: %s\n", short(p.Identity.HeadCommit))
		}
		if len(p.Identity.Remotes) > 0 {
			s += fmt.Sprintf("Remote: %s\n", p.Identity.Remotes[0].NormalizedURL)
		}
		if p.Identity.Dirty {
			s += "Working tree: dirty\n"
		}
	}
	if lang := p.PrimaryLanguage(); lang != "" {
		s += fmt.Sprintf("Primary language: %s\n", lang)
	}
	if p.SelectedFiles > 0 {
		s += fmt.Sprintf("Selected files: %d of %d\n", p.SelectedFiles, p.FileCount)
	}
	if len(p.Dependencies) > 0 {
		s += fmt.Sprintf("Direct dependencies: %d\n", len(p.Dependencies))
	}
	if p.HasTests {
		s += "Tests: present\n"
	}
	if p.HasCI {
		s += "CI: present\n"
	}
	return s
}

func short(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}
