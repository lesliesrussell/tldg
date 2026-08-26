package index

import "testing"

// tldg-eca

func TestFTSRoundTrip(t *testing.T) {
	st, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer st.Close()
	if err := st.InitSchema(); err != nil {
		t.Fatalf("schema: %v", err)
	}
	chunks := []Chunk{
		{RelPath: "README.md", LineStart: 1, LineEnd: 10, Category: "readme", Content: "tldg is a local repository intelligence tool"},
		{RelPath: "main.go", LineStart: 1, LineEnd: 20, Category: "entrypoint", Content: "package main func main prints hello"},
	}
	for _, c := range chunks {
		if err := st.AddChunk(c); err != nil {
			t.Fatalf("add: %v", err)
		}
	}
	if n, _ := st.Count(); n != 2 {
		t.Fatalf("count = %d, want 2", n)
	}
	hits, err := st.Search("repository intelligence", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) == 0 || hits[0].RelPath != "README.md" {
		t.Fatalf("expected README.md hit, got %+v", hits)
	}
}

func TestResetRecreates(t *testing.T) {
	st, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer st.Close()
	if err := st.InitSchema(); err != nil {
		t.Fatal(err)
	}
	st.AddChunk(Chunk{RelPath: "a", Content: "alpha"})
	if err := st.Reset(); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if n, _ := st.Count(); n != 0 {
		t.Errorf("after reset count = %d, want 0", n)
	}
	// Index must still be usable after reset.
	if err := st.AddChunk(Chunk{RelPath: "b", Content: "beta"}); err != nil {
		t.Errorf("add after reset: %v", err)
	}
}
