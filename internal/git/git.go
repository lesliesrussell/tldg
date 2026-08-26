// Package git inspects Git worktrees via a backend abstraction. The default
// backend shells out to the system git binary (spec §26 decision 1, §9.2).
package git

import (
	"context"
	"time"
)

// tldg-5xh

// Remote is a configured Git remote with a normalized URL.
type Remote struct {
	Name          string
	URL           string
	NormalizedURL string
}

// Identity is the collected repository identity for a worktree (spec §9.2).
type Identity struct {
	WorktreeRoot   string
	Branch         string
	DetachedHEAD   bool
	HeadCommit     string
	HeadCommitTime time.Time
	Remotes        []Remote
	NearestTag     string
	Dirty          bool
	IsGitRepo      bool
}

// Backend abstracts Git object/history access so the CLI implementation can be
// swapped for go-git later (spec §26).
type Backend interface {
	// Available reports whether the backend can run (e.g. git on PATH).
	Available(ctx context.Context) (version string, err error)
	// IsWorktree reports whether dir is within a Git worktree.
	IsWorktree(ctx context.Context, dir string) bool
	// InspectIdentity collects repository identity for the worktree at dir.
	InspectIdentity(ctx context.Context, dir string) (Identity, error)
}
