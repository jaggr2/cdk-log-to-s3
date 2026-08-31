# @jaggr2/cdk-log-to-s3

CDK constructs for a Lambda **external extension** that subscribes to the
Telemetry API and writes your function logs to S3 as **Parquet**, queryable
straight from Athena.

CloudWatch Logs is expensive to retain and awkward to query across functions.
This ships the logs to columnar storage instead, where a month of data costs
cents and `SELECT ... WHERE level = 'ERROR'` runs across every function at once.

- **No Go, no Docker.** The layer binaries are prebuilt and ship inside the npm
  package. `cdk synth` just works.
- **No crawler, no MSCK REPAIR, no scheduled Lambda.** Partitions resolve
  through Athena partition projection.
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
WHERE year = '2026' AND month = '08' AND day = '31'
  AND level = 'ERROR'
ORDER BY ts DESC
LIMIT 100;
```

**Always filter on the partition columns.** Projection enumerates every
combination in range, so an unfiltered query makes Athena consider
`years x 12 x 31 x 24` partitions. Keep `projectionYearRange` tight.

`context` is stored as raw JSON text:

```sql
SELECT json_extract_scalar(context, '$.count') AS count, message
FROM my_app_logs.app_logs
WHERE year = '2026' AND month = '08' AND source = 'SyncProcessor';
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

Partitioned by `year` / `month` / `day` / `hour`, all strings.

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

Set `createDatabase: false` when the database already exists -
`AWS::Glue::Database` fails if it does.

## Development

```bash
npx projen build          # compile, test, lint, docs
npx projen build:layers   # rebuild the Go binaries (needs Go)
npx projen test:go        # Go unit tests
```

The layer zips in `assets/` are committed. CI rebuilds them whenever
`extension-go/` changes and fails if the result differs, so the binaries always
match the source. Builds are reproducible: `-trimpath`, an empty build id, a
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
