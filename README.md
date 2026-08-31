# @jaggr2/cdk-log-to-s3

CDK constructs for a Lambda **external extension** that subscribes to the
Telemetry API and writes your function logs to S3 as **Parquet**, queryable
straight from Athena.

CloudWatch Logs is expensive to retain and awkward to query across functions.
This ships the logs to columnar storage instead, where a month of data costs
cents and `SELECT ... WHERE level = 'ERROR'` runs across every function at once.

- **No Go, no Docker.** The layer binaries are prebuilt and ship inside the npm
  package. `cdk synth` just works.
- **No crawler and no MSCK REPAIR.** Partitions resolve through Athena
  partition projection.
- **Optional daily compaction** for the small files a chatty function produces.
- **arm64 and x86_64.**

```bash
npm install @jaggr2/cdk-log-to-s3
```

## Quick start

```ts
import { LogAnalytics, LogBucket, LogToS3Extension } from '@jaggr2/cdk-log-to-s3';

const logsBucket = new LogBucket(this, 'Logs');

const extension = new LogToS3Extension(this, 'LogExtension', { logsBucket });
extension.attachTo(myFunction);
extension.attachTo(myOtherFunction);

// Optional: a Glue table and Athena workgroup for querying it.
LogAnalytics.fromExtension(this, 'Analytics', extension, {
  databaseName: 'my_app_logs',
});
```

`attachTo` adds the layer, injects the extension's environment variables, and
grants `s3:PutObject` scoped to the key prefix. It is additive - any layer the
function already had is preserved - and idempotent.

## Emitting structured logs

The extension parses any line on stdout, but it fills in `source`,
`correlation_id`, `context`, `stack_trace` and `caller` only for records in the
[documented JSON contract](./docs/CONTRACT.md). The companion package emits it:

```bash
npm install @jaggr2/log-to-s3-logger
```

```ts
import { createStructuredLogger, runWithCorrelationId } from '@jaggr2/log-to-s3-logger';

const log = createStructuredLogger('SyncProcessor');

export const handler = async (event, context) =>
  runWithCorrelationId(context.awsRequestId, async () => {
    log.info('Events synced', { count: 5 });
    log.error('Sync failed', err, { attempt: 2 });
  });
```

Anything else - `console.log` from a dependency, a runtime warning - still lands
in the table as a plain-text row.

## Querying

```sql
SELECT from_iso8601_timestamp(timestamp) AS ts, level, source, message, correlation_id
FROM my_app_logs.app_logs
WHERE dt = '2026/08/31'
  AND level = 'ERROR'
ORDER BY ts DESC
LIMIT 100;
```

`dt` is the partition column, formatted `yyyy/MM/dd`. It is zero-padded and
fixed width, so it sorts lexicographically in the same order it sorts
chronologically - which makes range predicates correct, not just convenient:

```sql
WHERE dt BETWEEN '2026/08/01' AND '2026/08/31'
```

**Always filter on `dt`.** Projection enumerates every day in the window, so an
unfiltered query makes Athena consider all of them (~730 at the default
two-year window).

`context` is stored as raw JSON text:

```sql
SELECT json_extract_scalar(context, '$.count') AS count, message
FROM my_app_logs.app_logs
WHERE dt BETWEEN '2026/08/01' AND '2026/08/31'
  AND source = 'SyncProcessor';
```

## Schema

| Column | Notes |
|---|---|
| `timestamp` | RFC3339 string. Use `from_iso8601_timestamp()`. |
| `level` | `DEBUG` \| `INFO` \| `WARN` \| `ERROR` |
| `source` | From `__source` |
| `correlation_id` | From `__correlation_id` |
| `request_id` | Lambda request id, recovered from the runtime log prefix |
| `message` | |
| `function_name` | |
| `context` | Raw JSON text |
| `stack_trace` | |
| `caller` | From `__caller` |

Partitioned by a single `dt` string column, formatted `yyyy/MM/dd`. One sliding
date column rather than four numeric ones: partition projection enumerates
every combination in range, and `year`/`month`/`day`/`hour` produced tens of
thousands of partitions that grew without bound.

## Constructs

### `LogToS3Extension`

| Prop | Default | |
|---|---|---|
| `logsBucket` | *required* | Any `IBucket` |
| `architecture` | `ARM_64` | Must match the functions it is attached to |
| `keyPrefix` | `'logs/'` | Must match `LogAnalytics` |
| `logLevel` | `INFO` | Records below this are dropped |
| `flushInterval` | 15s | |
| `maxBufferSize` | 10 MiB | Charged to the function's memory |
| `compression` | `SNAPPY` | |
| `includePlatformReport` | `true` | One duration/memory row per invocation |
| `extensionDebug` | `false` | Verbose self-logging; costs CloudWatch ingest |

