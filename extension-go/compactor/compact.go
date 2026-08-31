package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/jaggr2/cdk-log-to-s3/extension/internal/logrecord"
)

const (
	// Output files are named part-<hash>.parquet. The prefix marks a file as
	// already compacted so later runs do not feed it back into themselves.
	compactedPrefix = "part-"

	// The manifest is hidden from Athena. Hive-style readers skip files whose
	// name begins with an underscore or a dot, which is what lets a
	// non-Parquet bookkeeping file live inside the table location at all.
	manifestPrefix = "_part-"
	manifestSuffix = ".manifest.json"
)

// manifest records the sources a completed merge consumed.
//
// It is what makes compaction crash-safe. The output is written first, then
// the manifest, then the sources are deleted, then the manifest. A run that
// dies anywhere in that sequence leaves enough behind for the next run to
// finish the deletion rather than merge the same rows a second time.
type manifest struct {
	Sources []string `json:"sources"`
	Outputs []string `json:"outputs"`
	Written string   `json:"written"`
}

type PartitionResult struct {
	Partition      string `json:"partition"`
	SourceFiles    int    `json:"sourceFiles"`
	SourceBytes    int64  `json:"sourceBytes"`
	OutputKey      string `json:"outputKey,omitempty"`
	OutputBytes    int    `json:"outputBytes,omitempty"`
	Rows           int    `json:"rows,omitempty"`
	RecoveredRuns  int    `json:"recoveredRuns,omitempty"`
	Skipped        bool   `json:"skipped"`
	SkippedBecause string `json:"skippedBecause,omitempty"`
}

// CompactPartition merges the small Parquet files in one day partition into a
// single larger one.
//
// Many small files are the natural output of the extension: it flushes on a
// timer, on a size threshold and at the end of every invocation. Athena pays a
// per-file cost opening footers, so a partition of thousands of tiny files is
// slow to scan regardless of how little data it holds.
func CompactPartition(ctx context.Context, store ObjectStore, cfg *Config, partition string) (PartitionResult, error) {
	result := PartitionResult{Partition: partition}

	objects, err := store.List(ctx, partition)
	if err != nil {
		return result, err
	}

	// Finish any interrupted run before deciding what to merge. Skipping this
	// would merge rows that a previous run already wrote out, duplicating them.
	recovered, err := finishInterruptedRuns(ctx, store, objects)
	if err != nil {
		return result, err
	}
	result.RecoveredRuns = recovered
	if recovered > 0 {
		if objects, err = store.List(ctx, partition); err != nil {
			return result, err
		}
	}

	sources := selectSources(objects, cfg)
	result.SourceFiles = len(sources)
	for _, o := range sources {
		result.SourceBytes += o.Size
	}

	if len(sources) < cfg.MinFiles {
		result.Skipped = true
		result.SkippedBecause = fmt.Sprintf("only %d source files, minimum is %d", len(sources), cfg.MinFiles)
		return result, nil
	}

	keys := make([]string, len(sources))
	for i, o := range sources {
		keys[i] = o.Key
	}

	// The output name is a function of exactly which sources went into it, so
	// re-running after a crash rewrites the identical object instead of adding
	// a second copy of the same rows.
	id := digest(keys)
	outputKey := partition + compactedPrefix + id + ".parquet"
	manifestKey := partition + manifestPrefix + id + manifestSuffix

	entries := make([]logrecord.Entry, 0, len(sources)*64)
	for _, key := range keys {
		data, err := store.Get(ctx, key)
		if err != nil {
			return result, err
		}
		rows, err := logrecord.Read(data)
		if err != nil {
			return result, fmt.Errorf("%s: %w", key, err)
		}
		entries = append(entries, rows...)
	}

	// Sorting by timestamp tightens the per-row-group min/max statistics, so a
	// query with a timestamp predicate can skip row groups it would otherwise
	// have to decode. It also makes the merged output deterministic.
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Timestamp != entries[j].Timestamp {
			return entries[i].Timestamp < entries[j].Timestamp
		}
		return entries[i].RequestID < entries[j].RequestID
	})

	merged, err := logrecord.Write(entries, cfg.Compression)
	if err != nil {
		return result, err
	}

	if err := store.Put(ctx, outputKey, merged); err != nil {
		return result, err
	}

	body, err := json.Marshal(manifest{
		Sources: keys,
		Outputs: []string{outputKey},
		Written: cfg.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return result, err
	}
	if err := store.Put(ctx, manifestKey, body); err != nil {
		return result, err
	}

	if err := store.Delete(ctx, keys); err != nil {
		return result, err
	}
	if err := store.Delete(ctx, []string{manifestKey}); err != nil {
		return result, err
	}

	result.OutputKey = outputKey
	result.OutputBytes = len(merged)
	result.Rows = len(entries)
	return result, nil
}

// finishInterruptedRuns completes runs that died between writing the output
// and finishing the deletion of their sources. Named to avoid shadowing the
// builtin.
func finishInterruptedRuns(ctx context.Context, store ObjectStore, objects []Object) (int, error) {
	present := make(map[string]bool, len(objects))
	for _, o := range objects {
		present[o.Key] = true
	}

	count := 0
	for _, o := range objects {
		base := path.Base(o.Key)
		if !strings.HasPrefix(base, manifestPrefix) || !strings.HasSuffix(base, manifestSuffix) {
			continue
		}

		data, err := store.Get(ctx, o.Key)
		if err != nil {
			return count, err
		}
		var m manifest
		if err := json.Unmarshal(data, &m); err != nil {
			// An unreadable manifest must not wedge the job forever, but its
			// sources cannot be deleted safely either. Leave both alone.
			continue
		}

		stale := make([]string, 0, len(m.Sources))
		for _, src := range m.Sources {
			if present[src] {
				stale = append(stale, src)
			}
		}
		if len(stale) > 0 {
			if err := store.Delete(ctx, stale); err != nil {
				return count, err
			}
		}
		if err := store.Delete(ctx, []string{o.Key}); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// selectSources picks the files worth merging: real Parquet output, never a
// previous merge, never hidden bookkeeping. The caps bound one run's memory
// and runtime; whatever is left over is picked up by the next run, and every
// run still reduces the file count.
func selectSources(objects []Object, cfg *Config) []Object {
	candidates := make([]Object, 0, len(objects))
	for _, o := range objects {
		base := path.Base(o.Key)
		switch {
		case !strings.HasSuffix(base, ".parquet"):
			continue
		case strings.HasPrefix(base, compactedPrefix):
			continue
		case strings.HasPrefix(base, "_"), strings.HasPrefix(base, "."):
			continue
		}
		candidates = append(candidates, o)
	}

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Key < candidates[j].Key })

	var (
		out   []Object
		bytes int64
	)
	for _, o := range candidates {
		if len(out) >= cfg.MaxFiles {
			break
		}
		if len(out) > 0 && bytes+o.Size > cfg.MaxBytes {
			break
		}
		out = append(out, o)
		bytes += o.Size
	}
	return out
}

// digest fingerprints the exact source set, so the output key identifies what
// went into it.
func digest(keys []string) string {
	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}
