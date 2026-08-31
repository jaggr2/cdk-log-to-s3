package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Level is an ordered log-level threshold. Records below the configured
// level are dropped before they ever reach the buffer, so they cost nothing
// beyond the CloudWatch ingest the runtime already paid for.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func ParseLevel(s string) (Level, bool) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG", "TRACE":
		return LevelDebug, true
	case "INFO":
		return LevelInfo, true
	case "WARN", "WARNING":
		return LevelWarn, true
	case "ERROR", "FATAL", "CRITICAL":
		return LevelError, true
	default:
		return LevelInfo, false
	}
}

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

const (
	// Default extension name. Must match the file name under /opt/extensions,
	// otherwise the Extensions API rejects the registration.
	defaultExtensionName = "log-to-s3"

	envBucket          = "LOG_TO_S3_BUCKET"
	envPrefix          = "LOG_TO_S3_PREFIX"
	envLevel           = "LOG_TO_S3_LEVEL"
	envFlushSeconds    = "LOG_TO_S3_FLUSH_INTERVAL_SECONDS"
	envMaxBufferBytes  = "LOG_TO_S3_MAX_BUFFER_BYTES"
	envTelemetryPort   = "LOG_TO_S3_TELEMETRY_PORT"
	envCompression     = "LOG_TO_S3_COMPRESSION"
	envPlatformReport  = "LOG_TO_S3_INCLUDE_PLATFORM_REPORT"
	envDebug           = "LOG_TO_S3_DEBUG"

	// Names used by the pre-extraction version of this extension. Honoured so
	// an in-place upgrade does not need a simultaneous env-var change.
	envBucketDeprecated = "APP_LOGS_BUCKET"
	envLevelDeprecated  = "LOG_LEVEL"
)

type Config struct {
	Bucket                string
	Prefix                string
	Level                 Level
	FlushInterval         time.Duration
	MaxBufferBytes        int
	TelemetryPort         string
	Compression           string
	IncludePlatformReport bool
	Debug                 bool

	// Resolved once at startup rather than per record.
	FunctionName  string
	Region        string
	ExtensionName string
}

// loadConfig reads the environment and returns the resolved configuration
// together with any warnings. Warnings are logged once at startup instead of
// being swallowed, so a typo in a construct prop is visible in CloudWatch.
func loadConfig() (*Config, []string) {
	var warn []string

	cfg := &Config{
		Bucket:                firstNonEmpty(os.Getenv(envBucket), os.Getenv(envBucketDeprecated)),
		Prefix:                NormalizePrefix(getEnv(envPrefix, "logs/")),
		FlushInterval:         15 * time.Second,
		MaxBufferBytes:        10 * 1024 * 1024,
		TelemetryPort:         getEnv(envTelemetryPort, "2020"),
		Compression:           "snappy",
		IncludePlatformReport: true,
		FunctionName:          os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		// No fallback region. A library must not guess the caller's region;
		// when this is empty the SDK's own resolution chain takes over.
		Region:        os.Getenv("AWS_REGION"),
		ExtensionName: defaultExtensionName,
	}

	levelRaw := firstNonEmpty(os.Getenv(envLevel), os.Getenv(envLevelDeprecated))
	if levelRaw == "" {
		cfg.Level = LevelInfo
	} else if lvl, ok := ParseLevel(levelRaw); ok {
		cfg.Level = lvl
	} else {
		cfg.Level = LevelInfo
		warn = append(warn, fmt.Sprintf("%s=%q is not a known level, using INFO", envLevel, levelRaw))
	}

	if raw := os.Getenv(envFlushSeconds); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			cfg.FlushInterval = time.Duration(n) * time.Second
		} else {
			warn = append(warn, fmt.Sprintf("%s=%q is not a positive integer, using %s", envFlushSeconds, raw, cfg.FlushInterval))
		}
	}

	if raw := os.Getenv(envMaxBufferBytes); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			cfg.MaxBufferBytes = n
		} else {
			warn = append(warn, fmt.Sprintf("%s=%q is not a positive integer, using %d", envMaxBufferBytes, raw, cfg.MaxBufferBytes))
		}
	}

	if raw := os.Getenv(envCompression); raw != "" {
		codec := strings.ToLower(strings.TrimSpace(raw))
		if IsSupportedCompression(codec) {
			cfg.Compression = codec
		} else {
			warn = append(warn, fmt.Sprintf("%s=%q is not a supported codec, using snappy", envCompression, raw))
		}
	}

	if raw := os.Getenv(envPlatformReport); raw != "" {
		if b, err := strconv.ParseBool(raw); err == nil {
			cfg.IncludePlatformReport = b
		} else {
			warn = append(warn, fmt.Sprintf("%s=%q is not a boolean, using true", envPlatformReport, raw))
		}
	}

	if raw := os.Getenv(envDebug); raw != "" {
		if b, err := strconv.ParseBool(raw); err == nil {
			cfg.Debug = b
		} else {
			warn = append(warn, fmt.Sprintf("%s=%q is not a boolean, using false", envDebug, raw))
		}
	}

	return cfg, warn
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
