package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// NormalizePrefix returns a prefix that is either empty or ends in exactly one
// slash, with any leading slashes removed. S3 keys have no leading slash, and
// the Glue storage.location.template built by the CDK side assumes the same
// normalisation - a mismatch produces a table that silently returns no rows.
func NormalizePrefix(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimLeft(p, "/")
	p = strings.TrimRight(p, "/")
	if p == "" {
		return ""
	}
	return p + "/"
}

// PartitionPath returns the directory a given day's objects live in, e.g.
// "logs/2026/08/31/". It is the value the Athena `dt` partition column takes,
// prefixed, and the unit the compactor operates on.
func PartitionPath(prefix string, day time.Time) string {
	day = day.UTC()
	return fmt.Sprintf("%s%04d/%02d/%02d/",
		NormalizePrefix(prefix), day.Year(), int(day.Month()), day.Day())
}

// BuildKey lays out one Parquet file inside its day partition.
//
// Every directory level below the prefix belongs to the partition and nothing
// else. That constraint comes from Athena partition projection:
// storage.location.template must contain a placeholder for every partition
// column, and static or dynamic segments below the last placeholder cannot be
// expressed. The function name therefore lives in the file name (and in the
// function_name column), not in a directory of its own.
//
// The layout is yyyy/MM/dd, matching projection.dt.format. Day granularity
// rather than hourly keeps the projected partition count near 730 over a
// two-year window instead of tens of thousands.
//
// The random suffix is always present. Without it, two flushes within one
// invocation - the interval ticker, the size threshold and the runtimeDone
// flush can all fire for the same request - would produce identical keys and
// the second would silently overwrite the first.
func BuildKey(prefix string, now time.Time, functionName, correlationID string) string {
	return fmt.Sprintf("%s%s-%s-%s.parquet",
		PartitionPath(prefix, now),
		sanitize(functionName, "unknown"),
		sanitize(correlationID, "none"),
		randomSuffix(),
	)
}

// sanitize keeps the key printable and free of path separators.
func sanitize(s, fallback string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
		if b.Len() >= 64 {
			break
		}
	}
	out := b.String()
	if out == "" {
		return fallback
	}
	return out
}

func randomSuffix() string {
	var buf [6]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// crypto/rand failing is not a reason to drop logs; fall back to the
		// clock, which is still unique enough to avoid an overwrite.
		return fmt.Sprintf("%012x", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}
