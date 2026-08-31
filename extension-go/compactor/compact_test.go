package main

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jaggr2/cdk-log-to-s3/extension/internal/logrecord"
)

// fakeStore is an in-memory ObjectStore with a hook for simulating a crash at
// any point in the protocol.
type fakeStore struct {
	objects map[string][]byte
	// hook returns an error to abort an operation. For Delete it is consulted
	// per key, so a run can be interrupted partway through a batch.
	hook func(op, key string) error
}

func newFakeStore() *fakeStore {
	return &fakeStore{objects: map[string][]byte{}}
}

func (f *fakeStore) List(_ context.Context, prefix string) ([]Object, error) {
	var out []Object
	for k, v := range f.objects {
		if strings.HasPrefix(k, prefix) {
			out = append(out, Object{Key: k, Size: int64(len(v))})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

func (f *fakeStore) Get(_ context.Context, key string) ([]byte, error) {
	if f.hook != nil {
		if err := f.hook("get", key); err != nil {
			return nil, err
		}
	}
	data, ok := f.objects[key]
	if !ok {
		return nil, fmt.Errorf("no such key: %s", key)
	}
	return data, nil
}

func (f *fakeStore) Put(_ context.Context, key string, body []byte) error {
	if f.hook != nil {
		if err := f.hook("put", key); err != nil {
			return err
		}
	}
	f.objects[key] = body
	return nil
}

func (f *fakeStore) Delete(_ context.Context, keys []string) error {
	for _, k := range keys {
		if f.hook != nil {
			if err := f.hook("delete", k); err != nil {
				return err
			}
		}
		delete(f.objects, k)
	}
	return nil
}

func testConfig() *Config {
	return &Config{
		Prefix:       "logs/",
		Compression:  "snappy",
		MinFiles:     3,
		MaxFiles:     2000,
		MaxBytes:     256 * 1024 * 1024,
		LookbackDays: 7,
		Now:          func() time.Time { return time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC) },
	}
}

// seed writes `count` single-row Parquet objects into a partition.
func seed(t *testing.T, f *fakeStore, partition string, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		entry := logrecord.Entry{
			Timestamp: fmt.Sprintf("2026-08-30T%02d:00:00Z", i%24),
			Level:     "INFO",
			Source:    "Test",
			Message:   fmt.Sprintf("row %d", i),
			RequestID: fmt.Sprintf("req-%03d", i),
		}
		data, err := logrecord.Write([]logrecord.Entry{entry}, "snappy")
		if err != nil {
			t.Fatal(err)
		}
		f.objects[fmt.Sprintf("%sfn-cid-%03d.parquet", partition, i)] = data
	}
}

// totalRows reads every Parquet object Athena would read under a prefix, which
// means skipping the underscore-prefixed files Hive treats as hidden.
// Duplicated rows show up here as an inflated count, which is what the crash
// tests check.
func totalRows(t *testing.T, f *fakeStore, prefix string) int {
	t.Helper()
	total := 0
	for k, v := range f.objects {
		base := path.Base(k)
		if !strings.HasPrefix(k, prefix) || !strings.HasSuffix(base, ".parquet") {
			continue
		}
		if strings.HasPrefix(base, "_") || strings.HasPrefix(base, ".") {
			continue
		}
		rows, err := logrecord.Read(v)
		if err != nil {
			t.Fatalf("%s is not readable Parquet: %v", k, err)
		}
		total += len(rows)
	}
	return total
}

func countKeys(f *fakeStore, prefix, contains string) int {
	n := 0
	for k := range f.objects {
		if strings.HasPrefix(k, prefix) && strings.Contains(k, contains) {
			n++
		}
	}
	return n
}

const partition = "logs/2026/08/30/"

func TestCompactMergesAndDeletesSources(t *testing.T) {
	f := newFakeStore()
	seed(t, f, partition, 10)

	res, err := CompactPartition(context.Background(), f, testConfig(), partition)
	if err != nil {
		t.Fatal(err)
	}

	if res.Skipped {
		t.Fatalf("unexpectedly skipped: %s", res.SkippedBecause)
	}
	if res.SourceFiles != 10 || res.Rows != 10 {
		t.Errorf("SourceFiles=%d Rows=%d, want 10/10", res.SourceFiles, res.Rows)
	}
	if len(f.objects) != 1 {
		t.Fatalf("partition holds %d objects, want exactly the merged one: %v", len(f.objects), keysOf(f))
	}
	if _, ok := f.objects[res.OutputKey]; !ok {
		t.Errorf("output %s missing", res.OutputKey)
	}
	if !strings.HasPrefix(res.OutputKey, partition+"part-") {
		t.Errorf("output key %q should be a part- file inside the partition", res.OutputKey)
	}
	if totalRows(t, f, partition) != 10 {
		t.Errorf("row count changed: %d, want 10", totalRows(t, f, partition))
	}
	// No bookkeeping left behind.
	if countKeys(f, partition, ".manifest.json") != 0 {
		t.Error("a manifest survived a successful run")
	}
}

func TestCompactSortsByTimestamp(t *testing.T) {
	f := newFakeStore()
	seed(t, f, partition, 10)

	res, err := CompactPartition(context.Background(), f, testConfig(), partition)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := logrecord.Read(f.objects[res.OutputKey])
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(rows); i++ {
		if rows[i-1].Timestamp > rows[i].Timestamp {
			t.Fatalf("rows are not sorted: %q before %q", rows[i-1].Timestamp, rows[i].Timestamp)
		}
	}
}

func TestCompactSkipsBelowMinFiles(t *testing.T) {
	f := newFakeStore()
	seed(t, f, partition, 2)

	res, err := CompactPartition(context.Background(), f, testConfig(), partition)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Skipped {
		t.Error("should have skipped a partition below MinFiles")
	}
	if len(f.objects) != 2 {
		t.Errorf("a skipped partition was modified: %v", keysOf(f))
	}
}

// A previous merge must never be fed back into the next one, or its rows would
// be duplicated on every run.
func TestCompactIgnoresAlreadyCompactedAndHiddenFiles(t *testing.T) {
	f := newFakeStore()
	seed(t, f, partition, 4)

	prior, err := logrecord.Write([]logrecord.Entry{{Timestamp: "2026-08-30T00:00:00Z", Message: "prior"}}, "snappy")
	if err != nil {
		t.Fatal(err)
	}
	f.objects[partition+"part-deadbeef.parquet"] = prior
	f.objects[partition+"_something.parquet"] = []byte("hidden, not parquet")
	f.objects[partition+"notes.txt"] = []byte("not parquet")

	res, err := CompactPartition(context.Background(), f, testConfig(), partition)
	if err != nil {
		t.Fatal(err)
	}

	if res.SourceFiles != 4 {
		t.Errorf("SourceFiles=%d, want only the 4 uncompacted files", res.SourceFiles)
	}
	if _, ok := f.objects[partition+"part-deadbeef.parquet"]; !ok {
		t.Error("the earlier merge was consumed; its rows would now be duplicated")
	}
	if _, ok := f.objects[partition+"notes.txt"]; !ok {
		t.Error("a non-Parquet file was deleted")
	}
	// 4 merged + 1 prior row, counted across both part files.
	if got := totalRows(t, f, partition); got != 5 {
		t.Errorf("total rows = %d, want 5", got)
	}
}

func TestCompactOutputKeyIsDeterministic(t *testing.T) {
	runOnce := func() string {
		f := newFakeStore()
		seed(t, f, partition, 5)
		res, err := CompactPartition(context.Background(), f, testConfig(), partition)
		if err != nil {
			t.Fatal(err)
		}
		return res.OutputKey
	}
	if a, b := runOnce(), runOnce(); a != b {
		t.Errorf("output key is not deterministic: %q vs %q", a, b)
	}
}

func TestCompactRespectsCaps(t *testing.T) {
	f := newFakeStore()
	seed(t, f, partition, 20)

	cfg := testConfig()
	cfg.MaxFiles = 5

	res, err := CompactPartition(context.Background(), f, cfg, partition)
	if err != nil {
		t.Fatal(err)
	}
	if res.SourceFiles != 5 {
		t.Errorf("SourceFiles=%d, want the MaxFiles cap of 5", res.SourceFiles)
	}
	// The remainder survives for the next run; the file count still dropped.
	if got := totalRows(t, f, partition); got != 20 {
		t.Errorf("total rows = %d, want 20", got)
	}
	if len(f.objects) != 16 {
		t.Errorf("objects = %d, want 15 leftovers + 1 merged", len(f.objects))
	}
}

// --- crash safety ----------------------------------------------------------
//
// The protocol is: write output, write manifest, delete sources, delete
// manifest. Each of these tests kills a run at one of those seams and checks
// that a second run converges on the correct state without duplicating rows.

func TestCrashAfterOutputBeforeManifest(t *testing.T) {
	f := newFakeStore()
	seed(t, f, partition, 6)

	boom := errors.New("boom")
	f.hook = func(op, key string) error {
		if op == "put" && strings.Contains(key, ".manifest.json") {
			return boom
		}
		return nil
	}
	if _, err := CompactPartition(context.Background(), f, testConfig(), partition); !errors.Is(err, boom) {
		t.Fatalf("expected the injected failure, got %v", err)
	}

	// Sources intact, an orphan output present.
	if countKeys(f, partition, "part-") != 1 {
		t.Fatal("expected one orphaned output")
	}

	f.hook = nil
	if _, err := CompactPartition(context.Background(), f, testConfig(), partition); err != nil {
		t.Fatal(err)
	}

	if got := totalRows(t, f, partition); got != 6 {
		t.Errorf("total rows = %d, want 6 - the orphaned output was not reconciled", got)
	}
	if len(f.objects) != 1 {
		t.Errorf("objects = %d, want 1: %v", len(f.objects), keysOf(f))
	}
}

func TestCrashAfterManifestBeforeDelete(t *testing.T) {
	f := newFakeStore()
	seed(t, f, partition, 6)

	boom := errors.New("boom")
	f.hook = func(op, key string) error {
		if op == "delete" && strings.HasSuffix(key, ".parquet") {
			return boom
		}
		return nil
	}
	if _, err := CompactPartition(context.Background(), f, testConfig(), partition); !errors.Is(err, boom) {
		t.Fatalf("expected the injected failure, got %v", err)
	}

	// Output and manifest written, every source still present: the moment at
	// which a naive implementation would double-count on the next run.
	if countKeys(f, partition, ".manifest.json") != 1 {
		t.Fatal("expected a manifest to be left behind")
	}
	if got := totalRows(t, f, partition); got != 12 {
		t.Fatalf("expected 12 rows mid-crash (6 sources + 6 merged), got %d", got)
	}

	f.hook = nil
	if _, err := CompactPartition(context.Background(), f, testConfig(), partition); err != nil {
		t.Fatal(err)
	}

	if got := totalRows(t, f, partition); got != 6 {
		t.Errorf("total rows = %d, want 6 - rows were duplicated", got)
	}
	if countKeys(f, partition, ".manifest.json") != 0 {
		t.Error("manifest survived recovery")
	}
	if len(f.objects) != 1 {
		t.Errorf("objects = %d, want 1: %v", len(f.objects), keysOf(f))
	}
}

func TestCrashPartwayThroughDelete(t *testing.T) {
	f := newFakeStore()
	seed(t, f, partition, 6)

	boom := errors.New("boom")
	deleted := 0
	f.hook = func(op, key string) error {
		if op == "delete" && strings.HasSuffix(key, ".parquet") {
			deleted++
			if deleted > 3 {
				return boom
			}
		}
		return nil
	}
	if _, err := CompactPartition(context.Background(), f, testConfig(), partition); !errors.Is(err, boom) {
		t.Fatalf("expected the injected failure, got %v", err)
	}

	// Half the sources are gone; the merged output holds all six rows. Without
	// the manifest the next run would merge the survivors again and duplicate
	// them.
	if got := totalRows(t, f, partition); got != 9 {
		t.Fatalf("expected 9 rows mid-crash (3 survivors + 6 merged), got %d", got)
	}

	f.hook = nil
	if _, err := CompactPartition(context.Background(), f, testConfig(), partition); err != nil {
		t.Fatal(err)
	}

	if got := totalRows(t, f, partition); got != 6 {
		t.Errorf("total rows = %d, want 6 - the partial delete was not reconciled", got)
	}
	if len(f.objects) != 1 {
		t.Errorf("objects = %d, want 1: %v", len(f.objects), keysOf(f))
	}
}

// New files arriving between a crash and the recovery run must not confuse it.
func TestRecoveryWithNewArrivals(t *testing.T) {
	f := newFakeStore()
	seed(t, f, partition, 6)

	boom := errors.New("boom")
	f.hook = func(op, key string) error {
		if op == "delete" && strings.HasSuffix(key, ".parquet") {
			return boom
		}
		return nil
	}
	if _, err := CompactPartition(context.Background(), f, testConfig(), partition); !errors.Is(err, boom) {
		t.Fatalf("expected the injected failure, got %v", err)
	}
	f.hook = nil

	// Four late arrivals land before the next run.
	for i := 100; i < 104; i++ {
		entry := logrecord.Entry{Timestamp: "2026-08-30T23:00:00Z", Message: fmt.Sprintf("late %d", i)}
		data, err := logrecord.Write([]logrecord.Entry{entry}, "snappy")
		if err != nil {
			t.Fatal(err)
		}
		f.objects[fmt.Sprintf("%sfn-cid-%03d.parquet", partition, i)] = data
	}

	if _, err := CompactPartition(context.Background(), f, testConfig(), partition); err != nil {
		t.Fatal(err)
	}

	// 6 original + 4 late, each exactly once.
	if got := totalRows(t, f, partition); got != 10 {
		t.Errorf("total rows = %d, want 10", got)
	}
	if countKeys(f, partition, ".manifest.json") != 0 {
		t.Error("a manifest survived")
	}
}

func TestCompactEmptyPartition(t *testing.T) {
	f := newFakeStore()
	res, err := CompactPartition(context.Background(), f, testConfig(), partition)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Skipped {
		t.Error("an empty partition should be skipped")
	}
}

func keysOf(f *fakeStore) []string {
	var out []string
	for k := range f.objects {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