A layer carries a native binary, so one instance serves one architecture.
Attaching to a function of the other architecture throws with a message telling
you to create a second instance.

**Across stacks, instantiate one per stack** rather than exporting the ARN. The
asset hash is identical, so the zip is uploaded once and the extra
`AWS::Lambda::LayerVersion` resources are free - and you avoid a CloudFormation
export that cannot be changed while it is in use. If you genuinely need to share
one layer version, put its ARN in SSM and use `LogToS3Extension.fromAttributes`.

### `LogBucket`

An `s3.Bucket` subclass: private, SSE-S3, TLS-only, `RemovalPolicy.RETAIN`, with
a 30-day IA / 90-day Glacier / 180-day expiry lifecycle. Tune with
`infrequentAccessAfter`, `glacierAfter`, `expireAfter` (a zero `Duration`
disables a step), or replace the lifecycle entirely with `lifecycleRules`.
Passing both throws rather than silently dropping one.

You do not have to use it - `logsBucket` accepts any `IBucket`.

### `LogAnalytics`

A Glue database and table plus an Athena workgroup and results bucket.
`grantQuery(grantee)` issues resource-scoped permissions for all of it.

Prefer `LogAnalytics.fromExtension(...)`, which takes the bucket and key prefix
from the extension. A prefix mismatch produces a table that returns no rows,
with no error from Athena or CloudFormation.

`projectionWindow` (default `Duration.days(730)`) is a **sliding** window ending
today, so the partition count stays bounded instead of growing forever. It must
stay wider than however long you keep the data: objects older than the window
are still in S3 but Athena cannot generate a partition for them, so they become
unqueryable with no error anywhere. The `LogBucket` default expires objects
after 180 days, well inside it.

Set `createDatabase: false` when the database already exists -
`AWS::Glue::Database` fails if it does.

### `LogCompaction`

A daily Lambda that merges the many small Parquet files the extension produces
into fewer, larger ones.

```ts
LogCompaction.fromExtension(this, 'Compaction', extension);
```

The extension flushes on a timer, on a size threshold and at the end of every
invocation, so a busy function can leave thousands of tiny objects in a day
partition. Athena pays a per-file cost opening footers, so that is slow to scan
regardless of how little data it holds. Compaction is the answer to that -
*not* finer partitioning, which only moves the cost into the query planner.

| Prop | Default | |
|---|---|---|
| `schedule` | daily 03:00 UTC | Any `events.Schedule` |
| `enabled` | `true` | `false` deploys the function without a schedule |
| `lookback` | 7 days | Closed days considered per run, so a failed run catches up |
| `minFilesPerPartition` | 8 | Below this, merging costs more than it saves |
| `maxFilesPerRun` | 2000 | Leftovers are picked up next run |
| `maxBytesPerRun` | 256 MiB | Keep below `memorySize`; rows are merged in memory |
| `memorySize` | 1024 MiB | |
| `timeout` | 5 min | |

Only closed days are compacted - today is left alone while the extension is
still writing into it. Merged rows are sorted by `timestamp`, which tightens
row-group statistics so a query with a timestamp predicate can skip more.

**It never touches the Glue catalog, and `MSCK REPAIR` has no role here.** That
command existed for tables whose partitions were discovered by scanning S3.
Under partition projection there is nothing to register: Athena computes
partitions from the `dt` range at query time, and rewriting files *inside* a
partition does not change the partition set. The construct therefore needs no
Glue permissions at all - only prefix-scoped read, write and delete on the
bucket.

Compaction is crash-safe. It writes the merged file, records the sources it
consumed in a hidden `_part-<id>.manifest.json`, deletes those sources, then
deletes the manifest. A run that dies at any point leaves enough behind for the
next run to finish the deletion rather than merge the same rows twice. The
output key is a hash of the exact source set, so a repeated write is idempotent.
(The manifest is hidden from Athena by its leading underscore, the same
convention that keeps Hadoop `_SUCCESS` files out of query results.)

## Development

```bash
npx projen build          # compile, test, lint, docs
npx projen build:layers   # rebuild the Go binaries (needs Go)
npx projen test:go        # Go unit tests
```

The zips in `assets/` are committed - the extension layer and the compaction
function, each for both architectures. CI rebuilds them whenever `extension-go/`
changes and fails if the result differs, so the binaries always match the
source. Builds are reproducible: `-trimpath`, an empty build id, a
pinned toolchain in `go.mod`, and a fixed timestamp on every zip entry.

To verify against real AWS:

```bash
npx projen integ:deploy && npx projen integ:verify && npx projen integ:destroy
```

That is the only check that proves partition projection resolves - a synth test
can confirm the location template is the string we intended, not that Athena
finds the files.

## License

MIT
