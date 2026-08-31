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

// BuildKey lays out one Parquet file in a fully Hive-partitioned path.
//
// Every directory level below the prefix is a partition key and nothing else.
// That constraint comes from Athena partition projection: storage.location.template
// must contain a placeholder for every partition column, and static or dynamic
// segments below the last placeholder cannot be expressed. The function name
// therefore lives in the file name (and in the function_name column), not in a
// directory of its own.
//
// The random suffix is always present. Without it, two flushes within one
// invocation - the interval ticker, the size threshold and the runtimeDone
// flush can all fire for the same request - would produce identical keys and
// the second would silently overwrite the first.
func BuildKey(prefix string, now time.Time, functionName, correlationID string) string {
	now = now.UTC()
	return fmt.Sprintf("%syear=%04d/month=%02d/day=%02d/hour=%02d/%s-%s-%s.parquet",
		NormalizePrefix(prefix),
		now.Year(), int(now.Month()), now.Day(), now.Hour(),
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
