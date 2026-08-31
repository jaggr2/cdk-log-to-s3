package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
}

func NewListener(cfg *Config, up *Uploader) *Listener {
	l := &Listener{
		cfg:    cfg,
		parser: NewParser(cfg, time.Now),
		up:     up,
		done:   make(chan struct{}),
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

	w.WriteHeader(http.StatusOK)

	// runtimeDone means the handler has returned and the sandbox may be frozen
	// before the next tick, so this flush cannot wait for the timer.
	if flushRequested && l.buffer.SizeBytes() > 0 {
		l.flushLogged("runtimeDone")
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
