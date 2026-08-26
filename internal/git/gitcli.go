package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// tldg-5xh

// CLI is a Backend that shells out to the system git binary.
type CLI struct {
	// Bin is the git executable; defaults to "git".
	Bin string
}

// NewCLI returns a git CLI backend.
func NewCLI() *CLI { return &CLI{Bin: "git"} }

func (c *CLI) bin() string {
	if c.Bin == "" {
		return "git"
	}
	return c.Bin
}

// run executes git in dir and returns trimmed stdout. Stderr is folded into the
// error. File contents are never logged (spec §9.2).
func (c *CLI) run(ctx context.Context, dir string, args ...string) (string, error) {
	full := append([]string{"-C", dir}, args...)
	cmd := exec.CommandContext(ctx, c.bin(), full...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return strings.TrimSpace(out.String()), nil
}

// Available reports the git version.
func (c *CLI) Available(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, c.bin(), "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git not available: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// IsWorktree reports whether dir is inside a Git worktree.
func (c *CLI) IsWorktree(ctx context.Context, dir string) bool {
	out, err := c.run(ctx, dir, "rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// InspectIdentity collects repository identity (spec §9.2). Missing pieces are
// tolerated (e.g. no commits yet, no remotes, no tags).
func (c *CLI) InspectIdentity(ctx context.Context, dir string) (Identity, error) {
	id := Identity{}
	if !c.IsWorktree(ctx, dir) {
		return id, nil
	}
	id.IsGitRepo = true

	if root, err := c.run(ctx, dir, "rev-parse", "--show-toplevel"); err == nil {
		id.WorktreeRoot = root
	}

	// Branch / detached HEAD.
	if branch, err := c.run(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		if branch == "HEAD" {
			id.DetachedHEAD = true
		} else {
			id.Branch = branch
		}
	}

	// HEAD commit + timestamp.
	if h, err := c.run(ctx, dir, "rev-parse", "HEAD"); err == nil {
		id.HeadCommit = h
		if ts, err := c.run(ctx, dir, "log", "-1", "--format=%ct"); err == nil {
			if secs, perr := strconv.ParseInt(ts, 10, 64); perr == nil {
				id.HeadCommitTime = time.Unix(secs, 0)
			}
		}
	}

	// Remotes.
	if out, err := c.run(ctx, dir, "remote", "-v"); err == nil && out != "" {
		id.Remotes = parseRemotes(out)
	}

	// Nearest tag.
	if tag, err := c.run(ctx, dir, "describe", "--tags", "--abbrev=0"); err == nil {
		id.NearestTag = tag
	}

	// Dirty state (porcelain; contents not captured).
	if out, err := c.run(ctx, dir, "status", "--porcelain"); err == nil {
		id.Dirty = strings.TrimSpace(out) != ""
	}

	return id, nil
}

// parseRemotes parses `git remote -v` output into unique remotes (fetch URLs).
func parseRemotes(out string) []Remote {
	seen := map[string]bool{}
	var rs []Remote
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name, u := fields[0], fields[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		rs = append(rs, Remote{Name: name, URL: u, NormalizedURL: NormalizeRemoteURL(u)})
	}
	return rs
}

// NormalizeRemoteURL normalizes common Git remote forms to a comparable
// host/owner/repo shape. It strips credentials, ".git" suffix, and converts
// scp-like syntax (git@host:owner/repo) to host/owner/repo.
func NormalizeRemoteURL(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimSuffix(s, ".git")
	// scp-like: git@github.com:owner/repo
	if !strings.Contains(s, "://") && strings.Contains(s, "@") && strings.Contains(s, ":") {
		at := strings.Index(s, "@")
		rest := s[at+1:]
		return strings.Replace(rest, ":", "/", 1)
	}
	// URL forms: scheme://[user@]host/path
	if i := strings.Index(s, "://"); i >= 0 {
		rest := s[i+3:]
		if at := strings.Index(rest, "@"); at >= 0 {
			rest = rest[at+1:]
		}
		return rest
	}
	return s
}
