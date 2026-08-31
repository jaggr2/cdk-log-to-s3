package main

import (
	"errors"
	"sync"
	"testing"
)

func TestBufferAppendAndFlush(t *testing.T) {
	var got []LogEntry
	b := NewLogBuffer(1024, func(e []LogEntry) error {
		got = append(got, e...)
		return nil
	})

	b.Append(LogEntry{Message: "one"})
	b.Append(LogEntry{Message: "two"})

	if b.Len() != 2 {
		t.Fatalf("Len = %d, want 2", b.Len())
	}
	if b.SizeBytes() == 0 {
		t.Fatal("SizeBytes = 0 after appending")
	}

	if err := b.Flush(); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("flushed %d entries, want 2", len(got))
	}
	if b.Len() != 0 || b.SizeBytes() != 0 {
		t.Errorf("buffer not reset: len=%d bytes=%d", b.Len(), b.SizeBytes())
	}
}

func TestBufferEmptyFlushDoesNotCallFlushFn(t *testing.T) {
	called := false
	b := NewLogBuffer(1024, func([]LogEntry) error {
		called = true
		return nil
	})

	if err := b.Flush(); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Error("flushFn was called for an empty buffer")
	}
}

func TestBufferFlushErrorPropagates(t *testing.T) {
	want := errors.New("upload failed")
	b := NewLogBuffer(1024, func([]LogEntry) error { return want })

	b.Append(LogEntry{Message: "x"})
	if err := b.Flush(); !errors.Is(err, want) {
		t.Fatalf("Flush() = %v, want %v", err, want)
	}
	// The entries are dropped either way; what matters is that the buffer is
	// clear so the next invocation is not billed for the failed batch again.
	if b.Len() != 0 {
		t.Errorf("Len = %d after a failed flush, want 0", b.Len())
	}
}

func TestBufferFull(t *testing.T) {
	b := NewLogBuffer(500, func([]LogEntry) error { return nil })

	if b.Full() {
		t.Error("empty buffer reported as full")
	}
	// estimateSize adds 200 bytes of overhead per entry.
	b.Append(LogEntry{Message: "a"})
	b.Append(LogEntry{Message: "b"})
	b.Append(LogEntry{Message: "c"})
	if !b.Full() {
		t.Errorf("buffer with %d bytes should be full at 500", b.SizeBytes())
	}
}

func TestBufferConcurrentAppend(t *testing.T) {
	b := NewLogBuffer(1<<30, func([]LogEntry) error { return nil })

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				b.Append(LogEntry{Message: "concurrent"})
			}
		}()
	}
	wg.Wait()

	if b.Len() != 1600 {
		t.Fatalf("Len = %d, want 1600", b.Len())
	}
}

// Flush resets the buffer before flushFn runs, so appends made during a slow
// upload belong to the next batch instead of being lost.
func TestBufferAppendDuringFlush(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	flushed := make(chan int, 1)

	b := NewLogBuffer(1<<30, func(e []LogEntry) error {
		close(entered)
		<-release
		flushed <- len(e)
		return nil
	})

	b.Append(LogEntry{Message: "first"})

	done := make(chan error, 1)
	go func() { done <- b.Flush() }()

	// Only append once the flush is demonstrably past the buffer swap.
	<-entered
	b.Append(LogEntry{Message: "second"})
	close(release)

	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if n := <-flushed; n != 1 {
		t.Errorf("flushed %d entries, want 1", n)
	}
	if b.Len() != 1 {
		t.Errorf("Len = %d, want the entry appended during the flush to survive", b.Len())
	}
}
