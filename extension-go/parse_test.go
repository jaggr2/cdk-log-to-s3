package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

const testReqID = "8f2c1a44-0d3e-4b6a-9c11-2f7d5e6a8b90"

func testParser(t *testing.T) *Parser {
	t.Helper()
	cfg := &Config{
		FunctionName:          "my-fn",
		IncludePlatformReport: true,
		Level:                 LevelInfo,
	}
	fixed := time.Date(2026, 8, 31, 14, 5, 0, 0, time.UTC)
	return NewParser(cfg, func() time.Time { return fixed })
}

// functionBatch builds the wire shape of a text-format function record: the
// Telemetry API delivers it as a JSON *string*.
func functionBatch(t *testing.T, record string) TelemetryBatch {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return TelemetryBatch{Time: "2026-08-31T14:05:00.000Z", Type: "function", Record: raw}
}

func TestParseFunctionStructuredWithCloudWatchPrefix(t *testing.T) {
	p := testParser(t)
	body := `{"__log_level":"WARN","__source":"SyncProcessor","__correlation_id":"cid-1","__caller":"src/lib/sync.ts:142:11","message":"Events synced","context":{"count":5}}`
	record := "2026-08-31T14:05:00.123Z\t" + testReqID + "\tWARN\t" + body + "\n"

	entries, flush := p.ParseBatch(functionBatch(t, record))
	if flush {
		t.Error("a function record must not request a flush")
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	e := entries[0]

	if e.Level != "WARN" {
		t.Errorf("Level = %q", e.Level)
	}
	if e.Source != "SyncProcessor" {
		t.Errorf("Source = %q", e.Source)
	}
	if e.CorrelationID != "cid-1" {
		t.Errorf("CorrelationID = %q", e.CorrelationID)
	}
	if e.Message != "Events synced" {
		t.Errorf("Message = %q", e.Message)
	}
	if e.FunctionName != "my-fn" {
		t.Errorf("FunctionName = %q", e.FunctionName)
	}
	// The request id used to be discarded together with the prefix.
	if e.RequestID != testReqID {
		t.Errorf("RequestID = %q, want %q", e.RequestID, testReqID)
	}
	if e.Timestamp != "2026-08-31T14:05:00.123Z" {
		t.Errorf("Timestamp = %q, want the timestamp from the record prefix", e.Timestamp)
	}
	if optStr(e.Caller) != "src/lib/sync.ts:142:11" {
		t.Errorf("Caller = %q", optStr(e.Caller))
	}
	if optStr(e.Context) != `{"count":5}` {
		t.Errorf("Context = %q", optStr(e.Context))
	}
	if e.StackTrace != nil {
		t.Errorf("StackTrace = %q, want nil when absent", optStr(e.StackTrace))
	}
}

func TestParseFunctionStructuredWithoutPrefix(t *testing.T) {
	p := testParser(t)
	body := `{"__log_level":"ERROR","__source":"Api","__correlation_id":"","message":"boom","stackTrace":"Error: boom\n    at handler (index.js:1:1)"}`

	entries, _ := p.ParseBatch(functionBatch(t, body))
	e := entries[0]

	if e.Level != "ERROR" {
		t.Errorf("Level = %q", e.Level)
	}
	if e.Message != "boom" {
		t.Errorf("Message = %q", e.Message)
	}
	if !strings.Contains(optStr(e.StackTrace), "at handler") {
		t.Errorf("StackTrace = %q", optStr(e.StackTrace))
	}
	if e.RequestID != "" {
		t.Errorf("RequestID = %q, want empty when there is no prefix", e.RequestID)
	}
	// Falls back to the batch timestamp.
	if e.Timestamp != "2026-08-31T14:05:00.000Z" {
		t.Errorf("Timestamp = %q", e.Timestamp)
	}
}

func TestParseFunctionPlainText(t *testing.T) {
	p := testParser(t)
	entries, _ := p.ParseBatch(functionBatch(t, "2026-08-31T14:05:00.123Z\t"+testReqID+"\tINFO\thello world\n"))
	e := entries[0]

	if e.Level != "INFO" {
		t.Errorf("Level = %q", e.Level)
	}
	if e.Message != "hello world" {
		t.Errorf("Message = %q", e.Message)
	}
	if e.RequestID != testReqID {
		t.Errorf("RequestID = %q", e.RequestID)
	}
}

func TestParseFunctionPlainTextErrorHeuristic(t *testing.T) {
	p := testParser(t)
	entries, _ := p.ParseBatch(functionBatch(t, "something went wrong: error connecting"))
	if entries[0].Level != "ERROR" {
		t.Errorf("Level = %q, want ERROR", entries[0].Level)
	}
}

// The old parser searched for the last tab-brace pair anywhere in the record,
// so an ordinary log line containing one was silently truncated.
func TestParseFunctionPlainTextContainingTabBrace(t *testing.T) {
	p := testParser(t)
	record := "payload was:\t{not really json"

	entries, _ := p.ParseBatch(functionBatch(t, record))
	if entries[0].Message != record {
		t.Errorf("Message = %q, want the record kept whole", entries[0].Message)
	}
}

// JSON that lacks the __log_level sentinel is not one of our records and must
// fall through to plain-text handling rather than being half-parsed.
func TestParseFunctionJSONWithoutSentinel(t *testing.T) {
	p := testParser(t)
	body := `{"message":"not ours","level":"INFO"}`

	entries, _ := p.ParseBatch(functionBatch(t, body))
	e := entries[0]
	if e.Message != body {
		t.Errorf("Message = %q, want the raw JSON text", e.Message)
	}
	if e.Source != "" {
		t.Errorf("Source = %q, want empty", e.Source)
	}
}

// LAMBDA_LOG_FORMAT=JSON delivers record as an object, not a string. The
// previous implementation produced garbage for this shape.
func TestParseFunctionJSONLogFormat(t *testing.T) {
	p := testParser(t)
	inner := `{"__log_level":"DEBUG","__source":"Worker","__correlation_id":"cid-9","message":"tick"}`
	record := map[string]interface{}{
		"timestamp": "2026-08-31T14:06:00.000Z",
		"level":     "debug",
		"requestId": testReqID,
		"message":   inner,
	}
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}

	entries, _ := p.ParseBatch(TelemetryBatch{Type: "function", Record: raw})
	e := entries[0]

	if e.Level != "DEBUG" {
		t.Errorf("Level = %q", e.Level)
	}
	if e.Source != "Worker" {
		t.Errorf("Source = %q", e.Source)
	}
	if e.CorrelationID != "cid-9" {
		t.Errorf("CorrelationID = %q", e.CorrelationID)
	}
	if e.Message != "tick" {
		t.Errorf("Message = %q", e.Message)
	}
	if e.RequestID != testReqID {
		t.Errorf("RequestID = %q", e.RequestID)
	}
	if e.Timestamp != "2026-08-31T14:06:00.000Z" {
		t.Errorf("Timestamp = %q", e.Timestamp)
	}
}

