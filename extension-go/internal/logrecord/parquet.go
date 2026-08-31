package logrecord

import (
	"bytes"
	"fmt"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress"
)

// CodecFor maps a configured codec name onto a Parquet compression codec. The
// set is mirrored by the LogCompression enum on the CDK side; adding one here
// means adding it there too.
func CodecFor(name string) (compress.Codec, bool) {
	switch name {
	case "snappy":
		return &parquet.Snappy, true
	case "zstd":
		return &parquet.Zstd, true
	case "gzip":
		return &parquet.Gzip, true
	case "uncompressed", "none":
		return &parquet.Uncompressed, true
	default:
		return nil, false
	}
}

func IsSupportedCompression(name string) bool {
	_, ok := CodecFor(name)
	return ok
}

// Write serialises entries into an in-memory Parquet file.
//
// Compression is applied here rather than through struct tags: a tag would
// take precedence over this option and make the configured codec a setting
// that silently does nothing. An unknown codec falls back to snappy rather
// than failing the flush.
func Write(entries []Entry, compression string) ([]byte, error) {
	codec, ok := CodecFor(compression)
	if !ok {
		codec = &parquet.Snappy
	}

	buf := new(bytes.Buffer)
	w := parquet.NewGenericWriter[Entry](buf, parquet.Compression(codec))

	if _, err := w.Write(entries); err != nil {
		w.Close()
		return nil, fmt.Errorf("write parquet rows: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close parquet writer: %w", err)
	}
	return buf.Bytes(), nil
}

// Read parses an in-memory Parquet file back into entries.
func Read(data []byte) ([]Entry, error) {
	rows, err := parquet.Read[Entry](bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("read parquet: %w", err)
	}
	return rows, nil
}
