import { App, Duration, Stack } from "aws-cdk-lib";
import { Match, Template } from "aws-cdk-lib/assertions";
import * as iam from "aws-cdk-lib/aws-iam";
import { LogAnalytics, LogBucket, LogToS3Extension } from "../src";

function testStack(): Stack {
  return new Stack(new App(), "Test", {
    env: { account: "123456789012", region: "eu-central-1" },
  });
}

/** The column list the Go writer produces, in order. */
const EXPECTED_COLUMNS = [
  "timestamp",
  "level",
  "source",
  "correlation_id",
  "request_id",
  "message",
  "function_name",
  "context",
  "stack_trace",
  "caller",
];

describe("LogAnalytics", () => {
  test("builds a location template that names every partition and ends with a slash", () => {
    const stack = testStack();
    new LogAnalytics(stack, "Analytics", {
      logsBucket: new LogBucket(stack, "Logs"),
      databaseName: "my_logs",
    });

    const table: any = Object.values(
      Template.fromStack(stack).findResources("AWS::Glue::Table"),
    )[0];
    const template =
      table.Properties.TableInput.Parameters["storage.location.template"];
    const rendered = JSON.stringify(template);

    // Athena requires a placeholder for every partition column, and the
    // template must end in a slash. This is the single most breakage-prone
    // value in the library.
    expect(rendered).toContain("logs/${dt}/");
    // The old four-column layout must not linger anywhere.
    expect(rendered).not.toContain("year=");
    expect(rendered).not.toContain("hour=");
  });

  test("declares the full projection configuration", () => {
    const stack = testStack();
    new LogAnalytics(stack, "Analytics", {
      logsBucket: new LogBucket(stack, "Logs"),
      databaseName: "my_logs",
      projectionWindow: Duration.days(365),
    });

    Template.fromStack(stack).hasResourceProperties("AWS::Glue::Table", {
      TableInput: Match.objectLike({
        TableType: "EXTERNAL_TABLE",
        Parameters: Match.objectLike({
          EXTERNAL: "TRUE",
          classification: "parquet",
          "parquet.compression": "SNAPPY",
          "projection.enabled": "true",
          "projection.dt.type": "date",
          "projection.dt.format": "yyyy/MM/dd",
          // Sliding, so the partition count stays bounded instead of growing
          // a partition per day forever.
          "projection.dt.range": "NOW-365DAYS,NOW",
          "projection.dt.interval": "1",
          "projection.dt.interval.unit": "DAYS",
        }),
      }),
    });
  });

  test("defaults to a two year window", () => {
    const stack = testStack();
    new LogAnalytics(stack, "Analytics", {
      logsBucket: new LogBucket(stack, "Logs"),
      databaseName: "my_logs",
    });

    Template.fromStack(stack).hasResourceProperties("AWS::Glue::Table", {
      TableInput: Match.objectLike({
        Parameters: Match.objectLike({
          "projection.dt.range": "NOW-730DAYS,NOW",
        }),
      }),
    });
  });

  test("the columns match the schema the Go writer produces", () => {
    const stack = testStack();
    new LogAnalytics(stack, "Analytics", {
      logsBucket: new LogBucket(stack, "Logs"),
      databaseName: "my_logs",
    });

    const table: any = Object.values(
      Template.fromStack(stack).findResources("AWS::Glue::Table"),
    )[0];
    const columns = table.Properties.TableInput.StorageDescriptor.Columns;

    expect(columns.map((c: any) => c.Name)).toEqual(EXPECTED_COLUMNS);
    expect(columns.every((c: any) => c.Type === "string")).toBe(true);
  });

  test("has a single string partition key", () => {
    const stack = testStack();
    new LogAnalytics(stack, "Analytics", {
      logsBucket: new LogBucket(stack, "Logs"),
      databaseName: "my_logs",
    });

    // String, not date: the projected value carries the slashes of the
    // yyyy/MM/dd path format, which a date column would not round-trip.
    Template.fromStack(stack).hasResourceProperties("AWS::Glue::Table", {
      TableInput: Match.objectLike({
        PartitionKeys: [{ Name: "dt", Type: "string" }],
      }),
    });
  });

  // Regression: the pre-extraction setup registered partitions with a daily
  // cron running MSCK REPAIR from an inline-code Lambda holding glue:* on *.
  test("needs no Lambda, no schedule and no partition registration", () => {
    const stack = testStack();
    new LogAnalytics(stack, "Analytics", {
      logsBucket: new LogBucket(stack, "Logs"),
      databaseName: "my_logs",
    });

    const template = Template.fromStack(stack);
    template.resourceCountIs("AWS::Lambda::Function", 0);
    template.resourceCountIs("AWS::Events::Rule", 0);
    template.resourceCountIs("AWS::Glue::Partition", 0);
  });

  test("creates the database by default and can skip it", () => {
    const withDb = testStack();
    new LogAnalytics(withDb, "Analytics", {
      logsBucket: new LogBucket(withDb, "Logs"),
      databaseName: "my_logs",
    });
    Template.fromStack(withDb).resourceCountIs("AWS::Glue::Database", 1);

    // createDatabase:false is the migration path: AWS::Glue::Database fails
    // when the database was already created outside CloudFormation.
    const withoutDb = testStack();
    new LogAnalytics(withoutDb, "Analytics", {
      logsBucket: new LogBucket(withoutDb, "Logs"),
      databaseName: "my_logs",
      createDatabase: false,
    });
    Template.fromStack(withoutDb).resourceCountIs("AWS::Glue::Database", 0);
    Template.fromStack(withoutDb).resourceCountIs("AWS::Glue::Table", 1);
  });

  test("configures an enforcing workgroup with encrypted results", () => {
    const stack = testStack();
    new LogAnalytics(stack, "Analytics", {
      logsBucket: new LogBucket(stack, "Logs"),
      databaseName: "my_logs",
    });

    Template.fromStack(stack).hasResourceProperties("AWS::Athena::WorkGroup", {
      Name: "my_logs-wg",
      State: "ENABLED",
      WorkGroupConfiguration: Match.objectLike({
        EnforceWorkGroupConfiguration: true,
        PublishCloudWatchMetricsEnabled: true,
        ResultConfiguration: Match.objectLike({
          EncryptionConfiguration: { EncryptionOption: "SSE_S3" },
        }),
      }),
    });
  });

  test("can skip the workgroup and reuse a results bucket", () => {
    const stack = testStack();
    new LogAnalytics(stack, "Analytics", {
      logsBucket: new LogBucket(stack, "Logs"),
      resultsBucket: new LogBucket(stack, "Results"),
      databaseName: "my_logs",
      createWorkgroup: false,
    });

    const template = Template.fromStack(stack);
    template.resourceCountIs("AWS::Athena::WorkGroup", 0);
    // Logs bucket and results bucket only; none created for Athena output.
    template.resourceCountIs("AWS::S3::Bucket", 2);
  });

  test("rejects a sub-day projection window", () => {
    const stack = testStack();
    expect(
      () =>
        new LogAnalytics(stack, "Analytics", {
          logsBucket: new LogBucket(stack, "Logs"),
          databaseName: "my_logs",
          projectionWindow: Duration.hours(12),
        }),
    ).toThrow(/at least one day/);
  });

  describe("grantQuery", () => {
    test("scopes every statement to a resource", () => {
      const stack = testStack();
      const analytics = new LogAnalytics(stack, "Analytics", {
        logsBucket: new LogBucket(stack, "Logs"),
        databaseName: "my_logs",
      });
      analytics.grantQuery(
        new iam.Role(stack, "Analyst", {
          assumedBy: new iam.AccountRootPrincipal(),
        }),
      );

      const statements = Object.values(
        Template.fromStack(stack).findResources("AWS::IAM::Policy"),
      ).flatMap((p: any) => p.Properties.PolicyDocument.Statement);

      expect(statements.length).toBeGreaterThan(0);
      for (const statement of statements) {
        // The pre-extraction stack used athena:* and glue:* on "*".
        expect(statement.Resource).not.toBe("*");
        expect(JSON.stringify(statement.Resource)).not.toBe('["*"]');
      }
    });

    test("covers athena, glue and both buckets", () => {
      const stack = testStack();
      const analytics = new LogAnalytics(stack, "Analytics", {
        logsBucket: new LogBucket(stack, "Logs"),
        databaseName: "my_logs",
      });
      analytics.grantQuery(
        new iam.Role(stack, "Analyst", {
          assumedBy: new iam.AccountRootPrincipal(),
        }),
      );

      const rendered = JSON.stringify(
        Object.values(
          Template.fromStack(stack).findResources("AWS::IAM::Policy"),
        ),
      );

      expect(rendered).toContain("athena:StartQueryExecution");
      expect(rendered).toContain("glue:GetPartitions");
      expect(rendered).toContain("s3:GetObject");
      expect(rendered).toContain(":workgroup/my_logs-wg");
      expect(rendered).toContain(":table/my_logs/app_logs");
    });
  });

  describe("fromExtension", () => {
    test("inherits the bucket and prefix so they cannot drift", () => {
      const stack = testStack();
      const ext = new LogToS3Extension(stack, "Ext", {
        logsBucket: new LogBucket(stack, "Logs"),
        keyPrefix: "custom/place",
      });
      LogAnalytics.fromExtension(stack, "Analytics", ext, {
        databaseName: "my_logs",
      });

      const table: any = Object.values(
        Template.fromStack(stack).findResources("AWS::Glue::Table"),
      )[0];
      const rendered = JSON.stringify(
        table.Properties.TableInput.Parameters["storage.location.template"],
      );

      // The extension normalises 'custom/place' to 'custom/place/'; the table
      // has to use exactly the same string.
      expect(rendered).toContain("custom/place/${dt}/");
      expect(ext.keyPrefix).toBe("custom/place/");
    });
  });
});
