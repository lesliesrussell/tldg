package render

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/lesliesrussell/tldg/internal/evidence"
)

// tldg-5xh

func TestJSONEnvelope(t *testing.T) {
	res := NewResult("summary")
	res.Answer = Answer{Text: "hi", Confidence: evidence.ConfidenceVerified}
	res.Evidence = []evidence.Evidence{{ID: "ev_1", Citation: "local:go.mod"}}

	var buf bytes.Buffer
	if err := JSON(&buf, res); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var back map[string]any
	if err := json.Unmarshal(buf.Bytes(), &back); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if back["schema_version"] != "1.0" {
		t.Errorf("schema_version = %v", back["schema_version"])
	}
	if back["command"] != "summary" {
		t.Errorf("command = %v", back["command"])
	}
	if _, ok := back["warnings"]; !ok {
		t.Error("warnings key must always be present")
	}
}

func TestTextAnswerFirst(t *testing.T) {
	res := NewResult("summary")
	res.Answer = Answer{Text: "The answer.", Confidence: evidence.ConfidenceStrong}
	res.Evidence = []evidence.Evidence{{Citation: "local:README.md:1-4", Title: "readme"}}
	var buf bytes.Buffer
	if err := Text(&buf, res); err != nil {
		t.Fatalf("Text: %v", err)
	}
	out := buf.String()
	if idx := bytes.Index(buf.Bytes(), []byte("The answer.")); idx != 0 {
		t.Errorf("answer should come first, output: %q", out)
	}
	if !bytes.Contains(buf.Bytes(), []byte("[local:README.md:1-4]")) {
		t.Errorf("evidence citation missing: %q", out)
	}
}
