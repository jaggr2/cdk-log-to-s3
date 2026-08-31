export type LogLevelName = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

/** Ordered so a threshold comparison works. Mirrors extension-go/config.go. */
export const LEVEL_ORDER: Record<LogLevelName, number> = {
  DEBUG: 0,
  INFO: 1,
  WARN: 2,
  ERROR: 3,
};

/**
 * The wire format the Lambda extension parses.
 *
 * `__log_level` is the discriminator: the Go parser treats a JSON line without
 * it as ordinary text, so it must always be present and non-empty. See
 * docs/CONTRACT.md.
 */
export interface StructuredRecord {
  __log_level: LogLevelName;
  __source: string;
  __correlation_id: string;
  __caller?: string;
  message: string;
  /** Arbitrary JSON object. Omitted when absent, never null. */
  context?: Record<string, unknown>;
  stackTrace?: string;
}

export interface StructuredLoggerOptions {
  /**
   * Correlation id for every record from this logger. Falls back to the
   * ambient id set by runWithCorrelationId.
   */
  correlationId?: string;

  /**
   * Minimum level to emit. Filtering happens here rather than in the
   * extension, so suppressed records cost no CloudWatch ingest at all.
   *
   * @default process.env.LOG_TO_S3_LEVEL ?? process.env.LOG_LEVEL ?? 'INFO'
   */
  level?: LogLevelName;

  /**
   * Capture the call site via a stack trace. Cheap, but not free in hot loops.
   *
   * @default true
   */
  captureCaller?: boolean;

  /**
   * Path fragment used to trim absolute paths out of the captured call site.
   *
   * @default '/src/'
   */
  callerRootMarker?: string;

  /**
   * Output sink. Exists so tests do not have to spy on the console.
   *
   * @default - console.error for ERROR, console.log otherwise
   */
  write?: (level: LogLevelName, line: string) => void;
}

export interface StructuredLogger {
  debug(message: string, context?: Record<string, unknown>): void;
  info(message: string, context?: Record<string, unknown>): void;
  warn(message: string, context?: Record<string, unknown>): void;
  error(message: string, err?: unknown, context?: Record<string, unknown>): void;

  /** A logger with a different source, inheriting everything else. */
  child(source: string): StructuredLogger;

  /** A logger pinned to a correlation id, inheriting everything else. */
  withCorrelationId(correlationId: string): StructuredLogger;
}
