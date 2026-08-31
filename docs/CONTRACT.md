# Wire contract

The Lambda extension parses the lines your application writes to stdout. This
document is the contract between the two halves of this repository:

- `@jaggr2/log-to-s3-logger` (`packages/logger`) produces the lines.
- The Go extension (`extension-go`) parses them into Parquet columns.

Both sides are pinned by [`contract-fixtures.json`](./contract-fixtures.json),
which `packages/logger/test/contract.test.ts` and
`extension-go/contract_test.go` both read. Changing the format on one side
without the other fails a test, rather than silently producing a table full of
nulls.

## The record

One JSON object per line, `JSON.stringify` output, no pretty-printing:

```json
{"__log_level":"INFO","__source":"SyncProcessor","__correlation_id":"7f3a","__caller":"src/lib/sync.ts:142:11","message":"Events synced","context":{"count":5}}
```

| Field | Required | Notes |
|---|---|---|
| `__log_level` | **yes** | `DEBUG` \| `INFO` \| `WARN` \| `ERROR`. The discriminator - see below. |
| `__source` | yes | Logical component. Lands in the `source` column. |
| `__correlation_id` | yes | Empty string when unknown, never omitted. |
| `__caller` | no | Source location. Omitted when caller capture is off. |
| `message` | yes | Human-readable text. |
| `context` | no | Arbitrary JSON **object**. Omit when absent; never `null`. |
| `stackTrace` | no | String. Omit when absent. |

## Rules that are load-bearing

**`__log_level` must be present and non-empty.** It is the discriminator the Go
parser uses. A JSON line without it is not treated as a partially structured
record - it falls through to plain-text handling, and `source`,
`correlation_id`, `context` and `caller` all come out empty. This is deliberate:
half-parsing arbitrary JSON that an application logged for its own reasons would
be worse than not parsing it.

**`context` is an object, not a string.** The extension stores it as the raw
JSON text in the `context` column. Query it with Athena's JSON functions, e.g.
`json_extract_scalar(context, '$.count')`.

**One record is one line.** `JSON.stringify` escapes newlines and tabs, so this
holds automatically for any value you pass. Do not wrap the output in extra
text.

**Level filtering happens in the logger, not the extension.** A record below the
configured level is never written at all, so it costs no CloudWatch ingest. The
extension applies the same threshold again as a backstop.

## What the extension adds

These columns are filled in by the extension, not by the logger:

| Column | Source |
|---|---|
| `timestamp` | The Telemetry API event time, or the timestamp in the runtime log prefix. |
| `request_id` | Recovered from the runtime log prefix. |
| `function_name` | `AWS_LAMBDA_FUNCTION_NAME`. |

## Lines that are not structured records

Anything else on stdout still lands in the table - `console.log` from a
dependency, a runtime warning, a stack trace printed by the platform. Such a
line becomes a row with `level` `INFO` (or `ERROR` if the text contains an
error-ish word), the whole line as `message`, and empty `source`,
`correlation_id`, `context` and `caller`.

The extension also understands:

- **JSON log format** (`LAMBDA_LOG_FORMAT=JSON`), where the runtime delivers a
  structured envelope instead of a prefixed string. If the `message` inside it
  is itself one of our records, the fields are read from there.
- **`platform.report`** events, turned into one row per invocation with duration
  and memory figures in `context`. Disable with `includePlatformReport: false`.
