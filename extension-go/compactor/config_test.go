package main

import (
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		envBucket, envPrefix, envCompression, envDebug,
		envMinFiles, envMaxFiles, envMaxBytes, envLookbackDays,
		"AWS_REGION",
	} {
		t.Setenv(k, "")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	clearEnv(t)

	cfg, warn := LoadConfig()
	if len(warn) != 0 {
		t.Fatalf("expected no warnings, got %v", warn)
	}
	if cfg.Prefix != "logs/" {
		t.Errorf("Prefix = %q", cfg.Prefix)
	}
	if cfg.Compression != "snappy" {
		t.Errorf("Compression = %q", cfg.Compression)
	}
	if cfg.MinFiles != 8 || cfg.MaxFiles != 2000 {
		t.Errorf("MinFiles=%d MaxFiles=%d", cfg.MinFiles, cfg.MaxFiles)
	}
	if cfg.MaxBytes != 256*1024*1024 {
		t.Errorf("MaxBytes = %d", cfg.MaxBytes)
	}
	if cfg.LookbackDays != 7 {
		t.Errorf("LookbackDays = %d", cfg.LookbackDays)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBucket, "my-bucket")
	t.Setenv(envPrefix, "/nested/logs")
	t.Setenv(envCompression, "ZSTD")
	t.Setenv(envMinFiles, "3")
	t.Setenv(envMaxFiles, "50")
	t.Setenv(envMaxBytes, "1048576")
	t.Setenv(envLookbackDays, "2")
	t.Setenv(envDebug, "true")

	cfg, warn := LoadConfig()
	if len(warn) != 0 {
		t.Fatalf("expected no warnings, got %v", warn)
	}
	if cfg.Bucket != "my-bucket" || cfg.Prefix != "nested/logs/" {
		t.Errorf("Bucket=%q Prefix=%q", cfg.Bucket, cfg.Prefix)
	}
	if cfg.Compression != "zstd" {
		t.Errorf("Compression = %q", cfg.Compression)
	}
	if cfg.MinFiles != 3 || cfg.MaxFiles != 50 || cfg.MaxBytes != 1048576 || cfg.LookbackDays != 2 {
		t.Errorf("caps not applied: %+v", cfg)
	}
	if !cfg.Debug {
		t.Error("Debug = false")
	}
}

func TestLoadConfigInvalidValuesWarn(t *testing.T) {
	clearEnv(t)
	t.Setenv(envCompression, "brotli-ish")
	t.Setenv(envMinFiles, "-1")
	t.Setenv(envMaxBytes, "lots")
	t.Setenv(envDebug, "sometimes")

	cfg, warn := LoadConfig()
	if len(warn) != 4 {
		t.Fatalf("expected 4 warnings, got %d: %v", len(warn), warn)
	}
	if cfg.Compression != "snappy" || cfg.MinFiles != 8 || cfg.MaxBytes != 256*1024*1024 || cfg.Debug {
		t.Errorf("defaults not preserved: %+v", cfg)
	}
}

// Today is excluded on purpose: the extension is still writing into it.
func TestPartitionsExcludeToday(t *testing.T) {
	cfg := &Config{
		Prefix:       "logs/",
		LookbackDays: 3,
		Now:          func() time.Time { return time.Date(2026, 3, 2, 4, 30, 0, 0, time.UTC) },
	}

	got := cfg.Partitions()
	want := []string{
		"logs/2026/03/01/",
		"logs/2026/02/28/",
		"logs/2026/02/27/",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d partitions, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("partition %d = %q, want %q", i, got[i], want[i])
		}
	}
	for _, p := range got {
		if p == "logs/2026/03/02/" {
			t.Error("today must not be compacted while it is still being written")
		}
	}
}

// A lookback longer than one day means a failed or skipped run catches up by
// itself rather than leaving a partition uncompacted forever.
func TestPartitionsCoverTheLookbackWindow(t *testing.T) {
	cfg := &Config{
		Prefix:       "",
		LookbackDays: 7,
		Now:          func() time.Time { return time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC) },
	}

	got := cfg.Partitions()
	if len(got) != 7 {
		t.Fatalf("got %d partitions, want 7", len(got))
	}
	if got[0] != "2026/01/02/" {
		t.Errorf("first = %q, want yesterday", got[0])
	}
	if got[6] != "2025/12/27/" {
		t.Errorf("last = %q, want the year boundary handled", got[6])
	}
}
