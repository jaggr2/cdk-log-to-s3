package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSharedContractFixtures reads the same fixture file the logger package
// asserts against, so a change to the wire format on either side fails a test
// here rather than silently producing a Parquet column full of nulls.
//
// See docs/CONTRACT.md.
func TestSharedContractFixtures(t *testing.T) {
	path := filepath.Join("..", "docs", "contract-fixtures.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}

	var fixtures struct {
		Cases []struct {
			Name   string `json:"name"`
			Line   string `json:"line"`
			Expect struct {
				Level         string  `json:"level"`
				Source        string  `json:"source"`
				CorrelationID string  `json:"correlationId"`
				Message       string  `json:"message"`
				Context       *string `json:"context"`
				StackTrace    *string `json:"stackTrace"`
				Caller        *string `json:"caller"`
			} `json:"expect"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatalf("cannot parse %s: %v", path, err)
	}
	if len(fixtures.Cases) == 0 {
		t.Fatal("no fixture cases found")
	}

	cfg := &Config{FunctionName: "my-fn", IncludePlatformReport: true, Level: LevelDebug}
	fixed := time.Date(2026, 8, 31, 14, 5, 0, 0, time.UTC)
	p := NewParser(cfg, func() time.Time { return fixed })

	for _, c := range fixtures.Cases {
		t.Run(c.Name, func(t *testing.T) {
			// The runtime delivers a function log as a JSON string.
			record, err := json.Marshal(c.Line)
			if err != nil {
				t.Fatal(err)
			}

			entries, _ := p.ParseBatch(TelemetryBatch{
				Time: "2026-08-31T14:05:00.000Z", Type: "function", Record: record,
			})
			if len(entries) != 1 {
				t.Fatalf("got %d entries, want 1", len(entries))
			}
			e := entries[0]

			if e.Level != c.Expect.Level {
				t.Errorf("Level = %q, want %q", e.Level, c.Expect.Level)
			}
			if e.Source != c.Expect.Source {
				t.Errorf("Source = %q, want %q", e.Source, c.Expect.Source)
			}
			if e.CorrelationID != c.Expect.CorrelationID {
				t.Errorf("CorrelationID = %q, want %q", e.CorrelationID, c.Expect.CorrelationID)
			}
			if e.Message != c.Expect.Message {
				t.Errorf("Message = %q, want %q", e.Message, c.Expect.Message)
			}
			assertOptional(t, "Context", e.Context, c.Expect.Context)
			assertOptional(t, "StackTrace", e.StackTrace, c.Expect.StackTrace)
			assertOptional(t, "Caller", e.Caller, c.Expect.Caller)
		})
	}
}

func assertOptional(t *testing.T, field string, got, want *string) {
	t.Helper()
	switch {
	case want == nil && got != nil:
		t.Errorf("%s = %q, want null", field, *got)
	case want != nil && got == nil:
		t.Errorf("%s = null, want %q", field, *want)
	case want != nil && got != nil && *got != *want:
		t.Errorf("%s = %q, want %q", field, *got, *want)
	}
}
