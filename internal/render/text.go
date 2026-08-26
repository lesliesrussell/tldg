package render

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// tldg-5xh

// JSON writes the result as stable, indented JSON (spec §8.2).
func JSON(w io.Writer, res *Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(res)
}

// Text writes answer-first terminal output with a trailing evidence list
// (spec §8.1).
func Text(w io.Writer, res *Result) error {
	var b strings.Builder

	if res.Answer.Text != "" {
		b.WriteString(res.Answer.Text)
		b.WriteString("\n")
	}

	if res.Answer.Confidence != "" {
		fmt.Fprintf(&b, "\nConfidence: %s\n", res.Answer.Confidence)
	}

	if len(res.Warnings) > 0 {
		b.WriteString("\nWarnings\n")
		for _, warn := range res.Warnings {
			fmt.Fprintf(&b, "  - %s\n", warn)
		}
	}

	if len(res.Evidence) > 0 {
		b.WriteString("\nEvidence\n")
		for _, e := range res.Evidence {
			if e.Citation != "" {
				fmt.Fprintf(&b, "  [%s]", e.Citation)
			} else {
				fmt.Fprintf(&b, "  %s", e.URI)
			}
			if e.Title != "" {
				fmt.Fprintf(&b, " — %s", e.Title)
			}
			b.WriteString("\n")
		}
	}

	_, err := io.WriteString(w, b.String())
	return err
}

// Markdown writes a Markdown rendering of the result (spec §8, --markdown).
func Markdown(w io.Writer, res *Result) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# tldg %s\n\n", res.Command)
	if res.Answer.Text != "" {
		b.WriteString(res.Answer.Text)
		b.WriteString("\n\n")
	}
	if res.Answer.Confidence != "" {
		fmt.Fprintf(&b, "**Confidence:** %s\n\n", res.Answer.Confidence)
	}
	if len(res.Warnings) > 0 {
		b.WriteString("## Warnings\n\n")
		for _, warn := range res.Warnings {
			fmt.Fprintf(&b, "- %s\n", warn)
		}
		b.WriteString("\n")
	}
	if len(res.Evidence) > 0 {
		b.WriteString("## Evidence\n\n")
		for _, e := range res.Evidence {
			id := e.URI
			if e.Citation != "" {
				id = "`[" + e.Citation + "]`"
			}
			if e.Title != "" {
				fmt.Fprintf(&b, "- %s — %s\n", id, e.Title)
			} else {
				fmt.Fprintf(&b, "- %s\n", id)
			}
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}
