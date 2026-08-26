package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	"github.com/lesliesrussell/tldg/internal/config"
	"github.com/lesliesrussell/tldg/internal/docs"
	"github.com/lesliesrussell/tldg/internal/evidence"
	"github.com/lesliesrussell/tldg/internal/git"
	"github.com/lesliesrussell/tldg/internal/index"
	"github.com/lesliesrussell/tldg/internal/language"
	"github.com/lesliesrussell/tldg/internal/render"
	"github.com/lesliesrussell/tldg/internal/repo"
	"github.com/lesliesrussell/tldg/internal/scan"
	"github.com/lesliesrussell/tldg/internal/target"
	"github.com/lesliesrussell/tldg/internal/version"
)

// tldg-eca
// Milestone-1 analysis pipeline: scan → extract → detect → index → bundle.

// analyzeProfile runs file selection, documentation/language extraction, and
// (unless disabled) lexical indexing, populating prof and returning the
// grounding bundle for synthesis.
func analyzeProfile(ctx context.Context, cfg *config.Config, opts SummaryOptions, t *target.Target, prof *repo.Profile, res *render.Result) (*evidence.Bundle, error) {
	scanRes, err := selectFiles(ctx, cfg, t)
	if err != nil {
		return nil, err
	}

	// Language + dependency detection.
	langs, deps := language.Detect(t.Path, scanRes.Selected)
	prof.Languages = langs
	prof.Dependencies = deps

	// Documentation index + signals.
	prof.Docs = docs.Index(scanRes)
	prof.FileCount = scanRes.Total
	prof.SelectedFiles = len(scanRes.Selected)
	prof.HasTests = scanRes.Has(scan.CatTest)
	prof.HasCI = scanRes.Has(scan.CatCI)
	for _, f := range scanRes.ByCategory(scan.CatEntrypoint, scan.CatDocker, scan.CatCI) {
		prof.EntryPoints = append(prof.EntryPoints, repo.EntryPoint{
			Path: f.RelPath, Kind: string(f.Category), Inferred: f.Category != scan.CatEntrypoint,
		})
	}
	if len(scanRes.Skipped) > 0 {
		prof.Notes = append(prof.Notes, fmt.Sprintf("%d files skipped (binary/oversized/ignored)", len(scanRes.Skipped)))
	}

	// Build the grounding bundle from the highest-priority files.
	ref := t.Identity.HeadCommit
	if ref == "" {
		ref = "working-tree"
	}
	bundle := evidence.BuildBundle(scanRes.Selected, evidence.BundleOptions{
		MaxChunks:      cfg.Retrieval.MaxChunks,
		MaxChunkTokens: cfg.Retrieval.MaxChunkTokens,
		Ref:            ref,
		Now:            time.Now(),
	})
	prof.Notes = append(prof.Notes, bundle.Notes...)

	// Persist the lexical FTS index unless suppressed.
	if !opts.NoIndex {
		if err := buildIndex(cfg, t, scanRes, bundle); err != nil {
			// Index failure is non-fatal for a summary (spec §20).
			res.Warnings = append(res.Warnings, "index build failed: "+err.Error())
		}
	}
	return bundle, nil
}

// selectFiles returns the scan result, preferring git ls-files (honors
// .gitignore) and falling back to a filesystem walk for non-Git directories.
func selectFiles(ctx context.Context, cfg *config.Config, t *target.Target) (*scan.Result, error) {
	opts := scan.Options{
		IgnoreNames:    cfg.Index.Ignore,
		MaxFileSizeKB:  cfg.Index.MaxFileSizeKB,
		FollowSymlinks: cfg.Index.FollowSymlinks,
	}
	if t.Identity.IsGitRepo {
		backend := git.NewCLI()
		rels, err := backend.ListFiles(ctx, t.Path, cfg.Index.IncludeUntracked)
		if err == nil {
			return scan.SelectFiles(t.Path, rels, opts), nil
		}
		// Fall through to filesystem walk on git listing error.
	}
	return scan.WalkDir(t.Path, opts)
}

// buildIndex persists chunks to a per-repository SQLite FTS index and writes a
// reproducibility manifest (spec §11.1, §11.7).
func buildIndex(cfg *config.Config, t *target.Target, scanRes *scan.Result, bundle *evidence.Bundle) error {
	repoID := repositoryID(t)
	dir := filepath.Join(cfg.Index.Root, repoID)
	if err := ensureWritable(dir); err != nil {
		return err
	}
	store, err := index.Open(filepath.Join(dir, "lexical.db"))
	if err != nil {
		return err
	}
	defer store.Close()
	if err := store.InitSchema(); err != nil {
		return err
	}
	if err := store.Reset(); err != nil {
		return err
	}
	for _, s := range bundle.Snippets {
		ls, le := 0, 0
		if s.Evidence.LineStart != nil {
			ls = *s.Evidence.LineStart
		}
		if s.Evidence.LineEnd != nil {
			le = *s.Evidence.LineEnd
		}
		if err := store.AddChunk(index.Chunk{
			RelPath:     relFromURI(s.Evidence.URI),
			LineStart:   ls,
			LineEnd:     le,
			Category:    s.Evidence.Title,
			Content:     s.Content,
			ContentHash: s.Evidence.ContentHash,
		}); err != nil {
			return err
		}
	}
	count, _ := store.Count()
	state := "none"
	if t.Identity.IsGitRepo {
		if t.Identity.Dirty {
			state = "dirty"
		} else {
			state = "clean"
		}
	}
	m := &index.Manifest{
		RepositoryID:     repoID,
		Revision:         t.Identity.HeadCommit,
		WorkingTreeState: state,
		TLDGVersion:      version.Version,
		ChunkConfigHash:  index.HashConfig(fmt.Sprintf("%d", cfg.Retrieval.MaxChunkTokens)),
		ChunkCount:       count,
		Timestamp:        time.Now(),
	}
	return m.Write(dir)
}

// repositoryID derives a stable index directory name for a target.
func repositoryID(t *target.Target) string {
	key := t.Path
	if len(t.Identity.Remotes) > 0 {
		key = t.Identity.Remotes[0].NormalizedURL
	}
	sum := sha256.Sum256([]byte(key))
	return filepath.Base(t.Path) + "-" + hex.EncodeToString(sum[:])[:12]
}

// relFromURI strips the "file:" scheme from a local evidence URI.
func relFromURI(uri string) string {
	const p = "file:"
	if len(uri) > len(p) && uri[:len(p)] == p {
		return uri[len(p):]
	}
	return uri
}
