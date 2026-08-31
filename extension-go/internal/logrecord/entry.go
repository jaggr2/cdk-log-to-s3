// Package logrecord holds the on-disk log schema and the Parquet codec, shared
// by the telemetry extension that writes the files and the compactor that
// rewrites them. Both must agree on the schema byte for byte; keeping one
// definition is what guarantees that.
package logrecord

// Entry is one row in the Parquet output.
//
// The column names below are the public schema of this library. They are
// mirrored by src/private/schema.ts, which is what the Glue table is built
// from. Changing a name here is a breaking change for every existing Athena
// query, so the two must be edited together.
//
// Compression is deliberately NOT specified in these tags. Codec selection
// happens at the writer (see Write) so that LOG_TO_S3_COMPRESSION is an actual
// knob rather than a decorative one; a per-column tag would silently win over
// the writer option.
type Entry struct {
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

// OptStr reads an optional column, treating nil as empty.
func OptStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StrPtr writes an optional column, mapping empty to nil so the column is
// genuinely null rather than an empty string.
func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
