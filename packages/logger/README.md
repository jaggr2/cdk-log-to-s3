# @jaggr2/log-to-s3-logger

Zero-dependency structured JSON logger for AWS Lambda. Companion to
[`@jaggr2/cdk-log-to-s3`](https://github.com/jaggr2/cdk-log-to-s3), whose
extension parses these records into Parquet columns.

Useful on its own too: it writes ordinary JSON lines to stdout, which
CloudWatch Logs Insights queries happily.

```bash
npm install @jaggr2/log-to-s3-logger
```

## Usage

```ts
import { createStructuredLogger, runWithCorrelationId } from '@jaggr2/log-to-s3-logger';

const log = createStructuredLogger('SyncProcessor');

export const handler = async (event: unknown, context: { awsRequestId: string }) =>
  runWithCorrelationId(context.awsRequestId, async () => {
    log.info('Events synced', { count: 5 });
    log.debug('Processing', { eventId: '123' });

    try {
      await sync();
    } catch (err) {
      log.error('Sync failed', err, { attempt: 2 });
    }
  });
```

`runWithCorrelationId` puts the id in an `AsyncLocalStorage`, so every logger
created inside it picks the id up automatically - across awaits, without
threading it through call signatures.

## API

```ts
createStructuredLogger(source: string, options?: StructuredLoggerOptions): StructuredLogger
```

| Option | Default | |
|---|---|---|
| `correlationId` | ambient, else `''` | Pins the id for this logger |
| `level` | `LOG_TO_S3_LEVEL`, else `LOG_LEVEL`, else `INFO` | Minimum level to emit |
| `captureCaller` | `true` | Records the call site via a stack trace |
| `callerRootMarker` | `'/src/'` | Where to trim absolute paths |
| `write` | console | Output sink; a test seam |

```ts
log.debug(message, context?)
log.info(message, context?)
log.warn(message, context?)
log.error(message, err?, context?)

log.child('OtherSource')          // same settings, different source
log.withCorrelationId('abc-123')  // same settings, pinned id
```

`error()` splits a thrown `Error` into `context.error` (`name`, `message`) and a
top-level `stackTrace`, which the extension stores in its own `stack_trace`
column so it stays queryable.

Correlation helpers:

```ts
runWithCorrelationId<T>(id: string, fn: () => T): T
getCorrelationId(): string | undefined
setCorrelationId(id: string): void   // only inside a runWithCorrelationId scope
```

## Level filtering happens here

A record below the configured level is never written, so it costs no CloudWatch
ingest. Set `LOG_TO_S3_LEVEL` on the function - the CDK construct does this for
you - and the same value governs both the logger and the extension.

## Output

One line per record:

```json
{"__log_level":"INFO","__source":"SyncProcessor","__correlation_id":"7f3a","__caller":"src/lib/sync.ts:142:11","message":"Events synced","context":{"count":5}}
```

`__log_level` must always be present: it is the discriminator the extension uses
to tell a structured record from an arbitrary JSON line an application happened
to log. The full contract, and the fixtures that pin both sides of it, are in
[docs/CONTRACT.md](https://github.com/jaggr2/cdk-log-to-s3/blob/main/docs/CONTRACT.md).

## License

MIT
