import { Duration, RemovalPolicy, Size, Stack } from "aws-cdk-lib";
import * as events from "aws-cdk-lib/aws-events";
import * as targets from "aws-cdk-lib/aws-events-targets";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as logs from "aws-cdk-lib/aws-logs";
import * as s3 from "aws-cdk-lib/aws-s3";
import { Construct } from "constructs";
import { ILogToS3Extension, LogCompression } from "./log-to-s3-extension";
import { resolveCompactorAsset } from "./private/assets";
import { DEFAULT_KEY_PREFIX, ENV, normalizePrefix } from "./private/env";

export interface LogCompactionProps {
  /** Bucket the extension writes to. */
  readonly logsBucket: s3.IBucket;

  /**
   * Must equal LogToS3ExtensionProps.keyPrefix, or the job will look in the
   * wrong place and quietly find nothing to do. Use
   * LogCompaction.fromExtension() to make that impossible.
   *
   * @default 'logs/'
   */
  readonly keyPrefix?: string;

  /**
   * When to run. Compaction only touches closed days, so anything daily works;
   * off-peak is polite but not required.
   *
   * @default - every day at 03:00 UTC
   */
  readonly schedule?: events.Schedule;

  /**
   * Create the schedule. Set false to deploy the function but trigger it
   * yourself.
   *
   * @default true
   */
  readonly enabled?: boolean;

  /**
   * How many closed days each run considers, most recent first. More than one
   * so a failed or skipped run catches up by itself.
   *
   * @default Duration.days(7)
   */
  readonly lookback?: Duration;

  /**
   * Leave a partition alone until it holds at least this many files. Below it
   * the read and rewrite costs more than the scan it saves.
   *
   * @default 8
   */
  readonly minFilesPerPartition?: number;

  /**
   * Cap on files merged per partition per run. Whatever is left over is picked
   * up next run, and every run still reduces the file count.
   *
   * @default 2000
   */
  readonly maxFilesPerRun?: number;

  /**
   * Cap on bytes read per partition per run. Keep it comfortably below
   * `memorySize`: rows are held in memory while merging.
   *
   * @default Size.mebibytes(256)
   */
  readonly maxBytesPerRun?: Size;

  /**
   * Codec for the merged output. Match the extension unless you want the
   * archive stored more densely than it was written.
   *
   * @default LogCompression.SNAPPY
   */
  readonly compression?: LogCompression;

  /**
   * @default lambda.Architecture.ARM_64
   */
  readonly architecture?: lambda.Architecture;

  /**
   * @default Size.mebibytes(1024)
   */
  readonly memorySize?: Size;

  /**
   * @default Duration.minutes(5)
   */
  readonly timeout?: Duration;

  /**
   * Retention for the job's own CloudWatch logs.
   *
   * @default logs.RetentionDays.ONE_MONTH
   */
  readonly logRetention?: logs.RetentionDays;

  /**
   * Report every partition considered, not only the ones compacted.
   *
   * @default false
   */
  readonly debug?: boolean;

  /**
   * @default RemovalPolicy.DESTROY
   */
  readonly removalPolicy?: RemovalPolicy;
}

/** Options for LogCompaction.fromExtension. */
export interface LogCompactionFromExtensionOptions {
  /**
   * @default - every day at 03:00 UTC
   */
  readonly schedule?: events.Schedule;

  /**
   * @default true
   */
  readonly enabled?: boolean;

  /**
   * @default Duration.days(7)
   */
  readonly lookback?: Duration;

  /**
   * @default 8
   */
  readonly minFilesPerPartition?: number;

  /**
   * @default 2000
   */
  readonly maxFilesPerRun?: number;

  /**
   * @default Size.mebibytes(256)
   */
  readonly maxBytesPerRun?: Size;

  /**
   * @default LogCompression.SNAPPY
   */
  readonly compression?: LogCompression;

  /**
   * @default Size.mebibytes(1024)
   */
  readonly memorySize?: Size;

  /**
   * @default Duration.minutes(5)
   */
  readonly timeout?: Duration;

  /**
   * @default logs.RetentionDays.ONE_MONTH
   */
  readonly logRetention?: logs.RetentionDays;

  /**
   * @default false
   */
  readonly debug?: boolean;

  /**
   * @default RemovalPolicy.DESTROY
   */
  readonly removalPolicy?: RemovalPolicy;
}

