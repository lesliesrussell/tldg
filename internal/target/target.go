// Package target parses and resolves a tldg target (spec §6). M0/M1 fully
// supports local paths; remote/provider/qualified forms are recognized but
// deferred to milestone 4.
package target

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/leslierussell/tldg/internal/git"
)

// tldg-5xh

// Kind classifies a parsed target.
type Kind string

const (
	KindLocal        Kind = "local"
	KindGitRemote    Kind = "git_remote"
	KindProviderURL  Kind = "provider_url"
	KindQualifiedRef Kind = "qualified_ref"
)

// ErrRemoteUnsupported indicates a non-local target that requires host adapters
// (milestone 4), not yet available.
var ErrRemoteUnsupported = errors.New("remote and provider targets are supported in milestone 4; use a local path for now")

// Target is a parsed and (for local) resolved analysis target.
type Target struct {
	Kind     Kind
	Raw      string
	Path     string       // absolute path for local targets
	Identity git.Identity // populated for local targets
	Reduced  bool         // true when a local dir is not a Git worktree
}

// qualifiedRe matches "host:owner/repo" style references (spec §6.4).
var qualifiedRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+:[^/].*/.+$`)

// Classify determines a target's Kind without resolving it.
func Classify(raw string) Kind {
	s := strings.TrimSpace(raw)
	switch {
	case s == "." || s == ".." || strings.HasPrefix(s, "/") ||
		strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") ||
		strings.HasPrefix(s, "~"):
		return KindLocal
	case strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://"):
		return KindProviderURL
	case strings.HasPrefix(s, "git@") || strings.HasPrefix(s, "ssh://") ||
		strings.HasPrefix(s, "git://"):
		return KindGitRemote
	case qualifiedRe.MatchString(s):
		return KindQualifiedRef
	default:
		// Bare relative path (e.g. "some-dir").
		return KindLocal
	}
}

// Resolve parses raw and, for local targets, resolves the worktree identity via
// the provided Git backend. Non-local targets return ErrRemoteUnsupported.
func Resolve(ctx context.Context, backend git.Backend, raw string) (*Target, error) {
	kind := Classify(raw)
	t := &Target{Kind: kind, Raw: raw}
	if kind != KindLocal {
		return t, ErrRemoteUnsupported
	}

	abs, err := expandLocal(raw)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("target %q: %w", raw, err)
	}
	if !info.IsDir() {
		abs = filepath.Dir(abs)
	}
	t.Path = abs

	if backend != nil && backend.IsWorktree(ctx, abs) {
		id, err := backend.InspectIdentity(ctx, abs)
		if err != nil {
			return nil, fmt.Errorf("inspect git identity: %w", err)
		}
		t.Identity = id
		if id.WorktreeRoot != "" {
			t.Path = id.WorktreeRoot
		}
	} else {
		t.Reduced = true // non-Git directory, reduced-mode analysis
	}
	return t, nil
}

// expandLocal expands ~ and returns an absolute, cleaned path.
func expandLocal(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "~" || strings.HasPrefix(s, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if s == "~" {
			s = home
		} else {
			s = filepath.Join(home, s[2:])
		}
	}
	abs, err := filepath.Abs(s)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}
