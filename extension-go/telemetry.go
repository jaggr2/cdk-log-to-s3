package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Listener owns the local HTTP endpoint the Telemetry API delivers to, the
// buffer, and the background flush tickers.
type Listener struct {
	cfg    *Config
	parser *Parser
	buffer *LogBuffer
	up     *Uploader
	server *http.Server

	cancel context.CancelFunc
	done   chan struct{}

	// runtimeDone is signalled when the Telemetry API reports that an
	// invocation finished. The event loop waits on it before releasing the
	// sandbox; see AwaitRuntimeDone.
	runtimeDone chan struct{}
}

func NewListener(cfg *Config, up *Uploader) *Listener {
	l := &Listener{
		cfg:    cfg,
		parser: NewParser(cfg, time.Now),
		up:     up,
		done:   make(chan struct{}),
		// Buffered and non-blocking, so a delivery never stalls on a loop that
		// is not currently waiting.
		runtimeDone: make(chan struct{}, 1),
	}
	l.buffer = NewLogBuffer(cfg.MaxBufferBytes, l.flushEntries)
	return l
}

func (l *Listener) Buffer() *LogBuffer { return l.buffer }

// Start brings up the HTTP listener and the two background flush triggers.
// Both tickers are tied to a context so Stop actually stops them; the previous
// implementation leaked them for the lifetime of the sandbox.
func (l *Listener) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", l.handleTelemetry)

	l.server = &http.Server{
		Addr:              ":" + l.cfg.TelemetryPort,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ready := make(chan error, 1)
	go func() {
		l.debugf("telemetry listener starting on port %s", l.cfg.TelemetryPort)
		ready <- nil
		if err := l.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "LOG_TO_S3_LISTENER_ERROR: %v\n", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	l.cancel = cancel

	flushInterval := l.cfg.FlushInterval
	if flushInterval <= 0 {
		flushInterval = 15 * time.Second
	}

	go func() {
		defer close(l.done)
		interval := time.NewTicker(flushInterval)
		defer interval.Stop()
		// The size threshold is polled rather than checked on append so a
		// burst of small records cannot hold the append lock during an upload.
		size := time.NewTicker(time.Second)
		defer size.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-interval.C:
				if l.buffer.SizeBytes() > 0 {
					l.flushLogged("timer")
				}
			case <-size.C:
				if l.buffer.Full() {
					l.flushLogged("size")
				}
			}
		}
	}()

	return <-ready
}

func (l *Listener) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	var batches []TelemetryBatch
	if err := json.Unmarshal(body, &batches); err != nil {
		l.debugf("telemetry parse error: %v", err)
		http.Error(w, "parse error", http.StatusBadRequest)
		return
	}

	// Logged under debug because silence here is otherwise indistinguishable
	// from a destination Lambda cannot reach. The types matter as much as the
	// count when diagnosing which records are and are not being delivered.
	if l.cfg.Debug {
		types := make([]string, 0, len(batches))
		for _, b := range batches {
			types = append(types, b.Type)
		}
		l.debugf("received %d telemetry event(s): %s", len(batches), strings.Join(types, ","))
	}

	flushRequested := false
	for _, batch := range batches {
		entries, flush := l.parser.ParseBatch(batch)
		for _, entry := range entries {
			if l.shouldInclude(entry) {
				l.buffer.Append(entry)
			}
		}
		flushRequested = flushRequested || flush
	}

	l.debugf("batch parsed: buffered=%d bytes=%d runtimeDone=%v",
		l.buffer.Len(), l.buffer.SizeBytes(), flushRequested)

	w.WriteHeader(http.StatusOK)

	// Signal only - the flush deliberately does NOT happen here.
	//
	// This handler runs on the HTTP server goroutine. Uploading from it races
	// the event loop: the loop stops waiting, calls /event/next, and Lambda
	// freezes the environment with the S3 request still in flight, so the
	// object is never written and nothing reports an error. The event loop
	// owns the flush instead, because it is what holds the invocation open.
	if flushRequested {
		select {
		case l.runtimeDone <- struct{}{}:
		default:
		}
	}
}

// AwaitRuntimeDone blocks until the Telemetry API reports that the current
// invocation finished, or until the timeout elapses. It reports whether the
// signal arrived.
//
// This is what keeps the log shipping honest. Lambda freezes the execution
// environment as soon as the runtime has responded AND every extension has
// called /event/next. An extension that calls it immediately is frozen before
// the Telemetry API delivers that invocation's records, so the logs sit in the
// platform buffer until the environment thaws again - the next invocation, or
// shutdown. On a busy function that merely delays delivery by one invocation;
// on a quiet one the logs can be minutes late, and a final invocation before
// an idle shutdown is only saved by the shutdown flush.
//
// Waiting costs billed duration, since Lambda bills until the last extension
// releases the invocation. A zero timeout opts out and restores the older
// deliver-on-next-invocation behaviour.
func (l *Listener) AwaitRuntimeDone(timeout time.Duration) bool {
	if timeout <= 0 {
		return false
	}
	// A signal already buffered is honoured rather than discarded: the
	// Telemetry API often delivers before the loop gets here, and throwing it
	// away would burn the whole timeout and then flush anyway. Returning early
	// costs nothing, because the caller flushes whatever is buffered either
	// way.
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-l.runtimeDone:
		return true
	case <-timer.C:
		l.debugf("timed out after %s waiting for runtimeDone", timeout)
		return false
	}
}

func (l *Listener) shouldInclude(entry LogEntry) bool {
	lvl, ok := ParseLevel(entry.Level)
	if !ok {
		// An unrecognised level is kept: dropping a record because its level
		// was spelled unusually would lose real errors.
		return true
	}
	return lvl >= l.cfg.Level
}

func (l *Listener) flushLogged(reason string) {
	if err := l.buffer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "LOG_TO_S3_FLUSH_ERROR %s: %v\n", reason, err)
	}
}

func (l *Listener) flushEntries(entries []LogEntry) error {
	if len(entries) == 0 || l.up == nil {
		return nil
	}

	data, err := WriteParquet(entries, l.cfg.Compression)
	if err != nil {
		return fmt.Errorf("write parquet: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), uploadTimeout)
	defer cancel()

	key, err := l.up.Upload(ctx, data, l.cfg.FunctionName, entries[0].CorrelationID)
	if err != nil {
		return fmt.Errorf("s3 upload %s: %w", key, err)
	}
	l.debugf("flushed %d entries (%d bytes) to %s", len(entries), len(data), key)
	return nil
}

// Stop drains the buffer and shuts the listener down.
func (l *Listener) Stop() {
	if l.cancel != nil {
		l.cancel()
		<-l.done
	}
	if l.buffer.SizeBytes() > 0 {
		l.flushLogged("shutdown")
	}
	if l.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		l.server.Shutdown(ctx)
	}
}

// debugf gates the chatty self-logging of this extension. It is off by default:
// every line it writes is itself billed CloudWatch ingest on every invocation.
func (l *Listener) debugf(format string, args ...interface{}) {
	if l.cfg.Debug {
		fmt.Fprintf(os.Stderr, "LOG-TO-S3: "+format+"\n", args...)
	}
}
