import { Duration, RemovalPolicy, Size } from "aws-cdk-lib";
import * as iam from "aws-cdk-lib/aws-iam";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as s3 from "aws-cdk-lib/aws-s3";
import { Construct, IConstruct } from "constructs";
import { resolveLayerAsset } from "./private/assets";
import { DEFAULT_KEY_PREFIX, ENV, normalizePrefix } from "./private/env";

/** Compression codec used for the Parquet output. */
export enum LogCompression {
  /** Good ratio at low CPU cost. The default, and what the Glue table declares. */
  SNAPPY = "snappy",
  /** Smaller files, more CPU. */
  ZSTD = "zstd",
  GZIP = "gzip",
  /** No compression. Mostly useful for debugging. */
  UNCOMPRESSED = "uncompressed",
}

/** Minimum level a record must have to be written to S3. */
export enum LogLevel {
  DEBUG = "DEBUG",
  INFO = "INFO",
  WARN = "WARN",
  ERROR = "ERROR",
}

export interface LogToS3ExtensionProps {
  /**
   * Bucket the extension writes Parquet files to. Any IBucket works: the
   * LogBucket from this package, a plain s3.Bucket, or an imported one.
   */
  readonly logsBucket: s3.IBucket;

  /**
   * Architecture of the functions this extension will be attached to. A layer
   * carries a native binary, so one instance serves one architecture; create a
   * second instance for functions on the other.
   *
   * @default lambda.Architecture.ARM_64
   */
  readonly architecture?: lambda.Architecture;

  /**
   * Key prefix to write under. Must match LogAnalyticsProps.keyPrefix, or the
   * table will be empty with no error reported anywhere.
   *
   * @default 'logs/'
   */
  readonly keyPrefix?: string;

  /**
   * Records below this level are dropped before being buffered.
   *
   * @default LogLevel.INFO
   */
  readonly logLevel?: LogLevel;

  /**
   * How often buffered records are written out, independent of size. Lower
   * values mean fresher data in Athena and more, smaller S3 objects.
   *
   * @default Duration.seconds(15)
   */
  readonly flushInterval?: Duration;

  /**
   * Buffer size that triggers a flush. This memory is charged to the function.
   *
   * @default Size.mebibytes(10)
   */
  readonly maxBufferSize?: Size;

  /**
   * @default LogCompression.SNAPPY
   */
  readonly compression?: LogCompression;

  /**
   * Port the extension listens on for the Telemetry API. Change it only if it
   * collides with something else in the sandbox.
   *
   * @default 2020
   */
  readonly telemetryPort?: number;

  /**
   * Emit one row per invocation with duration and memory metrics, derived from
   * the platform.report event.
   *
   * @default true
   */
  readonly includePlatformReport?: boolean;

  /**
   * Verbose self-logging from the extension. Off by default: every line it
   * writes is itself billed CloudWatch ingest on every invocation.
   *
   * @default false
   */
  readonly extensionDebug?: boolean;

  /**
   * @default - a name is generated
   */
  readonly layerVersionName?: string;

  /**
   * @default - a description mentioning the Telemetry API
   */
  readonly description?: string;

  /**
   * Runtimes the layer declares compatibility with. Left unset by default: the
   * extension is a standalone binary and works with every runtime, so pinning
   * a list only blocks legitimate consumers.
   *
   * @default - all runtimes
   */
  readonly compatibleRuntimes?: lambda.Runtime[];

  /**
   * @default RemovalPolicy.DESTROY
   */
  readonly removalPolicy?: RemovalPolicy;
}

export interface LogToS3AttachOptions {
  /**
   * Override the extension-wide level for this one function.
   *
   * @default - the level configured on the extension
   */
  readonly logLevel?: LogLevel;

  /**
   * Grant the function permission to write to the logs bucket, scoped to the
   * key prefix. Set false only if the role already has equivalent access.
   *
   * @default true
   */
  readonly grantWrite?: boolean;
}

/** Attributes needed to use a layer that was created elsewhere. */
export interface LogToS3ExtensionAttributes {
  readonly layerVersionArn: string;
  readonly logsBucket: s3.IBucket;

  /**
   * @default lambda.Architecture.ARM_64
   */
  readonly architecture?: lambda.Architecture;

  /**
   * @default 'logs/'
   */
  readonly keyPrefix?: string;

  /**
   * @default LogLevel.INFO
   */
  readonly logLevel?: LogLevel;

