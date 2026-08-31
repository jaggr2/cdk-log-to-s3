package main

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Object is one listed key and its size.
type Object struct {
	Key  string
	Size int64
}

// ObjectStore is the slice of S3 the compactor needs. It exists so the
// compaction protocol can be tested exhaustively - including its crash paths -
// without touching AWS.
type ObjectStore interface {
	List(ctx context.Context, prefix string) ([]Object, error)
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, body []byte) error
	Delete(ctx context.Context, keys []string) error
}

type s3Store struct {
	client *s3.Client
	bucket string
}

func NewS3Store(ctx context.Context, bucket, region string) (ObjectStore, error) {
	opts := []func(*awsconfig.LoadOptions) error{}
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}
	return &s3Store{client: s3.NewFromConfig(cfg), bucket: bucket}, nil
}

func (s *s3Store) List(ctx context.Context, prefix string) ([]Object, error) {
	var out []Object
	p := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list %s: %w", prefix, err)
		}
		for _, o := range page.Contents {
			out = append(out, Object{Key: aws.ToString(o.Key), Size: aws.ToInt64(o.Size)})
		}
	}
	return out, nil
}

func (s *s3Store) Get(ctx context.Context, key string) ([]byte, error) {
	res, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", key, err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", key, err)
	}
	return data, nil
}

func (s *s3Store) Put(ctx context.Context, key string, body []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(body),
	})
	if err != nil {
		return fmt.Errorf("put %s: %w", key, err)
	}
	return nil
}

// Delete removes keys in as few calls as possible. Batching matters here
// beyond efficiency: the window in which a crash can leave a partition
// half-deleted is exactly the time these calls take.
func (s *s3Store) Delete(ctx context.Context, keys []string) error {
	const batch = 1000

	for start := 0; start < len(keys); start += batch {
		end := start + batch
		if end > len(keys) {
			end = len(keys)
		}

		ids := make([]s3types.ObjectIdentifier, 0, end-start)
		for _, k := range keys[start:end] {
			ids = append(ids, s3types.ObjectIdentifier{Key: aws.String(k)})
		}

		res, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &s3types.Delete{Objects: ids, Quiet: aws.Bool(true)},
		})
		if err != nil {
			return fmt.Errorf("delete objects: %w", err)
		}
		if len(res.Errors) > 0 {
			return fmt.Errorf("delete %s: %s", aws.ToString(res.Errors[0].Key), aws.ToString(res.Errors[0].Message))
		}
	}
	return nil
}
