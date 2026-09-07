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

// runtimeDone signals the event loop; it must NOT upload from this goroutine.
//
// Uploading here races the event loop: the loop stops waiting, calls
// /event/next, and Lambda freezes the environment with the S3 request still in
// flight. The object is never written and nothing reports an error.
func TestHandleTelemetryRuntimeDoneSignalsWithoutFlushing(t *testing.T) {
	l, flushed, mu := testListener(t, LevelInfo)

	post(t, l, `[{"type":"function","record":"{\"__log_level\":\"INFO\",\"message\":\"before done\"}"}]`)
	if l.buffer.Len() != 1 {
		t.Fatalf("buffered %d entries, want 1", l.buffer.Len())
	}

	post(t, l, `[{"type":"platform.runtimeDone","record":{"requestId":"r","status":"success"}}]`)

	mu.Lock()
	if len(*flushed) != 0 {
		mu.Unlock()
		t.Fatalf("the telemetry handler uploaded %d entries; the event loop owns the flush", len(*flushed))
	}
	mu.Unlock()

	if l.buffer.Len() != 1 {
		t.Errorf("buffer holds %d entries, want the records left for the event loop", l.buffer.Len())
	}
	if !l.AwaitRuntimeDone(time.Second) {
		t.Error("runtimeDone was not signalled to the event loop")
	}

	// What the event loop then does.
	if err := l.buffer.Flush(); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(*flushed) != 1 {
		t.Errorf("flushed %d entries, want 1", len(*flushed))
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

// The event loop must hold the sandbox open until the invocation's telemetry
// has been delivered. Releasing immediately freezes the environment with the
// records still in the platform buffer.
func TestAwaitRuntimeDoneUnblocksOnRuntimeDone(t *testing.T) {
	l, _, _ := testListener(t, LevelInfo)

	go func() {
		post(t, l, `[{"type":"function","record":"{\"__log_level\":\"INFO\",\"message\":\"x\"}"}]`)
		post(t, l, `[{"type":"platform.runtimeDone","record":{"requestId":"r","status":"success"}}]`)
	}()

	if !l.AwaitRuntimeDone(5 * time.Second) {
		t.Fatal("AwaitRuntimeDone did not observe the runtimeDone event")
	}
}

func TestAwaitRuntimeDoneTimesOut(t *testing.T) {
	l, _, _ := testListener(t, LevelInfo)

	start := time.Now()
	if l.AwaitRuntimeDone(50 * time.Millisecond) {
		t.Error("expected a timeout, not a signal")
	}
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Errorf("returned after %s, want at least the timeout", elapsed)
	}
}

// Zero opts out, for callers who would rather not pay the billed duration.
func TestAwaitRuntimeDoneZeroReturnsImmediately(t *testing.T) {
	l, _, _ := testListener(t, LevelInfo)

	start := time.Now()
	if l.AwaitRuntimeDone(0) {
		t.Error("a zero timeout should report no signal")
	}
	if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
		t.Errorf("blocked for %s, want an immediate return", elapsed)
	}
}

// The Telemetry API frequently delivers before the event loop starts waiting.
// That signal must still count, or every invocation burns the full timeout.
func TestAwaitRuntimeDoneHonoursASignalThatArrivedFirst(t *testing.T) {
	l, _, _ := testListener(t, LevelInfo)

	post(t, l, `[{"type":"platform.runtimeDone","record":{"requestId":"a","status":"success"}}]`)

	start := time.Now()
	if !l.AwaitRuntimeDone(2 * time.Second) {
		t.Fatal("a signal delivered before the wait was discarded")
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Errorf("waited %s for a signal that had already arrived", elapsed)
	}
}