  /**
   * @default LogCompression.SNAPPY
   */
  readonly compression?: LogCompression;

  /**
   * @default 2020
   */
  readonly telemetryPort?: number;
}

export interface ILogToS3Extension extends IConstruct {
  /** The layer carrying the extension binary. */
  readonly layer: lambda.ILayerVersion;
  /** Bucket the extension writes to. */
  readonly logsBucket: s3.IBucket;
  /** Normalised key prefix, e.g. 'logs/'. */
  readonly keyPrefix: string;
  /** Architecture this layer was built for. */
  readonly architecture: lambda.Architecture;
  /** Environment variables attachTo() injects. Exposed for manual wiring. */
  readonly environment: { [key: string]: string };

  /**
   * Add the layer, the environment and the bucket grant to a function.
   *
   * Takes a concrete lambda.Function rather than an IFunction because
   * addLayers() and addEnvironment() only exist on the concrete class.
   * Functions imported with Function.fromFunctionArn cannot be modified by
   * CloudFormation anyway.
   */
  attachTo(fn: lambda.Function, options?: LogToS3AttachOptions): void;

  /** Grant write access to the log prefix without attaching the layer. */
  grantWriteLogs(grantee: iam.IGrantable): iam.Grant;
}

/**
 * A Lambda layer containing an external extension that subscribes to the
 * Telemetry API and writes structured logs to S3 as Parquet.
 *
 * Instantiate one per stack that owns functions, all pointing at the same
 * bucket. The layer asset hash is identical across stacks, so the zip is
 * uploaded once and the extra AWS::Lambda::LayerVersion resources are free.
 * That is deliberately cheaper than exporting the ARN across stacks, which
 * creates a CloudFormation export that cannot be changed while it is in use.
 *
 * @example
 * const bucket = new LogBucket(this, 'Logs');
 * const ext = new LogToS3Extension(this, 'LogExt', { logsBucket: bucket });
 * ext.attachTo(myFunction);
 */
export class LogToS3Extension extends Construct implements ILogToS3Extension {
  /** Reference a layer created by another app or stack. */
  public static fromAttributes(
    scope: Construct,
    id: string,
    attrs: LogToS3ExtensionAttributes,
  ): ILogToS3Extension {
    class Import extends Construct implements ILogToS3Extension {
      public readonly layer: lambda.ILayerVersion;
      public readonly logsBucket: s3.IBucket;
      public readonly keyPrefix: string;
      public readonly architecture: lambda.Architecture;
      public readonly environment: { [key: string]: string };

      private readonly attached = new Set<string>();

      constructor() {
        super(scope, id);
        this.layer = lambda.LayerVersion.fromLayerVersionArn(
          this,
          "Layer",
          attrs.layerVersionArn,
        );
        this.logsBucket = attrs.logsBucket;
        this.keyPrefix = normalizePrefix(attrs.keyPrefix ?? DEFAULT_KEY_PREFIX);
        this.architecture = attrs.architecture ?? lambda.Architecture.ARM_64;
        this.environment = buildEnvironment({
          bucketName: attrs.logsBucket.bucketName,
          keyPrefix: this.keyPrefix,
          logLevel: attrs.logLevel ?? LogLevel.INFO,
          compression: attrs.compression ?? LogCompression.SNAPPY,
          telemetryPort: attrs.telemetryPort,
        });
      }

      public attachTo(
        fn: lambda.Function,
        options?: LogToS3AttachOptions,
      ): void {
        attach(this, this.attached, fn, options);
      }

      public grantWriteLogs(grantee: iam.IGrantable): iam.Grant {
        return this.logsBucket.grantPut(grantee, `${this.keyPrefix}*`);
      }
    }

    return new Import();
  }

  public readonly layer: lambda.ILayerVersion;
  public readonly logsBucket: s3.IBucket;
  public readonly keyPrefix: string;
  public readonly architecture: lambda.Architecture;
  public readonly environment: { [key: string]: string };

  private readonly attached = new Set<string>();

