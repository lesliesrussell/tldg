package app

import (
	"context"

	"github.com/leslierussell/tldg/internal/config"
	"github.com/leslierussell/tldg/internal/evidence"
	"github.com/leslierussell/tldg/internal/render"
	"github.com/leslierussell/tldg/internal/repo"
	"github.com/leslierussell/tldg/internal/target"
)

// tldg-5xh
// Milestone-0 pipeline hooks. These are extended into the full scan → extract →
// index → synthesize pipeline in milestone 1.

// analyzeProfile populates the structured profile beyond Git identity. In
// milestone 0 it is a no-op; milestone 1 adds scan, docs, and language.
func analyzeProfile(ctx context.Context, cfg *config.Config, opts SummaryOptions, t *target.Target, prof *repo.Profile, res *render.Result) error {
	return nil
}

// synthesizeSummary produces the LLM-backed answer. In milestone 0 it renders
// the structured profile and notes that model synthesis lands in milestone 1.
func synthesizeSummary(ctx context.Context, cfg *config.Config, opts SummaryOptions, prof *repo.Profile, res *render.Result) error {
	res.Answer = render.Answer{
		Text:       profileText(prof),
		Confidence: evidence.ConfidenceVerified,
	}
	res.Warnings = append(res.Warnings, "model synthesis lands in milestone 1")
	return nil
}
