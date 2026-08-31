/**
 * Environment variable names read by the Go extension.
 *
 * These strings are a contract with extension-go/config.go; changing one here
 * without changing it there produces a setting that silently does nothing.
 *
 * Internal: not exported from the package entry point.
 */
export const ENV = {
  BUCKET: "LOG_TO_S3_BUCKET",
  PREFIX: "LOG_TO_S3_PREFIX",
  LEVEL: "LOG_TO_S3_LEVEL",
  FLUSH_INTERVAL_SECONDS: "LOG_TO_S3_FLUSH_INTERVAL_SECONDS",
  MAX_BUFFER_BYTES: "LOG_TO_S3_MAX_BUFFER_BYTES",
  TELEMETRY_PORT: "LOG_TO_S3_TELEMETRY_PORT",
  COMPRESSION: "LOG_TO_S3_COMPRESSION",
  INCLUDE_PLATFORM_REPORT: "LOG_TO_S3_INCLUDE_PLATFORM_REPORT",
  DEBUG: "LOG_TO_S3_DEBUG",

  COMPACTION_MIN_FILES: "LOG_TO_S3_COMPACTION_MIN_FILES",
  COMPACTION_MAX_FILES: "LOG_TO_S3_COMPACTION_MAX_FILES",
  COMPACTION_MAX_BYTES: "LOG_TO_S3_COMPACTION_MAX_BYTES",
  COMPACTION_LOOKBACK_DAYS: "LOG_TO_S3_COMPACTION_LOOKBACK_DAYS",
} as const;

/** Default S3 key prefix, shared by the extension and the Glue table. */
export const DEFAULT_KEY_PREFIX = "logs/";

/**
 * Normalises a key prefix to the form the extension writes and the Glue
 * location template expects: no leading slash, exactly one trailing slash,
 * or empty. Must stay identical to NormalizePrefix in extension-go/s3key.go.
 */
export function normalizePrefix(prefix: string): string {
  const trimmed = prefix.trim().replace(/^\/+/, "").replace(/\/+$/, "");
  return trimmed === "" ? "" : `${trimmed}/`;
}
