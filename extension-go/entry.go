package main

// LogEntry is one row in the Parquet output.
//
// The column names below are the public schema of this library: they are
// mirrored by src/private/schema.ts, which is what the Glue table is built
// from. Changing a name here is a breaking change for every existing Athena
// query, so the two must be edited together.
//
// Compression is deliberately NOT specified in these tags. Codec selection
// happens at the writer (see writer.go) so that LOG_TO_S3_COMPRESSION is an
// actual knob rather than a decorative one; a per-column tag would silently
// win over the writer option.
type LogEntry struct {
	Timestamp     string  `parquet:"timestamp"`
	Level         string  `parquet:"level"`
	Source        string  `parquet:"source"`
	CorrelationID string  `parquet:"correlation_id"`
	RequestID     string  `parquet:"request_id"`
	Message       string  `parquet:"message"`
	FunctionName  string  `parquet:"function_name"`
	Context       *string `parquet:"context,optional"`
	StackTrace    *string `parquet:"stack_trace,optional"`
	Caller        *string `parquet:"caller,optional"`
}

// estimateSize approximates the in-memory footprint of an entry. It only has
// to be monotonic and roughly proportional - it drives the flush threshold,
// not any allocation.
func estimateSize(e LogEntry) int {
	return len(e.Timestamp) + len(e.Level) + len(e.Source) +
		len(e.CorrelationID) + len(e.RequestID) + len(e.Message) +
		len(e.FunctionName) + len(optStr(e.Context)) +
		len(optStr(e.StackTrace)) + len(optStr(e.Caller)) + 200
}

func optStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
