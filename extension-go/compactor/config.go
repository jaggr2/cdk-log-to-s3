package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jaggr2/cdk-log-to-s3/extension/internal/layout"
	"github.com/jaggr2/cdk-log-to-s3/extension/internal/logrecord"
)

const (
	envBucket       = "LOG_TO_S3_BUCKET"
	envPrefix       = "LOG_TO_S3_PREFIX"
	envCompression  = "LOG_TO_S3_COMPRESSION"
	envDebug        = "LOG_TO_S3_DEBUG"
	envMinFiles     = "LOG_TO_S3_COMPACTION_MIN_FILES"
	envMaxFiles     = "LOG_TO_S3_COMPACTION_MAX_FILES"
	envMaxBytes     = "LOG_TO_S3_COMPACTION_MAX_BYTES"
	envLookbackDays = "LOG_TO_S3_COMPACTION_LOOKBACK_DAYS"
)

type Config struct {
	Bucket      string
	Prefix      string
	Compression string

	// MinFiles is the point below which merging is not worth the read/write.
	MinFiles int
	// MaxFiles and MaxBytes bound one run. Anything left over is picked up on
	// the next run; every run still reduces the file count.
	MaxFiles int
	MaxBytes int64
	// LookbackDays is how many closed days to consider each run. More than one
	// so a failed or skipped run is caught up automatically.
	LookbackDays int

	Region string
	Debug  bool

	// Now is injected so the day window is testable.
	Now func() time.Time
}

func LoadConfig() (*Config, []string) {
	var warn []string

	cfg := &Config{
		Bucket:       os.Getenv(envBucket),
		Prefix:       layout.NormalizePrefix(getEnv(envPrefix, "logs/")),
		Compression:  "snappy",
		MinFiles:     8,
		MaxFiles:     2000,
		MaxBytes:     256 * 1024 * 1024,
		LookbackDays: 7,
		Region:       os.Getenv("AWS_REGION"),
		Now:          time.Now,
	}

	if raw := os.Getenv(envCompression); raw != "" {
		codec := strings.ToLower(strings.TrimSpace(raw))
		if logrecord.IsSupportedCompression(codec) {
			cfg.Compression = codec
		} else {
			warn = append(warn, fmt.Sprintf("%s=%q is not a supported codec, using snappy", envCompression, raw))
		}
	}

	cfg.MinFiles = envInt(envMinFiles, cfg.MinFiles, &warn)
	cfg.MaxFiles = envInt(envMaxFiles, cfg.MaxFiles, &warn)
	cfg.LookbackDays = envInt(envLookbackDays, cfg.LookbackDays, &warn)

	if raw := os.Getenv(envMaxBytes); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
			cfg.MaxBytes = n
		} else {
			warn = append(warn, fmt.Sprintf("%s=%q is not a positive integer, using %d", envMaxBytes, raw, cfg.MaxBytes))
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

// Partitions lists the day partitions to consider, most recent first.
//
// Today is deliberately excluded: the extension is still writing into it, and
// compacting a partition while new files land there would merge a moving
// target for no benefit - those files get compacted tomorrow anyway.
func (c *Config) Partitions() []string {
	today := c.Now().UTC().Truncate(24 * time.Hour)

	out := make([]string, 0, c.LookbackDays)
	for i := 1; i <= c.LookbackDays; i++ {
		out = append(out, layout.PartitionPath(c.Prefix, today.AddDate(0, 0, -i)))
	}
	return out
}

func envInt(key string, fallback int, warn *[]string) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	if n, err := strconv.Atoi(raw); err == nil && n > 0 {
		return n
	}
	*warn = append(*warn, fmt.Sprintf("%s=%q is not a positive integer, using %d", key, raw, fallback))
	return fallback
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
