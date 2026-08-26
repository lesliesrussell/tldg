// Package app orchestrates tldg commands over the internal subsystems.
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/leslierussell/tldg/internal/config"
	"github.com/leslierussell/tldg/internal/git"
	"github.com/leslierussell/tldg/internal/index"
	"github.com/leslierussell/tldg/internal/models"
)

// tldg-5xh

// CheckStatus is the outcome of a single doctor check.
type CheckStatus string

const (
	StatusOK   CheckStatus = "ok"
	StatusWarn CheckStatus = "warn"
	StatusFail CheckStatus = "fail"
	StatusNA   CheckStatus = "n/a"
)

// Check is a single diagnostic result.
type Check struct {
	Name   string      `json:"name"`
	Status CheckStatus `json:"status"`
	Detail string      `json:"detail"`
}

// DoctorReport aggregates doctor checks (spec §21.1).
type DoctorReport struct {
	Checks []Check `json:"checks"`
}

// HasFailure reports whether any hard check failed.
func (r *DoctorReport) HasFailure() bool {
	for _, c := range r.Checks {
		if c.Status == StatusFail {
			return true
		}
	}
	return false
}

// DoctorOptions controls doctor behavior.
type DoctorOptions struct {
	Offline bool
}

// RunDoctor performs environment diagnostics (spec §21.1 subset for M0/M1).
func RunDoctor(ctx context.Context, cfg *config.Config, opts DoctorOptions) *DoctorReport {
	r := &DoctorReport{}

	// Git.
	backend := git.NewCLI()
	if ver, err := backend.Available(ctx); err == nil {
		r.add("git", StatusOK, ver)
	} else {
		r.add("git", StatusFail, err.Error())
	}

	// Config validity + active paths.
	probs := config.Validate(cfg)
	if len(probs) == 0 {
		p := cfg.Paths()
		r.add("config", StatusOK, fmt.Sprintf("valid (config=%s)", p.Config))
	} else {
		msgs := make([]string, 0, len(probs))
		for _, pr := range probs {
			msgs = append(msgs, pr.String())
		}
		r.add("config", StatusFail, strings.Join(msgs, "; "))
	}

	// Writable data/cache/index directories.
	p := cfg.Paths()
	for name, dir := range map[string]string{
		"data-dir":  p.Data,
		"cache-dir": p.Cache,
		"index-dir": cfg.Index.Root,
	} {
		if err := ensureWritable(dir); err != nil {
			r.add(name, StatusFail, err.Error())
		} else {
			r.add(name, StatusOK, dir)
		}
	}

	// SQLite integrity (in-memory probe).
	if st, err := index.Open(":memory:"); err != nil {
		r.add("sqlite", StatusFail, err.Error())
	} else {
		if err := st.IntegrityCheck(); err != nil {
			r.add("sqlite", StatusFail, err.Error())
		} else {
			r.add("sqlite", StatusOK, "integrity ok (modernc.org/sqlite)")
		}
		st.Close()
	}

	// Model + embedding reachability (skipped when offline).
	if opts.Offline {
		r.add("model", StatusNA, "skipped (--offline)")
		r.add("embeddings", StatusNA, "skipped (--offline)")
	} else {
		checkModel(ctx, r, cfg)
		checkEmbeddings(ctx, r, cfg)
	}

	// Deferred subsystems.
	r.add("keychain", StatusNA, "credential storage arrives in milestone 4")
	r.add("hosts", StatusNA, "host adapters arrive in milestone 4")

	return r
}

func checkModel(ctx context.Context, r *DoctorReport, cfg *config.Config) {
	mc, name, err := cfg.ActiveModel("")
	if err != nil {
		r.add("model", StatusFail, err.Error())
		return
	}
	prov, err := models.New(mc)
	if err != nil {
		r.add("model", StatusFail, err.Error())
		return
	}
	avail, err := prov.Ping(ctx)
	if err != nil {
		r.add("model", StatusFail, fmt.Sprintf("%s: %v", name, err))
		return
	}
	if models.HasModel(avail, mc.Model) {
		r.add("model", StatusOK, fmt.Sprintf("%s reachable, model %q present", prov.Name(), mc.Model))
	} else {
		r.add("model", StatusWarn, fmt.Sprintf("%s reachable but model %q not found (have %d models)", prov.Name(), mc.Model, len(avail)))
	}
}

func checkEmbeddings(ctx context.Context, r *DoctorReport, cfg *config.Config) {
	mc, ok := cfg.DefaultEmbedding()
	if !ok {
		r.add("embeddings", StatusNA, "no embedding model configured")
		return
	}
	prov, err := models.New(mc)
	if err != nil {
		r.add("embeddings", StatusWarn, err.Error())
		return
	}
	avail, err := prov.Ping(ctx)
	if err != nil {
		r.add("embeddings", StatusWarn, err.Error())
		return
	}
	if models.HasModel(avail, mc.Model) {
		r.add("embeddings", StatusOK, fmt.Sprintf("model %q present", mc.Model))
	} else {
		r.add("embeddings", StatusWarn, fmt.Sprintf("model %q not found", mc.Model))
	}
}

func (r *DoctorReport) add(name string, status CheckStatus, detail string) {
	r.Checks = append(r.Checks, Check{Name: name, Status: status, Detail: detail})
}

// ensureWritable creates dir if needed and probes writability.
func ensureWritable(dir string) error {
	if dir == "" {
		return fmt.Errorf("empty path")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	probe := filepath.Join(dir, ".tldg-write-probe")
	f, err := os.Create(probe)
	if err != nil {
		return fmt.Errorf("%s not writable: %w", dir, err)
	}
	f.Close()
	os.Remove(probe)
	return nil
}

// RenderDoctorText writes a human-readable doctor report.
func RenderDoctorText(w io.Writer, r *DoctorReport) error {
	icon := map[CheckStatus]string{
		StatusOK: "✓", StatusWarn: "!", StatusFail: "✗", StatusNA: "-",
	}
	var b strings.Builder
	b.WriteString("tldg doctor\n\n")
	for _, c := range r.Checks {
		fmt.Fprintf(&b, "  %s %-12s %s\n", icon[c.Status], c.Name, c.Detail)
	}
	if r.HasFailure() {
		b.WriteString("\nOne or more checks failed.\n")
	} else {
		b.WriteString("\nAll required checks passed.\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// RenderDoctorJSON writes the doctor report as JSON.
func RenderDoctorJSON(w io.Writer, r *DoctorReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}
