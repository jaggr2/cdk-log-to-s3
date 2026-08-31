import { App, Duration, Size, Stack } from "aws-cdk-lib";
import { Match, Template } from "aws-cdk-lib/assertions";
import * as events from "aws-cdk-lib/aws-events";
import * as lambda from "aws-cdk-lib/aws-lambda";
import {
  LogBucket,
  LogCompaction,
  LogCompression,
  LogToS3Extension,
} from "../src";

function testStack(): Stack {
  return new Stack(new App(), "Test", {
    env: { account: "123456789012", region: "eu-central-1" },
  });
}

function policyStatements(template: Template): any[] {
  return Object.values(template.findResources("AWS::IAM::Policy")).flatMap(
    (p: any) => p.Properties.PolicyDocument.Statement,
  );
}

function allActions(template: Template): string[] {
  return policyStatements(template).flatMap((s) =>
    Array.isArray(s.Action) ? s.Action : [s.Action],
  );
}

describe("LogCompaction", () => {
  test("runs a Go binary on provided.al2023", () => {
    const stack = testStack();
    new LogCompaction(stack, "Compaction", {
      logsBucket: new LogBucket(stack, "Logs"),
    });

    Template.fromStack(stack).hasResourceProperties("AWS::Lambda::Function", {
      Runtime: "provided.al2023",
      // The build script places the executable at the archive root under this
      // exact name; provided.al2023 will not start without it.
      Handler: "bootstrap",
      Architectures: ["arm64"],
    });
  });

  test("schedules itself daily at 03:00 UTC by default", () => {
    const stack = testStack();
    new LogCompaction(stack, "Compaction", {
      logsBucket: new LogBucket(stack, "Logs"),
    });

    Template.fromStack(stack).hasResourceProperties("AWS::Events::Rule", {
      ScheduleExpression: "cron(0 3 * * ? *)",
      State: "ENABLED",
    });
  });

  test("accepts a custom schedule and can be left unscheduled", () => {
    const custom = testStack();
    new LogCompaction(custom, "Compaction", {
      logsBucket: new LogBucket(custom, "Logs"),
      schedule: events.Schedule.rate(Duration.hours(6)),
    });
    Template.fromStack(custom).hasResourceProperties("AWS::Events::Rule", {
      ScheduleExpression: "rate(6 hours)",
    });

    const manual = testStack();
    const compaction = new LogCompaction(manual, "Compaction", {
      logsBucket: new LogBucket(manual, "Logs"),
      enabled: false,
    });
    Template.fromStack(manual).resourceCountIs("AWS::Events::Rule", 0);
    expect(compaction.rule).toBeUndefined();
  });

  test("passes its configuration through the environment", () => {
    const stack = testStack();
    new LogCompaction(stack, "Compaction", {
      logsBucket: new LogBucket(stack, "Logs"),
      keyPrefix: "app-logs/",
      compression: LogCompression.ZSTD,
      lookback: Duration.days(3),
      minFilesPerPartition: 20,
      maxFilesPerRun: 500,
      maxBytesPerRun: Size.mebibytes(64),
      debug: true,
    });

    Template.fromStack(stack).hasResourceProperties("AWS::Lambda::Function", {
      Environment: {
        Variables: Match.objectLike({
          LOG_TO_S3_PREFIX: "app-logs/",
          LOG_TO_S3_COMPRESSION: "zstd",
          LOG_TO_S3_COMPACTION_LOOKBACK_DAYS: "3",
          LOG_TO_S3_COMPACTION_MIN_FILES: "20",
          LOG_TO_S3_COMPACTION_MAX_FILES: "500",
          LOG_TO_S3_COMPACTION_MAX_BYTES: "67108864",
          LOG_TO_S3_DEBUG: "true",
        }),
      },
    });
  });

  test("omits optional settings so the Go defaults apply", () => {
    const stack = testStack();
    new LogCompaction(stack, "Compaction", {
      logsBucket: new LogBucket(stack, "Logs"),
    });

    const fn: any = Object.values(
      Template.fromStack(stack).findResources("AWS::Lambda::Function"),
    )[0];

    expect(Object.keys(fn.Properties.Environment.Variables).sort()).toEqual([
      "LOG_TO_S3_BUCKET",
      "LOG_TO_S3_COMPACTION_LOOKBACK_DAYS",
      "LOG_TO_S3_COMPRESSION",
      "LOG_TO_S3_PREFIX",
    ]);
  });

  test("rejects a sub-day lookback", () => {
    const stack = testStack();
    expect(
      () =>
        new LogCompaction(stack, "Compaction", {
          logsBucket: new LogBucket(stack, "Logs"),
          lookback: Duration.hours(6),
        }),
    ).toThrow(/at least one day/);
  });

  describe("permissions", () => {
    test("can read, write and delete, all scoped to the key prefix", () => {
      const stack = testStack();
      new LogCompaction(stack, "Compaction", {
        logsBucket: new LogBucket(stack, "Logs"),
        keyPrefix: "logs/",
      });

      const template = Template.fromStack(stack);
      const actions = allActions(template);

      // CDK emits the wildcard forms (s3:GetObject*, s3:DeleteObject*).
      expect(actions.some((a) => a.startsWith("s3:GetObject"))).toBe(true);
      expect(actions.some((a) => a.startsWith("s3:PutObject"))).toBe(true);
      expect(actions.some((a) => a.startsWith("s3:DeleteObject"))).toBe(true);

      // Every object-level resource is confined to the prefix. The bare bucket
      // ARN is expected too: ListBucket is a bucket-level action and cannot be
      // scoped to a key.
      const objectArns = policyStatements(template)
        .flatMap((s) => (Array.isArray(s.Resource) ? s.Resource : [s.Resource]))
        .map((r) => JSON.stringify(r))
        .filter((r) => r.includes("Fn::Join"));

      expect(objectArns.length).toBeGreaterThan(0);
      for (const arn of objectArns) {
        expect(arn).toContain("/logs/*");
      }
    });

    // Compaction rewrites files inside a partition; the partition set never
    // changes. Under projection there is nothing to register, so MSCK REPAIR
    // has no role and the broad glue:* grant it needed must not come back.
    test("needs no Glue or Athena permissions at all", () => {
      const stack = testStack();
      new LogCompaction(stack, "Compaction", {
        logsBucket: new LogBucket(stack, "Logs"),
      });

      const rendered = JSON.stringify(
        Object.values(
          Template.fromStack(stack).findResources("AWS::IAM::Policy"),
        ),
      );
      expect(rendered).not.toContain("glue:");
      expect(rendered).not.toContain("athena:");
    });

    test("grants nothing on the whole bucket contents", () => {
      const stack = testStack();
      new LogCompaction(stack, "Compaction", {
        logsBucket: new LogBucket(stack, "Logs"),
        keyPrefix: "logs/",
      });

      for (const statement of policyStatements(Template.fromStack(stack))) {
        expect(JSON.stringify(statement.Resource)).not.toContain('","/*"');
        expect(statement.Resource).not.toBe("*");
      }
    });
  });

  describe("fromExtension", () => {
    test("inherits bucket, prefix and architecture", () => {
      const stack = testStack();
      const extension = new LogToS3Extension(stack, "Ext", {
        logsBucket: new LogBucket(stack, "Logs"),
        keyPrefix: "custom/place",
        architecture: lambda.Architecture.X86_64,
      });
      const compaction = LogCompaction.fromExtension(
        stack,
        "Compaction",
        extension,
      );

      expect(compaction.keyPrefix).toBe("custom/place/");
      Template.fromStack(stack).hasResourceProperties("AWS::Lambda::Function", {
        Runtime: "provided.al2023",
        Architectures: ["x86_64"],
        Environment: {
          Variables: Match.objectLike({ LOG_TO_S3_PREFIX: "custom/place/" }),
        },
      });
    });

    test("uses a different asset per architecture", () => {
      const arm = testStack();
      LogCompaction.fromExtension(
        arm,
        "Compaction",
        new LogToS3Extension(arm, "Ext", {
          logsBucket: new LogBucket(arm, "Logs"),
          architecture: lambda.Architecture.ARM_64,
        }),
      );

      const x86 = testStack();
      LogCompaction.fromExtension(
        x86,
        "Compaction",
        new LogToS3Extension(x86, "Ext", {
          logsBucket: new LogBucket(x86, "Logs"),
          architecture: lambda.Architecture.X86_64,
        }),
      );

      const keyOf = (stack: Stack) =>
        (
          Object.values(
            Template.fromStack(stack).findResources("AWS::Lambda::Function"),
          )[0] as any
        ).Properties.Code.S3Key;

      expect(keyOf(arm)).not.toEqual(keyOf(x86));
    });
  });

  test("gives the job its own log group with bounded retention", () => {
    const stack = testStack();
    new LogCompaction(stack, "Compaction", {
      logsBucket: new LogBucket(stack, "Logs"),
    });

    Template.fromStack(stack).hasResourceProperties("AWS::Logs::LogGroup", {
      RetentionInDays: 30,
    });
  });
});