/**
 * A daily job that merges the many small Parquet files the extension produces
 * into fewer, larger ones.
 *
 * The extension flushes on a timer, on a size threshold and at the end of every
 * invocation, so a busy function can leave thousands of tiny objects in a day
 * partition. Athena pays a per-file cost opening footers, so that is slow to
 * scan regardless of how little data it holds. Compaction is the answer to that
 * - not finer partitioning, which only moves the cost into the query planner.
 *
 * It works purely at the S3 level and never touches the Glue catalog. Under
 * partition projection there is nothing to register: Athena computes partitions
 * from the `dt` range at query time, and rewriting files inside a partition
 * does not change the partition set. This is why the construct needs no Glue
 * permissions and why `MSCK REPAIR` has no role here.
 *
 * Only closed days are compacted; today is left alone while the extension is
 * still writing into it.
 *
 * @example
 * const compaction = LogCompaction.fromExtension(this, 'Compaction', extension);
 */
export class LogCompaction extends Construct {
  /**
   * Build the job for an extension, taking the bucket, key prefix,
   * architecture and codec from it so they cannot drift apart.
   */
  public static fromExtension(
    scope: Construct,
    id: string,
    extension: ILogToS3Extension,
    options: LogCompactionFromExtensionOptions = {},
  ): LogCompaction {
    return new LogCompaction(scope, id, {
      ...options,
      logsBucket: extension.logsBucket,
      keyPrefix: extension.keyPrefix,
      architecture: extension.architecture,
    });
  }

  /** The compaction function. */
  public readonly handler: lambda.Function;
  /** The schedule, or undefined when `enabled` was false. */
  public readonly rule?: events.Rule;
  /** Normalised key prefix the job operates under. */
  public readonly keyPrefix: string;

  constructor(scope: Construct, id: string, props: LogCompactionProps) {
    super(scope, id);

    const architecture = props.architecture ?? lambda.Architecture.ARM_64;
    this.keyPrefix = normalizePrefix(props.keyPrefix ?? DEFAULT_KEY_PREFIX);

    const lookback = props.lookback ?? Duration.days(7);
    // Checked in milliseconds: toDays() raises its own error for anything that
    // is not a whole number of days, which would mask this message.
    if (lookback.toMilliseconds() < Duration.days(1).toMilliseconds()) {
      throw new Error("LogCompaction: lookback must be at least one day.");
    }
    const lookbackDays = lookback.toDays();

    const environment: { [key: string]: string } = {
      [ENV.BUCKET]: props.logsBucket.bucketName,
      [ENV.PREFIX]: this.keyPrefix,
      [ENV.COMPRESSION]: props.compression ?? LogCompression.SNAPPY,
      [ENV.COMPACTION_LOOKBACK_DAYS]: `${lookbackDays}`,
    };
    if (props.minFilesPerPartition !== undefined) {
      environment[ENV.COMPACTION_MIN_FILES] = `${props.minFilesPerPartition}`;
    }
    if (props.maxFilesPerRun !== undefined) {
      environment[ENV.COMPACTION_MAX_FILES] = `${props.maxFilesPerRun}`;
    }
    if (props.maxBytesPerRun !== undefined) {
      environment[ENV.COMPACTION_MAX_BYTES] =
        `${props.maxBytesPerRun.toBytes()}`;
    }
    if (props.debug !== undefined) {
      environment[ENV.DEBUG] = `${props.debug}`;
    }

    this.handler = new lambda.Function(this, "Handler", {
      // A Go binary needs no managed runtime; provided.al2023 just executes
      // the `bootstrap` file the build script places at the archive root.
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: "bootstrap",
      architecture,
      code: lambda.Code.fromAsset(resolveCompactorAsset(architecture)),
      memorySize: (props.memorySize ?? Size.mebibytes(1024)).toMebibytes(),
      timeout: props.timeout ?? Duration.minutes(5),
      environment,
      logGroup: new logs.LogGroup(this, "Logs", {
        retention: props.logRetention ?? logs.RetentionDays.ONE_MONTH,
        removalPolicy: props.removalPolicy ?? RemovalPolicy.DESTROY,
      }),
      description: "log-to-s3: merges small Parquet log files into larger ones",
    });

    // Read to merge, write the merged output, delete the sources. All scoped
    // to the prefix; nothing else in the bucket is reachable.
    props.logsBucket.grantRead(this.handler, `${this.keyPrefix}*`);
    props.logsBucket.grantPut(this.handler, `${this.keyPrefix}*`);
    props.logsBucket.grantDelete(this.handler, `${this.keyPrefix}*`);

    if (props.enabled ?? true) {
      this.rule = new events.Rule(this, "Schedule", {
        schedule:
          props.schedule ?? events.Schedule.cron({ minute: "0", hour: "3" }),
        description: "Daily compaction of log-to-s3 Parquet files",
        targets: [new targets.LambdaFunction(this.handler)],
      });
    }
  }

  /** ARN of the compaction function, for wiring alarms or manual invocation. */
  public get functionArn(): string {
    return Stack.of(this).resolve(this.handler.functionArn);
  }
}
