import { App, Stack } from "aws-cdk-lib";
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
    expect(rendered).toContain(
      "year=${year}/month=${month}/day=${day}/hour=${hour}/",
    );
    expect(rendered).toContain("logs/");
    for (const col of ["year", "month", "day", "hour"]) {
      expect(rendered).toContain("${" + col + "}");
    }
  });

  test("declares the full projection configuration", () => {
    const stack = testStack();
    new LogAnalytics(stack, "Analytics", {
      logsBucket: new LogBucket(stack, "Logs"),
      databaseName: "my_logs",
      projectionYearRange: { start: 2026, end: 2030 },
    });

    Template.fromStack(stack).hasResourceProperties("AWS::Glue::Table", {
      TableInput: Match.objectLike({
        TableType: "EXTERNAL_TABLE",
        Parameters: Match.objectLike({
          EXTERNAL: "TRUE",
          classification: "parquet",
          "parquet.compression": "SNAPPY",
          "projection.enabled": "true",
          "projection.year.type": "integer",
          "projection.year.range": "2026,2030",
          "projection.month.type": "integer",
          "projection.month.range": "1,12",
          "projection.month.digits": "2",
          "projection.day.type": "integer",
          "projection.day.range": "1,31",
          "projection.day.digits": "2",
          "projection.hour.type": "integer",
          "projection.hour.range": "0,23",
          "projection.hour.digits": "2",
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

  test("partition keys are strings so the zero-padded values round-trip", () => {
    const stack = testStack();
    new LogAnalytics(stack, "Analytics", {
      logsBucket: new LogBucket(stack, "Logs"),
      databaseName: "my_logs",
    });

    Template.fromStack(stack).hasResourceProperties("AWS::Glue::Table", {
      TableInput: Match.objectLike({
        PartitionKeys: [
          { Name: "year", Type: "string" },
          { Name: "month", Type: "string" },
          { Name: "day", Type: "string" },
          { Name: "hour", Type: "string" },
        ],
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

  test("rejects an inverted year range", () => {
    const stack = testStack();
    expect(
      () =>
        new LogAnalytics(stack, "Analytics", {
          logsBucket: new LogBucket(stack, "Logs"),
          databaseName: "my_logs",
          projectionYearRange: { start: 2030, end: 2026 },
        }),
    ).toThrow(/must not be before start/);
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
      expect(rendered).toContain("custom/place/year=${year}");
      expect(ext.keyPrefix).toBe("custom/place/");
    });
  });
});
