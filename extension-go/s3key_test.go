package main

import (
	"regexp"
	"testing"
	"time"
)

// keyRe is the layout Athena partition projection depends on: the prefix, the
// yyyy/MM/dd partition directories, then the file. No other directory level
// may appear, which is why the function name lives in the file name.
var keyRe = regexp.MustCompile(
	`^(?:[^/=]+/)*\d{4}/\d{2}/\d{2}/[A-Za-z0-9._-]+\.parquet$`)

func TestNormalizePrefix(t *testing.T) {
	cases := map[string]string{
		"":            "",
		"/":           "",
		"logs":        "logs/",
		"logs/":       "logs/",
		"/logs/":      "logs/",
		"//logs//":    "logs/",
		"  logs/  ":   "logs/",
		"a/b":         "a/b/",
		"/nested/dir": "nested/dir/",
	}
	for in, want := range cases {
		if got := NormalizePrefix(in); got != want {
			t.Errorf("NormalizePrefix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildKeyLayout(t *testing.T) {
	now := time.Date(2026, 8, 5, 9, 30, 0, 0, time.UTC)
	key := BuildKey("logs/", now, "my-fn", "cid-1")

	if !keyRe.MatchString(key) {
		t.Fatalf("key %q does not match the projection-compatible layout", key)
	}
	wantDir := "logs/2026/08/05/"
	if key[:len(wantDir)] != wantDir {
		t.Errorf("key = %q, want it to start with %q", key, wantDir)
	}
	// Zero padding is what lets the yyyy/MM/dd values sort lexicographically
	// in the same order they sort chronologically.
	if want := "/08/05/"; !contains(key, want) {
		t.Errorf("key = %q, want zero-padded %q", key, want)
	}
}

// The partition value is what Athena projects into ${dt}; the extension and
// the table must agree on it exactly.
func TestPartitionPath(t *testing.T) {
	day := time.Date(2026, 1, 2, 23, 59, 0, 0, time.UTC)

	if got, want := PartitionPath("logs/", day), "logs/2026/01/02/"; got != want {
		t.Errorf("PartitionPath = %q, want %q", got, want)
	}
	if got, want := PartitionPath("", day), "2026/01/02/"; got != want {
		t.Errorf("PartitionPath = %q, want %q", got, want)
	}
	if got, want := PartitionPath("/nested/dir", day), "nested/dir/2026/01/02/"; got != want {
		t.Errorf("PartitionPath = %q, want %q", got, want)
	}
}

// Keys within a partition sort lexicographically in date order, which is what
// makes range predicates on the dt column correct rather than merely handy.
func TestPartitionPathsSortChronologically(t *testing.T) {
	days := []time.Time{
		time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	prev := ""
	for _, d := range days {
		p := PartitionPath("logs/", d)
		if prev != "" && !(prev < p) {
			t.Errorf("%q should sort before %q", prev, p)
		}
		prev = p
	}
}

func TestBuildKeyHasNoExtraDirectoryLevels(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 0, 0, 0, time.UTC)
	key := BuildKey("logs/", now, "my-fn", "cid-1")

	// The pre-extraction layout inserted "extension/<functionName>/" below the
	// partitions, which cannot be expressed as a location template.
	if contains(key, "/extension/") {
		t.Errorf("key = %q still contains an extension/ segment", key)
	}
	if count(key, "/") != 4 {
		t.Errorf("key = %q has %d slashes, want 4 (prefix + yyyy/MM/dd)", key, count(key, "/"))
	}
}

func TestBuildKeyEmptyPrefix(t *testing.T) {
	now := time.Date(2026, 12, 31, 23, 0, 0, 0, time.UTC)
	key := BuildKey("", now, "fn", "cid")

	if key[:5] != "2026/" {
		t.Errorf("key = %q, want it to start at the year segment", key)
	}
	if !keyRe.MatchString(key) {
		t.Errorf("key %q does not match the expected layout", key)
	}
}

func TestBuildKeyConvertsToUTC(t *testing.T) {
	zone := time.FixedZone("UTC+5", 5*3600)
	local := time.Date(2026, 3, 1, 2, 0, 0, 0, zone) // 2026-02-28T21:00Z

	key := BuildKey("logs/", local, "fn", "cid")
	want := "logs/2026/02/28/"
	if key[:len(want)] != want {
		t.Errorf("key = %q, want it to start with %q", key, want)
	}
}

func TestBuildKeySanitizesUnsafeInput(t *testing.T) {
	now := time.Date(2026, 8, 5, 9, 0, 0, 0, time.UTC)
	key := BuildKey("logs/", now, "my/fn name:v1", "cid with spaces")

	if !keyRe.MatchString(key) {
		t.Fatalf("key %q was not sanitised into a single path segment", key)
	}
	if contains(key, " ") || contains(key, ":") {
		t.Errorf("key = %q still contains unsafe characters", key)
	}
}

func TestBuildKeyFallsBackForEmptyIdentifiers(t *testing.T) {
	now := time.Date(2026, 8, 5, 9, 0, 0, 0, time.UTC)
	key := BuildKey("logs/", now, "", "")

	if !contains(key, "unknown-none-") {
		t.Errorf("key = %q, want the unknown/none fallbacks", key)
	}
	if !keyRe.MatchString(key) {
		t.Errorf("key %q does not match the expected layout", key)
	}
}

// Two flushes inside one invocation share a correlation id. Before the random
// suffix they produced the same key and the second overwrote the first.
func TestBuildKeyIsUniquePerCall(t *testing.T) {
	now := time.Date(2026, 8, 5, 9, 0, 0, 0, time.UTC)
	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		key := BuildKey("logs/", now, "my-fn", "same-correlation-id")
		if seen[key] {
			t.Fatalf("duplicate key generated: %q", key)
		}
		seen[key] = true
	}
}

func TestSanitizeTruncates(t *testing.T) {
	long := ""
	for i := 0; i < 200; i++ {
		long += "a"
	}
	if got := len(sanitize(long, "fallback")); got > 64 {
		t.Errorf("sanitize returned %d chars, want at most 64", got)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && indexOf(s, sub) >= 0 }

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func count(s, sub string) int {
	n := 0
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			n++
		}
	}
	return n
}
