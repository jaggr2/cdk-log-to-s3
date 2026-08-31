package main

import (
	"strings"
	"testing"
	"time"
)

// clearEnv removes every variable loadConfig looks at, so a test never
// inherits the developer's shell or another test's leftovers.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		envBucket, envPrefix, envLevel, envFlushSeconds, envMaxBufferBytes,
		envTelemetryPort, envCompression, envPlatformReport, envDebug,
		envBucketDeprecated, envLevelDeprecated,
		"AWS_REGION", "AWS_LAMBDA_FUNCTION_NAME",
	} {
		t.Setenv(k, "")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	clearEnv(t)

	cfg, warn := loadConfig()
	if len(warn) != 0 {
		t.Fatalf("expected no warnings, got %v", warn)
	}
	if cfg.Bucket != "" {
		t.Errorf("Bucket = %q, want empty", cfg.Bucket)
	}
	if cfg.Prefix != "logs/" {
		t.Errorf("Prefix = %q, want logs/", cfg.Prefix)
	}
	if cfg.Level != LevelInfo {
		t.Errorf("Level = %v, want INFO", cfg.Level)
	}
	if cfg.FlushInterval != 15*time.Second {
		t.Errorf("FlushInterval = %v, want 15s", cfg.FlushInterval)
	}
	if cfg.MaxBufferBytes != 10*1024*1024 {
		t.Errorf("MaxBufferBytes = %d, want 10485760", cfg.MaxBufferBytes)
	}
	if cfg.TelemetryPort != "2020" {
		t.Errorf("TelemetryPort = %q, want 2020", cfg.TelemetryPort)
	}
	if cfg.Compression != "snappy" {
		t.Errorf("Compression = %q, want snappy", cfg.Compression)
	}
	if !cfg.IncludePlatformReport {
		t.Error("IncludePlatformReport = false, want true")
	}
	if cfg.Debug {
		t.Error("Debug = true, want false")
	}
	// A library must never guess the region on the caller's behalf.
	if cfg.Region != "" {
		t.Errorf("Region = %q, want empty so the SDK chain resolves it", cfg.Region)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBucket, "my-bucket")
	t.Setenv(envPrefix, "/nested/logs")
	t.Setenv(envLevel, "warn")
	t.Setenv(envFlushSeconds, "5")
	t.Setenv(envMaxBufferBytes, "2048")
	t.Setenv(envTelemetryPort, "3000")
	t.Setenv(envCompression, "ZSTD")
	t.Setenv(envPlatformReport, "false")
	t.Setenv(envDebug, "true")
	t.Setenv("AWS_REGION", "eu-west-1")
	t.Setenv("AWS_LAMBDA_FUNCTION_NAME", "my-fn")

	cfg, warn := loadConfig()
	if len(warn) != 0 {
		t.Fatalf("expected no warnings, got %v", warn)
	}
	if cfg.Bucket != "my-bucket" {
		t.Errorf("Bucket = %q", cfg.Bucket)
	}
	if cfg.Prefix != "nested/logs/" {
		t.Errorf("Prefix = %q, want nested/logs/", cfg.Prefix)
	}
	if cfg.Level != LevelWarn {
		t.Errorf("Level = %v, want WARN", cfg.Level)
	}
	if cfg.FlushInterval != 5*time.Second {
		t.Errorf("FlushInterval = %v", cfg.FlushInterval)
	}
	if cfg.MaxBufferBytes != 2048 {
		t.Errorf("MaxBufferBytes = %d", cfg.MaxBufferBytes)
	}
	if cfg.TelemetryPort != "3000" {
		t.Errorf("TelemetryPort = %q", cfg.TelemetryPort)
	}
	if cfg.Compression != "zstd" {
		t.Errorf("Compression = %q, want zstd", cfg.Compression)
	}
	if cfg.IncludePlatformReport {
		t.Error("IncludePlatformReport = true, want false")
	}
	if !cfg.Debug {
		t.Error("Debug = false, want true")
	}
	if cfg.Region != "eu-west-1" {
		t.Errorf("Region = %q", cfg.Region)
	}
	if cfg.FunctionName != "my-fn" {
		t.Errorf("FunctionName = %q", cfg.FunctionName)
	}
}

