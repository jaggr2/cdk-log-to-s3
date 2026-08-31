// Command compactor is an AWS Lambda function that merges the many small
// Parquet files the log-to-s3 extension produces into fewer, larger ones.
//
// The extension flushes on a timer, on a size threshold and at the end of
// every invocation, so a busy function can leave thousands of tiny objects in
// a day partition. Athena pays a per-file cost opening footers, so that is
// slow to scan regardless of how little data it holds.
//
// It deliberately does NOT run MSCK REPAIR or touch the Glue catalog. Under
// partition projection there are no partitions to register: Athena computes
// them from the dt range at query time. Compaction only rewrites files inside
// a partition, so the partition set never changes.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

type Report struct {
	Partitions []PartitionResult `json:"partitions"`
	Compacted  int               `json:"compacted"`
	FilesIn    int               `json:"filesIn"`
	Errors     []string          `json:"errors,omitempty"`
}

func handler(ctx context.Context) (Report, error) {
	cfg, warnings := LoadConfig()
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "LOG_TO_S3_COMPACTION_CONFIG_WARNING: %s\n", w)
	}
	if cfg.Bucket == "" {
		return Report{}, fmt.Errorf("%s is not set", envBucket)
	}

	store, err := NewS3Store(ctx, cfg.Bucket, cfg.Region)
	if err != nil {
		return Report{}, err
	}

	report := Report{}
	for _, partition := range cfg.Partitions() {
		result, err := CompactPartition(ctx, store, cfg, partition)
		if err != nil {
			// One bad partition must not stop the others: the next run would
			// hit the same one first and never reach the rest.
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", partition, err))
			fmt.Fprintf(os.Stderr, "LOG_TO_S3_COMPACTION_ERROR %s: %v\n", partition, err)
			continue
		}

		report.FilesIn += result.SourceFiles
		if !result.Skipped {
			report.Compacted++
		}
		if cfg.Debug || !result.Skipped {
			report.Partitions = append(report.Partitions, result)
		}
	}

	if body, err := json.Marshal(report); err == nil {
		fmt.Fprintf(os.Stderr, "LOG_TO_S3_COMPACTION_REPORT %s\n", body)
	}

	if len(report.Errors) > 0 {
		return report, fmt.Errorf("%d partition(s) failed to compact", len(report.Errors))
	}
	return report, nil
}

func main() {
	lambda.Start(handler)
}
