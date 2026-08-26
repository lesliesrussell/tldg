package evidence

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lesliesrussell/tldg/internal/scan"
)

// tldg-eca

// BundleOptions bounds grounding-bundle construction (spec §10.4).
type BundleOptions struct {
	MaxChunks      int
	MaxChunkTokens int
	Ref            string
	Now            time.Time
}

// Snippet pairs an evidence record with the content to send to the model.
type Snippet struct {
	Evidence Evidence
	Content  string
}

// Bundle is a ranked, size-budgeted set of grounding snippets.
type Bundle struct {
	Snippets []Snippet
	Notes    []string
}

// Evidences returns just the evidence records (for output).
func (b *Bundle) Evidences() []Evidence {
	out := make([]Evidence, 0, len(b.Snippets))
	for _, s := range b.Snippets {
		out = append(out, s.Evidence)
	}
	return out
}

// BuildBundle reads the highest-priority selected files and produces a bounded,
// diverse set of grounding snippets with stable citation IDs (spec §10.4).
// Files are consumed in the order provided (already priority-sorted by scan).
func BuildBundle(files []scan.File, opts BundleOptions) *Bundle {
	if opts.MaxChunks <= 0 {
		opts.MaxChunks = 24
	}
	if opts.MaxChunkTokens <= 0 {
		opts.MaxChunkTokens = 1200
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	b := &Bundle{}
	seenCat := map[scan.Category]int{} // favor category diversity
	n := 0
	for _, f := range files {
		if n >= opts.MaxChunks {
			b.Notes = append(b.Notes, fmt.Sprintf("evidence truncated at %d chunks", opts.MaxChunks))
			break
		}
		// Cap over-representation of any single category.
		if seenCat[f.Category] >= 6 {
			continue
		}
		content, lines, ok := headExcerpt(f.AbsPath, opts.MaxChunkTokens)
		if !ok || strings.TrimSpace(content) == "" {
			continue
		}
		id := fmt.Sprintf("ev_%03d", n+1)
		ev := LocalFile(id, f.RelPath, opts.Ref, 1, lines, opts.Now)
		ev.Title = string(f.Category)
		sum := sha256.Sum256([]byte(content))
		ev.ContentHash = "sha256:" + hex.EncodeToString(sum[:])
		ev.Content = content
		b.Snippets = append(b.Snippets, Snippet{Evidence: ev, Content: content})
		seenCat[f.Category]++
		n++
	}
	return b
}

// headExcerpt reads up to a token budget (~4 chars/token) from the start of a
// file, returning the text and the number of lines included.
func headExcerpt(abs string, maxTokens int) (string, int, bool) {
	f, err := os.Open(abs)
	if err != nil {
		return "", 0, false
	}
	defer f.Close()
	maxBytes := maxTokens * 4
	var sb strings.Builder
	lines := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if sb.Len()+len(line)+1 > maxBytes {
			break
		}
		sb.WriteString(line)
		sb.WriteByte('\n')
		lines++
	}
	return sb.String(), lines, true
}
