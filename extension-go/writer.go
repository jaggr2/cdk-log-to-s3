package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress"
)

const uploadTimeout = 30 * time.Second

// codecs is the full set of compression codecs this extension accepts. It is
// mirrored by the LogCompression enum on the CDK side; adding one here means
// adding it there too.
func codecFor(name string) (compress.Codec, bool) {
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
	_, ok := codecFor(name)
	return ok
}

// WriteParquet serialises entries into an in-memory Parquet file.
//
// Compression is applied here rather than through parquet struct tags: a tag
// would take precedence over this option and make LOG_TO_S3_COMPRESSION a
// setting that silently does nothing.
func WriteParquet(entries []LogEntry, compression string) ([]byte, error) {
	codec, ok := codecFor(compression)
	if !ok {
		codec = &parquet.Snappy
	}

	buf := new(bytes.Buffer)
	w := parquet.NewGenericWriter[LogEntry](buf, parquet.Compression(codec))

	if _, err := w.Write(entries); err != nil {
		w.Close()
		return nil, fmt.Errorf("write parquet rows: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close parquet writer: %w", err)
	}
	return buf.Bytes(), nil
}

// Uploader is the S3 sink. It is nil-safe at the call site: when no bucket is
// configured the listener never constructs one.
type Uploader struct {
	client *s3.Client
	bucket string
	prefix string
	now    func() time.Time
}

func NewUploader(ctx context.Context, bucket, prefix, region string) (*Uploader, error) {
	opts := []func(*awsconfig.LoadOptions) error{}
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	return &Uploader{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
		prefix: NormalizePrefix(prefix),
		now:    time.Now,
	}, nil
}

// Upload writes one Parquet object and returns the key it used.
func (u *Uploader) Upload(ctx context.Context, data []byte, functionName, correlationID string) (string, error) {
	key := BuildKey(u.prefix, u.now(), functionName, correlationID)

	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/vnd.apache.parquet"),
	})
	if err != nil {
		return key, err
	}
	return key, nil
}
