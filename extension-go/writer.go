package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jaggr2/cdk-log-to-s3/extension/internal/logrecord"
)

const uploadTimeout = 30 * time.Second

// The codec table and the Parquet writer live in internal/logrecord so the
// compactor uses exactly the same schema and compression behaviour.
func IsSupportedCompression(name string) bool {
	return logrecord.IsSupportedCompression(name)
}

// WriteParquet serialises entries into an in-memory Parquet file.
func WriteParquet(entries []LogEntry, compression string) ([]byte, error) {
	return logrecord.Write(entries, compression)
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
