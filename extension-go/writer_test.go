package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/format"
)

// wantColumns is the public schema of this library. It must stay identical to
// the column list in src/private/schema.ts, which is what the Glue table is
// generated from; a mismatch produces a table full of nulls with no error
// anywhere.
var wantColumns = []string{
	"timestamp",
	"level",
	"source",
	"correlation_id",
	"request_id",
	"message",
	"function_name",
	"context",
	"stack_trace",
	"caller",
}

func sampleEntries() []LogEntry {
	ctx := `{"count":5}`
	stack := "Error: boom\n    at handler"
	caller := "src/lib/sync.ts:142:11"
	return []LogEntry{
		{
			Timestamp: "2026-08-31T14:05:00.123Z", Level: "INFO", Source: "SyncProcessor",
			CorrelationID: "cid-1", RequestID: "req-1", Message: "Events synced",
			FunctionName: "my-fn", Context: &ctx, Caller: &caller,
		},
		{
			Timestamp: "2026-08-31T14:05:01.000Z", Level: "ERROR", Source: "Api",
			CorrelationID: "cid-1", RequestID: "req-1", Message: "boom",
			FunctionName: "my-fn", StackTrace: &stack,
		},
	}
}

func TestSchemaColumnNames(t *testing.T) {
	schema := parquet.SchemaOf(LogEntry{})

	var got []string
	for _, f := range schema.Fields() {
		got = append(got, f.Name())
	}

	if len(got) != len(wantColumns) {
		t.Fatalf("schema has %d columns %v, want %d %v", len(got), got, len(wantColumns), wantColumns)
	}
	for i := range wantColumns {
		if got[i] != wantColumns[i] {
			t.Errorf("column %d = %q, want %q", i, got[i], wantColumns[i])
		}
	}

	// S3Path was always null and is gone; it must not creep back in.
	for _, name := range got {
		if strings.Contains(strings.ToLower(name), "s3") {
			t.Errorf("unexpected column %q", name)
		}
	}
}

func TestWriteParquetRoundTrip(t *testing.T) {
	entries := sampleEntries()

	data, err := WriteParquet(entries, "snappy")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.HasPrefix(data, []byte("PAR1")) || !bytes.HasSuffix(data, []byte("PAR1")) {
		t.Fatal("output is not a Parquet file")
	}

	rows, err := parquet.Read[LogEntry](bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != len(entries) {
		t.Fatalf("read %d rows, want %d", len(rows), len(entries))
	}

	if rows[0].Message != "Events synced" || rows[0].RequestID != "req-1" {
		t.Errorf("row 0 = %+v", rows[0])
	}
	if optStr(rows[0].Context) != `{"count":5}` {
		t.Errorf("row 0 context = %q", optStr(rows[0].Context))
	}
	// Optional columns must survive as null rather than as an empty string.
	if rows[0].StackTrace != nil {
		t.Errorf("row 0 stack_trace = %q, want nil", optStr(rows[0].StackTrace))
	}
	if rows[1].Caller != nil {
		t.Errorf("row 1 caller = %q, want nil", optStr(rows[1].Caller))
	}
}

// The Glue table declares parquet.compression=SNAPPY. Before this, the writer
// set no codec at all while the DDL claimed snappy.
func TestWriteParquetAppliesCompression(t *testing.T) {
	cases := map[string]format.CompressionCodec{
		"snappy":       format.Snappy,
		"zstd":         format.Zstd,
		"gzip":         format.Gzip,
		"uncompressed": format.Uncompressed,
	}

	for name, want := range cases {
		data, err := WriteParquet(sampleEntries(), name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		f, err := parquet.OpenFile(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, rg := range f.Metadata().RowGroups {
			for _, col := range rg.Columns {
				if col.MetaData.Codec != want {
					t.Errorf("%s: column codec = %v, want %v", name, col.MetaData.Codec, want)
				}
			}
		}
	}
}

// An unknown codec must not fail the flush; snappy is the documented fallback.
func TestWriteParquetUnknownCompressionFallsBackToSnappy(t *testing.T) {
	data, err := WriteParquet(sampleEntries(), "brotli-ish")
	if err != nil {
		t.Fatal(err)
	}

	f, err := parquet.OpenFile(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	for _, rg := range f.Metadata().RowGroups {
		for _, col := range rg.Columns {
			if col.MetaData.Codec != format.Snappy {
				t.Fatalf("codec = %v, want Snappy", col.MetaData.Codec)
			}
		}
	}
}

func TestIsSupportedCompression(t *testing.T) {
	for _, ok := range []string{"snappy", "zstd", "gzip", "uncompressed", "none"} {
		if !IsSupportedCompression(ok) {
			t.Errorf("%q should be supported", ok)
		}
	}
	if IsSupportedCompression("lzo") {
		t.Error("lzo should not be reported as supported")
	}
}
