package main

import "sync"

// LogBuffer accumulates entries between flushes. Callers append from the
// telemetry HTTP handler goroutine while the tickers and the event loop flush,
// so every field is mutex guarded.
type LogBuffer struct {
	mu       sync.Mutex
	entries  []LogEntry
	bytes    int
	maxBytes int
	flushFn  func(entries []LogEntry) error
}

func NewLogBuffer(maxBytes int, flushFn func([]LogEntry) error) *LogBuffer {
	return &LogBuffer{
		entries:  make([]LogEntry, 0, 1024),
		maxBytes: maxBytes,
		flushFn:  flushFn,
	}
}

func (b *LogBuffer) Append(entry LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = append(b.entries, entry)
	b.bytes += estimateSize(entry)
}

func (b *LogBuffer) SizeBytes() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.bytes
}

func (b *LogBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.entries)
}

// Full reports whether the buffer has reached its flush threshold.
func (b *LogBuffer) Full() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.maxBytes > 0 && b.bytes >= b.maxBytes
}

// Flush hands the accumulated entries to flushFn and resets the buffer. The
// reset happens before flushFn runs, so a slow or failing upload never blocks
// the next invocation's appends.
func (b *LogBuffer) Flush() error {
	b.mu.Lock()
	entries := b.entries
	b.entries = make([]LogEntry, 0, 1024)
	b.bytes = 0
	b.mu.Unlock()

	if len(entries) == 0 {
		return nil
	}
	return b.flushFn(entries)
}
