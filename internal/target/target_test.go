package target

import (
	"context"
	"errors"
	"testing"
)

// tldg-5xh

func TestClassify(t *testing.T) {
	cases := map[string]Kind{
		".":                              KindLocal,
		"./sub":                          KindLocal,
		"../up":                          KindLocal,
		"/abs/path":                      KindLocal,
		"~/src/x":                        KindLocal,
		"some-dir":                       KindLocal,
		"https://github.com/neovim/neovim": KindProviderURL,
		"git@github.com:owner/repo.git":  KindGitRemote,
		"ssh://git@host/team/repo.git":   KindGitRemote,
		"github:neovim/neovim":           KindQualifiedRef,
		"git.example.internal:team/svc":  KindQualifiedRef,
	}
	for in, want := range cases {
		if got := Classify(in); got != want {
			t.Errorf("Classify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveRemoteUnsupported(t *testing.T) {
	_, err := Resolve(context.Background(), nil, "github:owner/repo")
	if !errors.Is(err, ErrRemoteUnsupported) {
		t.Fatalf("expected ErrRemoteUnsupported, got %v", err)
	}
}

func TestResolveLocalDir(t *testing.T) {
	dir := t.TempDir()
	tg, err := Resolve(context.Background(), nil, dir)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if tg.Kind != KindLocal {
		t.Errorf("kind = %q, want local", tg.Kind)
	}
	if !tg.Reduced {
		t.Errorf("expected reduced mode for non-git dir")
	}
	if tg.Path != dir {
		t.Errorf("path = %q, want %q", tg.Path, dir)
	}
}

func TestResolveMissingTarget(t *testing.T) {
	_, err := Resolve(context.Background(), nil, "/no/such/path/xyzzy")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}
