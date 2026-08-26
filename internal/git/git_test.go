package git

import "testing"

// tldg-5xh

func TestNormalizeRemoteURL(t *testing.T) {
	cases := map[string]string{
		"git@github.com:owner/repo.git":            "github.com/owner/repo",
		"https://gitlab.com/group/project.git":     "gitlab.com/group/project",
		"ssh://git@git.example.internal/team/p.git": "git.example.internal/team/p",
		"https://user@github.com/o/r":              "github.com/o/r",
		"git://github.com/o/r.git":                 "github.com/o/r",
	}
	for in, want := range cases {
		if got := NormalizeRemoteURL(in); got != want {
			t.Errorf("NormalizeRemoteURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseRemotes(t *testing.T) {
	out := "origin\tgit@github.com:owner/repo.git (fetch)\n" +
		"origin\tgit@github.com:owner/repo.git (push)\n" +
		"upstream\thttps://github.com/up/repo.git (fetch)\n"
	rs := parseRemotes(out)
	if len(rs) != 2 {
		t.Fatalf("expected 2 unique remotes, got %d: %+v", len(rs), rs)
	}
	if rs[0].Name != "origin" || rs[0].NormalizedURL != "github.com/owner/repo" {
		t.Errorf("unexpected first remote: %+v", rs[0])
	}
}