func TestParseFunctionJSONLogFormatPlainMessage(t *testing.T) {
	p := testParser(t)
	raw := []byte(`{"timestamp":"2026-08-31T14:06:00.000Z","level":"ERROR","requestId":"` + testReqID + `","message":"plain text"}`)

	entries, _ := p.ParseBatch(TelemetryBatch{Type: "function", Record: raw})
	e := entries[0]
	if e.Level != "ERROR" {
		t.Errorf("Level = %q", e.Level)
	}
	if e.Message != "plain text" {
		t.Errorf("Message = %q", e.Message)
	}
}

func TestParseFunctionMultilineStackTrace(t *testing.T) {
	p := testParser(t)
	body := `{"__log_level":"ERROR","__source":"Api","message":"failed","stackTrace":"Error: failed\n    at a (x.js:1:1)\n    at b (y.js:2:2)"}`
	record := "2026-08-31T14:05:00.123Z\t" + testReqID + "\tERROR\t" + body

	entries, _ := p.ParseBatch(functionBatch(t, record))
	if strings.Count(optStr(entries[0].StackTrace), "\n") != 2 {
		t.Errorf("StackTrace lost its newlines: %q", optStr(entries[0].StackTrace))
	}
}

func TestParsePlatformRuntimeDoneRequestsFlush(t *testing.T) {
	p := testParser(t)
	entries, flush := p.ParseBatch(TelemetryBatch{
		Type:   "platform.runtimeDone",
		Record: json.RawMessage(`{"requestId":"` + testReqID + `","status":"success"}`),
	})
	if !flush {
		t.Error("platform.runtimeDone must request a flush")
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0", len(entries))
	}
}

func TestParsePlatformReport(t *testing.T) {
	p := testParser(t)
	record := `{"requestId":"` + testReqID + `","metrics":{"durationMs":123.4,"billedDurationMs":124,"memorySizeMB":512,"maxMemoryUsedMB":88}}`

	entries, flush := p.ParseBatch(TelemetryBatch{
		Time: "2026-08-31T14:05:00.000Z", Type: "platform.report",
		Record: json.RawMessage(record),
	})
	if flush {
		t.Error("platform.report must not request a flush")
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	e := entries[0]
	if e.Source != "platform" {
		t.Errorf("Source = %q", e.Source)
	}
	if e.RequestID != testReqID {
		t.Errorf("RequestID = %q", e.RequestID)
	}
	if !strings.Contains(e.Message, "123.4ms") {
		t.Errorf("Message = %q", e.Message)
	}
	var ctx map[string]float64
	if err := json.Unmarshal([]byte(optStr(e.Context)), &ctx); err != nil {
		t.Fatalf("Context is not valid JSON: %v", err)
	}
	if ctx["memoryUsedMB"] != 88 {
		t.Errorf("context memoryUsedMB = %v", ctx["memoryUsedMB"])
	}
}

func TestParsePlatformReportDisabled(t *testing.T) {
	p := testParser(t)
	p.cfg.IncludePlatformReport = false

	entries, _ := p.ParseBatch(TelemetryBatch{
		Type:   "platform.report",
		Record: json.RawMessage(`{"metrics":{"durationMs":1}}`),
	})
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0 when platform reports are disabled", len(entries))
	}
}

// The nested Logs API shape is still accepted so the extension is not tied to
// one generation of the API.
func TestParseLegacyNestedPlatformShape(t *testing.T) {
	p := testParser(t)

	entries, _ := p.ParseBatch(TelemetryBatch{
		Type:   "platform",
		Record: json.RawMessage(`{"type":"report","record":{"metrics":{"durationMs":7,"billedDurationMs":8,"memorySizeMB":128,"maxMemoryUsedMB":64}}}`),
	})
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}

	_, flush := p.ParseBatch(TelemetryBatch{
		Type:   "platform",
		Record: json.RawMessage(`{"type":"runtimeDone","record":{}}`),
	})
	if !flush {
		t.Error("nested runtimeDone must request a flush")
	}
}

func TestParseOtherPlatformEventsIgnored(t *testing.T) {
	p := testParser(t)
	for _, typ := range []string{"platform.initStart", "platform.start", "platform.extension"} {
		entries, flush := p.ParseBatch(TelemetryBatch{Type: typ, Record: json.RawMessage(`{}`)})
		if len(entries) != 0 || flush {
			t.Errorf("%s produced entries=%d flush=%v, want 0/false", typ, len(entries), flush)
		}
	}
}
