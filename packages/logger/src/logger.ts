import { getCorrelationId } from './correlation-id';
import {
  LEVEL_ORDER,
  LogLevelName,
  StructuredLogger,
  StructuredLoggerOptions,
  StructuredRecord,
} from './types';

const DEFAULT_CALLER_ROOT_MARKER = '/src/';

function defaultLevel(): LogLevelName {
  const raw = (process.env.LOG_TO_S3_LEVEL ?? process.env.LOG_LEVEL ?? 'INFO').toUpperCase();
  return raw in LEVEL_ORDER ? (raw as LogLevelName) : 'INFO';
}

function defaultWrite(level: LogLevelName, line: string): void {
  // The stream only affects how CloudWatch labels the line; the extension
  // parses both identically.
  if (level === 'ERROR') {
    console.error(line);
  } else {
    console.log(line);
  }
}

/**
 * Extracts the call site from a stack trace.
 *
 * `skip` counts the frames belonging to this module, so the frame reported is
 * the caller's, not the logger's.
 */
function captureCallSite(skip: number, rootMarker: string): string | undefined {
  const stack = new Error().stack;
  if (!stack) return undefined;

  const line = stack.split('\n')[skip + 1];
  if (!line) return undefined;

  const trimmed = line.trim();
  const match = trimmed.match(/\((.+?)\)$/) ?? trimmed.match(/^at\s+(.+)$/);
  if (!match) return trimmed;

  // Normalised to forward slashes so the marker matches on Windows too, and so
  // the caller column looks the same wherever the code was built.
  const location = match[1].replace(/\\/g, '/');
  const idx = location.indexOf(rootMarker);
  return idx >= 0 ? location.slice(idx + 1) : location;
}

/**
 * Turns an unknown thrown value into a context entry plus a stack trace.
 *
 * The stack goes into its own top-level field because the extension stores it
 * in a dedicated `stack_trace` column, which keeps it out of the JSON blob in
 * `context` and makes it directly queryable.
 */
function describeError(err: unknown): { entry: Record<string, unknown>; stackTrace?: string } {
  if (err instanceof Error) {
    return {
      entry: { name: err.name, message: err.message },
      stackTrace: err.stack,
    };
  }
  return { entry: { value: String(err) } };
}

export function createStructuredLogger(
  source: string,
  options: StructuredLoggerOptions = {},
): StructuredLogger {
  const level = options.level ?? defaultLevel();
  const threshold = LEVEL_ORDER[level];
  const captureCaller = options.captureCaller ?? true;
  const rootMarker = options.callerRootMarker ?? DEFAULT_CALLER_ROOT_MARKER;
  const write = options.write ?? defaultWrite;

  function emit(
    levelName: LogLevelName,
    message: string,
    context?: Record<string, unknown>,
    stackTrace?: string,
  ): void {
    if (LEVEL_ORDER[levelName] < threshold) return;

    const record: StructuredRecord = {
      __log_level: levelName,
      __source: source,
      __correlation_id: options.correlationId ?? getCorrelationId() ?? '',
      message,
    };

    if (captureCaller) {
      // 3 frames: captureCallSite, emit, and the public method.
      const caller = captureCallSite(3, rootMarker);
      if (caller) record.__caller = caller;
    }
    // Never emit an empty object or a null: the extension stores context as
    // raw JSON text and both would show up as noise in the column.
    if (context && Object.keys(context).length > 0) record.context = context;
    if (stackTrace) record.stackTrace = stackTrace;

    write(levelName, JSON.stringify(record));
  }

  const logger: StructuredLogger = {
    debug: (message, context) => emit('DEBUG', message, context),
    info: (message, context) => emit('INFO', message, context),
    warn: (message, context) => emit('WARN', message, context),

    error: (message, err, context) => {
      if (err === undefined) {
        emit('ERROR', message, context);
        return;
      }
      const { entry, stackTrace } = describeError(err);
      emit('ERROR', message, { ...context, error: entry }, stackTrace);
    },

    child: (childSource) => createStructuredLogger(childSource, { ...options, level }),

    withCorrelationId: (correlationId) =>
      createStructuredLogger(source, { ...options, level, correlationId }),
  };

  return logger;
}
