/**
 * The Parquet schema the extension writes, and therefore the Glue table
 * definition that can read it.
 *
 * This is the single source of truth for the column list. It is mirrored by
 * the parquet struct tags on LogEntry in extension-go/entry.go and asserted
 * against by extension-go/writer_test.go. A drift between the two produces a
 * table full of nulls and no error anywhere, so the two files must always be
 * edited together.
 *
 * Internal: not exported from the package entry point.
 */
export interface LogColumn {
  readonly name: string;
  readonly type: string;
}

export const LOG_COLUMNS: LogColumn[] = [
  /** RFC3339 with nanoseconds. Query with from_iso8601_timestamp(timestamp). */
  { name: "timestamp", type: "string" },
  { name: "level", type: "string" },
  /** Logical component that emitted the record, from __source. */
  { name: "source", type: "string" },
  { name: "correlation_id", type: "string" },
  /** Lambda request id, recovered from the runtime log prefix. */
  { name: "request_id", type: "string" },
  { name: "message", type: "string" },
  { name: "function_name", type: "string" },
  /** Arbitrary structured context, stored as raw JSON text. */
  { name: "context", type: "string" },
  { name: "stack_trace", type: "string" },
  /** Source location of the log call, from __caller. */
  { name: "caller", type: "string" },
];

/**
 * Partition keys, in path order.
 *
 * A single date column rather than year/month/day/hour: partition projection
 * enumerates every combination in range, and four dimensions produced tens of
 * thousands of partitions that grew without bound. One sliding date column is
 * ~730 partitions and stays that way. The value carries the path separators of
 * the yyyy/MM/dd format, so it is a string, not a date.
 */
export const PARTITION_COLUMNS: LogColumn[] = [{ name: "dt", type: "string" }];
