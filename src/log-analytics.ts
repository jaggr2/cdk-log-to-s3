import { Duration, RemovalPolicy, Stack } from "aws-cdk-lib";
import * as athena from "aws-cdk-lib/aws-athena";
import * as glue from "aws-cdk-lib/aws-glue";
import * as iam from "aws-cdk-lib/aws-iam";
import * as s3 from "aws-cdk-lib/aws-s3";
import { Construct } from "constructs";
import { ILogToS3Extension } from "./log-to-s3-extension";
import { DEFAULT_KEY_PREFIX, normalizePrefix } from "./private/env";
import { LOG_COLUMNS, PARTITION_COLUMNS } from "./private/schema";

/** Inclusive range of years the partition projection covers. */
export interface ProjectionYearRange {
  readonly start: number;
  readonly end: number;
}

export interface LogAnalyticsProps {
  /** Bucket the extension writes to. */
  readonly logsBucket: s3.IBucket;

  /**
   * Glue database name. Lowercase, digits and underscores only.
   */
  readonly databaseName: string;

  /**
   * Must equal LogToS3ExtensionProps.keyPrefix. A mismatch produces a table
   * that returns no rows, with no error from Athena or CloudFormation.
   * Use LogAnalytics.fromExtension() to make that impossible.
   *
   * @default 'logs/'
   */
  readonly keyPrefix?: string;

  /**
   * @default 'app_logs'
   */
  readonly tableName?: string;

  /**
   * Create the Glue database. Set false when it already exists - notably when
   * migrating from a setup that created it imperatively, since
   * AWS::Glue::Database fails if the database is already there.
   *
   * @default true
   */
  readonly createDatabase?: boolean;

  /**
   * Years the projection covers.
   *
   * Keep this tight. Projection enumerates every combination, so the partition
   * count is years x 12 x 31 x 24; a query without a year/month predicate has
   * to consider all of them.
   *
   * @default - the current year through the current year plus five
   */
  readonly projectionYearRange?: ProjectionYearRange;

  /**
   * @default true
   */
  readonly createWorkgroup?: boolean;

  /**
   * @default - derived from the database name
   */
  readonly workgroupName?: string;

  /**
   * Bucket for Athena query results.
   *
   * @default - a bucket is created with a 30 day expiry
   */
  readonly resultsBucket?: s3.IBucket;

  /**
   * @default Duration.days(30)
   */
  readonly resultsExpiration?: Duration;

  /**
   * @default RemovalPolicy.RETAIN
   */
  readonly removalPolicy?: RemovalPolicy;
}

/** Options for LogAnalytics.fromExtension. */
export interface LogAnalyticsFromExtensionOptions {
  readonly databaseName: string;

  /**
   * @default 'app_logs'
   */
  readonly tableName?: string;

  /**
   * @default true
   */
  readonly createDatabase?: boolean;

  /**
   * @default - the current year through the current year plus five
   */
  readonly projectionYearRange?: ProjectionYearRange;

  /**
   * @default true
   */
  readonly createWorkgroup?: boolean;

  /**
   * @default - derived from the database name
   */
  readonly workgroupName?: string;

  /**
   * @default - a bucket is created with a 30 day expiry
   */
  readonly resultsBucket?: s3.IBucket;

  /**
   * @default Duration.days(30)
   */
  readonly resultsExpiration?: Duration;

  /**
   * @default RemovalPolicy.RETAIN
   */
  readonly removalPolicy?: RemovalPolicy;
}

/**
 * A Glue table and Athena workgroup for querying the Parquet files the
 * extension writes.
 *
 * Partitions are resolved with Athena partition projection, computed from the
 * key layout at query time. Nothing has to register partitions, so there is no
 * crawler, no scheduled MSCK REPAIR and no broad glue:* permission anywhere.
 *
 * Because projection derives the S3 location from a template, every directory
 * level below the prefix must be a partition key - which is exactly the layout
 * the extension writes.
 *
 * @example
 * const analytics = LogAnalytics.fromExtension(this, 'Analytics', ext, {
 *   databaseName: 'my_app_logs',
 * });
 */
export class LogAnalytics extends Construct {
  /**
   * Build the analytics stack for an extension, taking the bucket and the key
   * prefix from it so the two cannot drift apart.
   */
  public static fromExtension(
    scope: Construct,
    id: string,
    extension: ILogToS3Extension,
    options: LogAnalyticsFromExtensionOptions,
  ): LogAnalytics {
    return new LogAnalytics(scope, id, {
      ...options,
      logsBucket: extension.logsBucket,
      keyPrefix: extension.keyPrefix,
    });
  }

  public readonly databaseName: string;
  public readonly tableName: string;
  /** The created database, or undefined when createDatabase was false. */
  public readonly database?: glue.CfnDatabase;
  public readonly table: glue.CfnTable;
  /** The created workgroup, or undefined when createWorkgroup was false. */
  public readonly workgroup?: athena.CfnWorkGroup;
  public readonly resultsBucket: s3.IBucket;

  private readonly keyPrefix: string;
  private readonly logsBucket: s3.IBucket;

