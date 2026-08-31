package main

import (
	"time"

	"github.com/jaggr2/cdk-log-to-s3/extension/internal/layout"
)

// The key layout lives in internal/layout so the compactor computes exactly
// the same partition paths this extension writes to. These are thin
// delegations; see that package for the reasoning behind the layout.

func NormalizePrefix(p string) string { return layout.NormalizePrefix(p) }

func PartitionPath(prefix string, day time.Time) string {
	return layout.PartitionPath(prefix, day)
}

func BuildKey(prefix string, now time.Time, functionName, correlationID string) string {
	return layout.BuildKey(prefix, now, functionName, correlationID)
}

func sanitize(s, fallback string) string { return layout.Sanitize(s, fallback) }
