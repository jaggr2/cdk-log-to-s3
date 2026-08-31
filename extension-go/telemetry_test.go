package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func testListener(t *testing.T, level Level) (*Listener, *[]LogEntry, *sync.Mutex) {
	t.Helper()

	cfg := &Config{
		FunctionName:          "my-fn",
		IncludePlatformReport: true,
		Level:                 level,
		MaxBufferBytes:        1 << 20,
		Compression:           "snappy",
		TelemetryPort:         "0",
		FlushInterval:         time.Second,
	}

	var mu sync.Mutex
	var flushed []LogEntry

	l := NewListener(cfg, nil)
	// Replace the S3-backed flush with an in-memory one; the upload path is
	// covered by writer_test.
	l.buffer = NewLogBuffer(cfg.MaxBufferBytes, func(e []LogEntry) error {
		mu.Lock()
		defer mu.Unlock()
		flushed = append(flushed, e...)
		return nil
	})
	return l, &flushed, &mu
}

func post(t *testing.T, l *Listener, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	l.handleTelemetry(rec, req)
	return rec
}

func TestHandleTelemetryBuffersFunctionRecords(t *testing.T) {
	l, _, _ := testListener(t, LevelInfo)

	body := `[
	  {"time":"2026-08-31T14:05:00.000Z","type":"function","record":"{\"__log_level\":\"INFO\",\"__source\":\"Api\",\"message\":\"hello\"}"},
	  {"time":"2026-08-31T14:05:01.000Z","type":"function","record":"{\"__log_level\":\"ERROR\",\"__source\":\"Api\",\"message\":\"boom\"}"}
	]`

	if rec := post(t, l, body); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if l.buffer.Len() != 2 {
		t.Fatalf("buffered %d entries, want 2", l.buffer.Len())
	}
}

func TestHandleTelemetryAppliesLevelThreshold(t *testing.T) {
	l, _, _ := testListener(t, LevelInfo)

	body := `[
	  {"type":"function","record":"{\"__log_level\":\"DEBUG\",\"message\":\"noisy\"}"},
	  {"type":"function","record":"{\"__log_level\":\"INFO\",\"message\":\"kept\"}"},
	  {"type":"function","record":"{\"__log_level\":\"WARN\",\"message\":\"kept\"}"}
	]`
	post(t, l, body)

	if l.buffer.Len() != 2 {
		t.Fatalf("buffered %d entries, want 2 (DEBUG dropped at LevelInfo)", l.buffer.Len())
	}
}

// The threshold is ordered now, not a DEBUG-versus-everything special case.
func TestHandleTelemetryErrorThresholdDropsWarn(t *testing.T) {
	l, _, _ := testListener(t, LevelError)

	body := `[
	  {"type":"function","record":"{\"__log_level\":\"WARN\",\"message\":\"dropped\"}"},
	  {"type":"function","record":"{\"__log_level\":\"ERROR\",\"message\":\"kept\"}"}
	]`
	post(t, l, body)

	if l.buffer.Len() != 1 {
		t.Fatalf("buffered %d entries, want 1", l.buffer.Len())
	}
}

func TestHandleTelemetryDebugLevelKeepsEverything(t *testing.T) {
	l, _, _ := testListener(t, LevelDebug)

	body := `[{"type":"function","record":"{\"__log_level\":\"DEBUG\",\"message\":\"kept\"}"}]`
	post(t, l, body)

	if l.buffer.Len() != 1 {
		t.Fatalf("buffered %d entries, want 1", l.buffer.Len())
	}
}

func TestHandleTelemetryRuntimeDoneTriggersFlush(t *testing.T) {
	l, flushed, mu := testListener(t, LevelInfo)

	post(t, l, `[{"type":"function","record":"{\"__log_level\":\"INFO\",\"message\":\"before done\"}"}]`)
	if l.buffer.Len() != 1 {
		t.Fatalf("buffered %d entries, want 1", l.buffer.Len())
	}

	post(t, l, `[{"type":"platform.runtimeDone","record":{"requestId":"r","status":"success"}}]`)

	mu.Lock()
	defer mu.Unlock()
	if len(*flushed) != 1 {
		t.Fatalf("flushed %d entries, want 1", len(*flushed))
	}
	if l.buffer.Len() != 0 {
		t.Errorf("buffer not drained after runtimeDone: %d", l.buffer.Len())
	}
}

func TestHandleTelemetryRejectsNonPost(t *testing.T) {
	l, _, _ := testListener(t, LevelInfo)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	l.handleTelemetry(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHandleTelemetryRejectsMalformedBody(t *testing.T) {
	l, _, _ := testListener(t, LevelInfo)

	if rec := post(t, l, "not json"); rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if l.buffer.Len() != 0 {
		t.Errorf("buffered %d entries from a malformed body", l.buffer.Len())
	}
}

// An unrecognised level must not be silently discarded: doing so would lose
// real errors from runtimes that spell levels differently.
func TestShouldIncludeKeepsUnknownLevels(t *testing.T) {
	l, _, _ := testListener(t, LevelError)

	if !l.shouldInclude(LogEntry{Level: "NOTICE"}) {
		t.Error("an unknown level should be kept")
	}
}

func TestListenerStartStop(t *testing.T) {
	l, _, _ := testListener(t, LevelInfo)
	l.cfg.TelemetryPort = "0" // let the OS pick a free port

	if err := l.Start(); err != nil {
		t.Fatal(err)
	}
	// Stop must return; before the tickers were context-bound they ran for the
	// lifetime of the process.
	l.Stop()
}
