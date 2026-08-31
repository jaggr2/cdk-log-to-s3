package main

import "github.com/jaggr2/cdk-log-to-s3/extension/internal/logrecord"

// LogEntry is the row written to Parquet. It is a type alias, not a new type,
// so the extension and the compactor share one schema definition and cannot
// drift apart. See internal/logrecord for the column names and their contract
// with src/private/schema.ts.
type LogEntry = logrecord.Entry

// estimateSize approximates the in-memory footprint of an entry. It only has
// to be monotonic and roughly proportional - it drives the flush threshold,
// not any allocation.
func estimateSize(e LogEntry) int {
	return len(e.Timestamp) + len(e.Level) + len(e.Source) +
		len(e.CorrelationID) + len(e.RequestID) + len(e.Message) +
		len(e.FunctionName) + len(optStr(e.Context)) +
		len(optStr(e.StackTrace)) + len(optStr(e.Caller)) + 200
}

func optStr(s *string) string { return logrecord.OptStr(s) }

func strPtr(s string) *string { return logrecord.StrPtr(s) }