  constructor(scope: Construct, id: string, props: LogAnalyticsProps) {
    super(scope, id);

    const stack = Stack.of(this);
    this.databaseName = props.databaseName;
    this.tableName = props.tableName ?? "app_logs";
    this.keyPrefix = normalizePrefix(props.keyPrefix ?? DEFAULT_KEY_PREFIX);
    this.logsBucket = props.logsBucket;

    if (props.createDatabase ?? true) {
      this.database = new glue.CfnDatabase(this, "Database", {
        catalogId: stack.account,
        databaseInput: {
          name: this.databaseName,
          description: `Logs written by the log-to-s3 Lambda extension (${this.tableName})`,
        },
      });
    }

    const location = `s3://${props.logsBucket.bucketName}/${this.keyPrefix}`;
    // The placeholders below are literal Athena syntax, not TypeScript
    // interpolation. The template must name every partition column and end
    // with a slash, which is why the key layout has no other directory levels.
    const locationTemplate =
      location + "year=${year}/month=${month}/day=${day}/hour=${hour}/";

    const currentYear = new Date().getUTCFullYear();
    const years = props.projectionYearRange ?? {
      start: currentYear,
      end: currentYear + 5,
    };
    if (years.end < years.start) {
      throw new Error(
        `LogAnalytics: projectionYearRange end (${years.end}) must not be before start (${years.start}).`,
      );
    }

    this.table = new glue.CfnTable(this, "Table", {
      catalogId: stack.account,
      databaseName: this.databaseName,
      tableInput: {
        name: this.tableName,
        tableType: "EXTERNAL_TABLE",
        parameters: {
          EXTERNAL: "TRUE",
          classification: "parquet",
          typeOfData: "file",
          "parquet.compression": "SNAPPY",

          "projection.enabled": "true",
          "projection.year.type": "integer",
          "projection.year.range": `${years.start},${years.end}`,
          "projection.month.type": "integer",
          "projection.month.range": "1,12",
          "projection.month.digits": "2",
          "projection.day.type": "integer",
          "projection.day.range": "1,31",
          "projection.day.digits": "2",
          "projection.hour.type": "integer",
          "projection.hour.range": "0,23",
          "projection.hour.digits": "2",
          "storage.location.template": locationTemplate,
        },
        // String, not int: the digits setting zero-pads the projected values,
        // and '05' only round-trips through a string column.
        partitionKeys: PARTITION_COLUMNS.map((c) => ({
          name: c.name,
          type: c.type,
        })),
        storageDescriptor: {
          location,
          inputFormat:
            "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat",
          outputFormat:
            "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat",
          compressed: false,
          serdeInfo: {
            serializationLibrary:
              "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe",
            parameters: { "serialization.format": "1" },
          },
          columns: LOG_COLUMNS.map((c) => ({ name: c.name, type: c.type })),
        },
      },
    });

    if (this.database) {
      this.table.addDependency(this.database);
    }

    this.resultsBucket =
      props.resultsBucket ??
      new s3.Bucket(this, "Results", {
        encryption: s3.BucketEncryption.S3_MANAGED,
        blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
        enforceSSL: true,
        removalPolicy: props.removalPolicy ?? RemovalPolicy.RETAIN,
        lifecycleRules: [
          {
            id: "expire-athena-results",
            enabled: true,
            expiration: props.resultsExpiration ?? Duration.days(30),
          },
        ],
      });

    if (props.createWorkgroup ?? true) {
      this.workgroup = new athena.CfnWorkGroup(this, "Workgroup", {
        name: props.workgroupName ?? `${this.databaseName}-wg`,
        state: "ENABLED",
        recursiveDeleteOption: true,
        workGroupConfiguration: {
          enforceWorkGroupConfiguration: true,
          publishCloudWatchMetricsEnabled: true,
          engineVersion: { selectedEngineVersion: "AUTO" },
          resultConfiguration: {
            outputLocation: `s3://${this.resultsBucket.bucketName}/athena-results/`,
            encryptionConfiguration: { encryptionOption: "SSE_S3" },
          },
        },
      });
    }
  }

  /**
   * Grant everything needed to run a query against this table: the workgroup,
   * the catalog entries, read access to the log prefix and read/write on the
   * results bucket. Every statement is scoped to a resource.
   */
  public grantQuery(grantee: iam.IGrantable): void {
    const stack = Stack.of(this);

    if (this.workgroup) {
      grantee.grantPrincipal.addToPrincipalPolicy(
        new iam.PolicyStatement({
          actions: [
            "athena:StartQueryExecution",
            "athena:StopQueryExecution",
            "athena:GetQueryExecution",
            "athena:GetQueryResults",
            "athena:GetWorkGroup",
          ],
          resources: [
            stack.formatArn({
              service: "athena",
              resource: "workgroup",
              resourceName: this.workgroup.name,
            }),
          ],
        }),
      );
    }

    grantee.grantPrincipal.addToPrincipalPolicy(
      new iam.PolicyStatement({
        actions: [
          "glue:GetDatabase",
          "glue:GetDatabases",
          "glue:GetTable",
          "glue:GetTables",
          "glue:GetPartitions",
        ],
        resources: [
          stack.formatArn({ service: "glue", resource: "catalog" }),
          stack.formatArn({
            service: "glue",
            resource: "database",
            resourceName: this.databaseName,
          }),
          stack.formatArn({
            service: "glue",
            resource: "table",
            resourceName: `${this.databaseName}/${this.tableName}`,
          }),
        ],
      }),
    );

    this.logsBucket.grantRead(grantee, `${this.keyPrefix}*`);
    this.resultsBucket.grantReadWrite(grantee, "athena-results/*");
  }
}
