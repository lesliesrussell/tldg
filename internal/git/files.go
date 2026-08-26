package git

import (
	"context"
	"strings"
)

// tldg-eca

// ListFiles returns repository files relative to the worktree root. Tracked
// files always honor .gitignore via git. When includeUntracked is true,
// untracked-but-not-ignored files are appended (git --exclude-standard).
func (c *CLI) ListFiles(ctx context.Context, dir string, includeUntracked bool) ([]string, error) {
	tracked, err := c.run(ctx, dir, "ls-files", "-z")
	if err != nil {
		return nil, err
	}
	files := splitZ(tracked)
	if includeUntracked {
		others, err := c.run(ctx, dir, "ls-files", "-z", "--others", "--exclude-standard")
		if err == nil {
			files = append(files, splitZ(others)...)
		}
	}
	return files, nil
}

// splitZ splits NUL-delimited git output, dropping empties.
func splitZ(s string) []string {
	var out []string
	for _, p := range strings.Split(s, "\x00") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