// The pre-extraction names must keep working so an existing deployment can
// upgrade the layer without changing its function environment in the same step.
func TestLoadConfigDeprecatedFallbacks(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBucketDeprecated, "legacy-bucket")
	t.Setenv(envLevelDeprecated, "DEBUG")

	cfg, warn := loadConfig()
	if len(warn) != 0 {
		t.Fatalf("expected no warnings, got %v", warn)
	}
	if cfg.Bucket != "legacy-bucket" {
		t.Errorf("Bucket = %q, want legacy-bucket", cfg.Bucket)
	}
	if cfg.Level != LevelDebug {
		t.Errorf("Level = %v, want DEBUG", cfg.Level)
	}
}

func TestLoadConfigNewNameWinsOverDeprecated(t *testing.T) {
	clearEnv(t)
	t.Setenv(envBucket, "new-bucket")
	t.Setenv(envBucketDeprecated, "legacy-bucket")

	cfg, _ := loadConfig()
	if cfg.Bucket != "new-bucket" {
		t.Errorf("Bucket = %q, want new-bucket", cfg.Bucket)
	}
}

func TestLoadConfigInvalidValuesWarnAndFallBack(t *testing.T) {
	clearEnv(t)
	t.Setenv(envLevel, "LOUD")
	t.Setenv(envFlushSeconds, "-1")
	t.Setenv(envMaxBufferBytes, "lots")
	t.Setenv(envCompression, "brotli-ish")
	t.Setenv(envPlatformReport, "perhaps")
	t.Setenv(envDebug, "sometimes")

	cfg, warn := loadConfig()
	if len(warn) != 6 {
		t.Fatalf("expected 6 warnings, got %d: %v", len(warn), warn)
	}
	for _, key := range []string{
		envLevel, envFlushSeconds, envMaxBufferBytes,
		envCompression, envPlatformReport, envDebug,
	} {
		if !strings.Contains(strings.Join(warn, "\n"), key) {
			t.Errorf("no warning mentions %s", key)
		}
	}

	if cfg.Level != LevelInfo {
		t.Errorf("Level = %v, want INFO", cfg.Level)
	}
	if cfg.FlushInterval != 15*time.Second {
		t.Errorf("FlushInterval = %v, want the default", cfg.FlushInterval)
	}
	if cfg.MaxBufferBytes != 10*1024*1024 {
		t.Errorf("MaxBufferBytes = %d, want the default", cfg.MaxBufferBytes)
	}
	if cfg.Compression != "snappy" {
		t.Errorf("Compression = %q, want snappy", cfg.Compression)
	}
	if !cfg.IncludePlatformReport {
		t.Error("IncludePlatformReport should stay at its default")
	}
	if cfg.Debug {
		t.Error("Debug should stay at its default")
	}
}

func TestParseLevelOrdering(t *testing.T) {
	if !(LevelDebug < LevelInfo && LevelInfo < LevelWarn && LevelWarn < LevelError) {
		t.Fatal("levels must be ordered so a threshold comparison works")
	}
	cases := map[string]Level{
		"debug": LevelDebug, "TRACE": LevelDebug,
		"info": LevelInfo, "Warn": LevelWarn, "WARNING": LevelWarn,
		"error": LevelError, "FATAL": LevelError, "CRITICAL": LevelError,
	}
	for in, want := range cases {
		got, ok := ParseLevel(in)
		if !ok || got != want {
			t.Errorf("ParseLevel(%q) = %v/%v, want %v/true", in, got, ok, want)
		}
	}
	if _, ok := ParseLevel("chatty"); ok {
		t.Error("ParseLevel should report unknown levels as unrecognised")
	}
}