  constructor(scope: Construct, id: string, props: LogToS3ExtensionProps) {
    super(scope, id);

    this.architecture = props.architecture ?? lambda.Architecture.ARM_64;
    this.logsBucket = props.logsBucket;
    this.keyPrefix = normalizePrefix(props.keyPrefix ?? DEFAULT_KEY_PREFIX);

    this.layer = new lambda.LayerVersion(this, "Layer", {
      code: lambda.Code.fromAsset(resolveLayerAsset(this.architecture)),
      compatibleArchitectures: [this.architecture],
      compatibleRuntimes: props.compatibleRuntimes,
      layerVersionName: props.layerVersionName,
      description:
        props.description ??
        "log-to-s3: captures Lambda telemetry and writes it to S3 as Parquet",
      removalPolicy: props.removalPolicy,
    });

    this.environment = buildEnvironment({
      bucketName: props.logsBucket.bucketName,
      keyPrefix: this.keyPrefix,
      logLevel: props.logLevel ?? LogLevel.INFO,
      compression: props.compression ?? LogCompression.SNAPPY,
      telemetryPort: props.telemetryPort,
      flushInterval: props.flushInterval,
      maxBufferSize: props.maxBufferSize,
      includePlatformReport: props.includePlatformReport,
      extensionDebug: props.extensionDebug,
    });
  }

  public attachTo(fn: lambda.Function, options?: LogToS3AttachOptions): void {
    attach(this, this.attached, fn, options);
  }

  public grantWriteLogs(grantee: iam.IGrantable): iam.Grant {
    return this.logsBucket.grantPut(grantee, `${this.keyPrefix}*`);
  }
}

/**
 * Shared attach implementation for both the owned and the imported extension.
 */
function attach(
  ext: ILogToS3Extension,
  attached: Set<string>,
  fn: lambda.Function,
  options?: LogToS3AttachOptions,
): void {
  if (fn.architecture.name !== ext.architecture.name) {
    throw new Error(
      `Cannot attach a ${ext.architecture.name} log-to-s3 extension to function "${fn.node.path}", ` +
        `which is ${fn.architecture.name}. A layer contains a native binary, so create a second ` +
        `LogToS3Extension with { architecture: lambda.Architecture.${fn.architecture.name.toUpperCase()} } ` +
        "for these functions.",
    );
  }

  // Attaching twice would add the layer twice and produce an invalid template.
  if (attached.has(fn.node.path)) {
    return;
  }
  attached.add(fn.node.path);

  // addLayers is additive and participates in the recognizeLayerVersion
  // feature flag, unlike assigning CfnFunction.layers, which replaces whatever
  // else the function already had.
  fn.addLayers(ext.layer);

  for (const [key, value] of Object.entries(ext.environment)) {
    fn.addEnvironment(key, value);
  }
  if (options?.logLevel) {
    fn.addEnvironment(ENV.LEVEL, options.logLevel);
  }

  if (options?.grantWrite ?? true) {
    // Scoped to the prefix rather than the whole bucket: the extension only
    // ever writes under it.
    ext.logsBucket.grantPut(fn, `${ext.keyPrefix}*`);
  }
}

interface EnvironmentOptions {
  readonly bucketName: string;
  readonly keyPrefix: string;
  readonly logLevel: LogLevel;
  readonly compression: LogCompression;
  readonly telemetryPort?: number;
  readonly flushInterval?: Duration;
  readonly maxBufferSize?: Size;
  readonly includePlatformReport?: boolean;
  readonly extensionDebug?: boolean;
}

function buildEnvironment(options: EnvironmentOptions): {
  [key: string]: string;
} {
  const env: { [key: string]: string } = {
    [ENV.BUCKET]: options.bucketName,
    [ENV.PREFIX]: options.keyPrefix,
    [ENV.LEVEL]: options.logLevel,
    [ENV.COMPRESSION]: options.compression,
  };

  if (options.telemetryPort !== undefined) {
    env[ENV.TELEMETRY_PORT] = `${options.telemetryPort}`;
  }
  if (options.flushInterval !== undefined) {
    // Checked in milliseconds: toSeconds() raises its own error for anything
    // that is not a whole number of seconds, which would mask this message.
    if (options.flushInterval.toMilliseconds() < 1000) {
      throw new Error(
        "LogToS3Extension: flushInterval must be at least one second.",
      );
    }
    env[ENV.FLUSH_INTERVAL_SECONDS] = `${options.flushInterval.toSeconds()}`;
  }
  if (options.maxBufferSize !== undefined) {
    env[ENV.MAX_BUFFER_BYTES] = `${options.maxBufferSize.toBytes()}`;
  }
  if (options.includePlatformReport !== undefined) {
    env[ENV.INCLUDE_PLATFORM_REPORT] = `${options.includePlatformReport}`;
  }
  if (options.extensionDebug !== undefined) {
    env[ENV.DEBUG] = `${options.extensionDebug}`;
  }

  return env;
}
