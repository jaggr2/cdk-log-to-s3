import * as path from "path";
import { App, Duration, Size, Stack } from "aws-cdk-lib";
import { Match, Template } from "aws-cdk-lib/assertions";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as s3 from "aws-cdk-lib/aws-s3";
import { LogBucket, LogCompression, LogLevel, LogToS3Extension } from "../src";

function testStack(): Stack {
  return new Stack(new App(), "Test", {
    env: { account: "123456789012", region: "eu-central-1" },
  });
}

function testFunction(
  stack: Stack,
  id: string,
  architecture: lambda.Architecture = lambda.Architecture.ARM_64,
): lambda.Function {
  return new lambda.Function(stack, id, {
    runtime: lambda.Runtime.NODEJS_20_X,
    handler: "index.handler",
    code: lambda.Code.fromInline("exports.handler = async () => {};"),
    architecture,
  });
}

/** All IAM policy statements in the synthesized template, flattened. */
function policyStatements(template: Template): any[] {
  const policies = template.findResources("AWS::IAM::Policy");
  return Object.values(policies).flatMap(
    (p: any) => p.Properties.PolicyDocument.Statement,
  );
}

function putObjectStatements(template: Template): any[] {
  return policyStatements(template).filter((s) => {
    const actions = Array.isArray(s.Action) ? s.Action : [s.Action];
    return actions.includes("s3:PutObject");
  });
}

describe("LogToS3Extension", () => {
  test("creates an arm64 layer by default", () => {
    const stack = testStack();
    new LogToS3Extension(stack, "Ext", {
      logsBucket: new LogBucket(stack, "Logs"),
    });

    Template.fromStack(stack).hasResourceProperties(
      "AWS::Lambda::LayerVersion",
      {
        CompatibleArchitectures: ["arm64"],
      },
    );
  });

  test("selects a different asset for x86_64", () => {
    const armStack = testStack();
    new LogToS3Extension(armStack, "Ext", {
      logsBucket: new LogBucket(armStack, "Logs"),
      architecture: lambda.Architecture.ARM_64,
    });

    const x86Stack = testStack();
    new LogToS3Extension(x86Stack, "Ext", {
      logsBucket: new LogBucket(x86Stack, "Logs"),
      architecture: lambda.Architecture.X86_64,
    });

    const armLayer: any = Object.values(
      Template.fromStack(armStack).findResources("AWS::Lambda::LayerVersion"),
    )[0];
    const x86Layer: any = Object.values(
      Template.fromStack(x86Stack).findResources("AWS::Lambda::LayerVersion"),
    )[0];

    expect(x86Layer.Properties.CompatibleArchitectures).toEqual(["x86_64"]);
    // Different binaries must produce different assets; an equal hash would
    // mean both architectures ship the same executable.
    expect(armLayer.Properties.Content.S3Key).not.toEqual(
      x86Layer.Properties.Content.S3Key,
    );
  });

  test("leaves compatibleRuntimes unset so any runtime can use the extension", () => {
    const stack = testStack();
    new LogToS3Extension(stack, "Ext", {
      logsBucket: new LogBucket(stack, "Logs"),
    });

    const layer: any = Object.values(
      Template.fromStack(stack).findResources("AWS::Lambda::LayerVersion"),
    )[0];
    expect(layer.Properties.CompatibleRuntimes).toBeUndefined();
  });

  test("attachTo injects the extension environment", () => {
    const stack = testStack();
    const ext = new LogToS3Extension(stack, "Ext", {
      logsBucket: new LogBucket(stack, "Logs"),
      logLevel: LogLevel.WARN,
      compression: LogCompression.ZSTD,
      keyPrefix: "app-logs/",
      flushInterval: Duration.seconds(30),
      maxBufferSize: Size.mebibytes(4),
      telemetryPort: 2021,
      includePlatformReport: false,
      extensionDebug: true,
    });
    ext.attachTo(testFunction(stack, "Fn"));

    Template.fromStack(stack).hasResourceProperties("AWS::Lambda::Function", {
      Environment: {
        Variables: Match.objectLike({
          LOG_TO_S3_PREFIX: "app-logs/",
          LOG_TO_S3_LEVEL: "WARN",
          LOG_TO_S3_COMPRESSION: "zstd",
          LOG_TO_S3_FLUSH_INTERVAL_SECONDS: "30",
          LOG_TO_S3_MAX_BUFFER_BYTES: "4194304",
          LOG_TO_S3_TELEMETRY_PORT: "2021",
          LOG_TO_S3_INCLUDE_PLATFORM_REPORT: "false",
          LOG_TO_S3_DEBUG: "true",
        }),
      },
    });
  });

  test("omits optional environment variables so the Go defaults apply", () => {
    const stack = testStack();
    const ext = new LogToS3Extension(stack, "Ext", {
      logsBucket: new LogBucket(stack, "Logs"),
    });
    ext.attachTo(testFunction(stack, "Fn"));

    const fn: any = Object.values(
      Template.fromStack(stack).findResources("AWS::Lambda::Function"),
    )[0];
    const vars = fn.Properties.Environment.Variables;

    expect(Object.keys(vars).sort()).toEqual([
      "LOG_TO_S3_BUCKET",
      "LOG_TO_S3_COMPRESSION",
      "LOG_TO_S3_LEVEL",
      "LOG_TO_S3_PREFIX",
    ]);
  });

  test("a per-function log level overrides the extension default", () => {
    const stack = testStack();
    const ext = new LogToS3Extension(stack, "Ext", {
      logsBucket: new LogBucket(stack, "Logs"),
      logLevel: LogLevel.INFO,
    });
    ext.attachTo(testFunction(stack, "Fn"), { logLevel: LogLevel.DEBUG });

    Template.fromStack(stack).hasResourceProperties("AWS::Lambda::Function", {
      Environment: {
        Variables: Match.objectLike({ LOG_TO_S3_LEVEL: "DEBUG" }),
      },
    });
  });

  // Regression: the pre-extraction code assigned CfnFunction.layers directly,
  // which replaced any layer the function already had.
  test("attachTo preserves layers the function already had", () => {
    const stack = testStack();
    const fn = testFunction(stack, "Fn");
    fn.addLayers(
      new lambda.LayerVersion(stack, "Other", {
        // Layers reject inline code, so any real directory will do here.
        code: lambda.Code.fromAsset(path.join(__dirname, "..", "assets")),
      }),
    );

    new LogToS3Extension(stack, "Ext", {
      logsBucket: new LogBucket(stack, "Logs"),
    }).attachTo(fn);

    const synthesized: any = Object.values(
      Template.fromStack(stack).findResources("AWS::Lambda::Function"),
    )[0];
    expect(synthesized.Properties.Layers).toHaveLength(2);
  });

  test("attachTo is idempotent", () => {
    const stack = testStack();
    const fn = testFunction(stack, "Fn");
    const ext = new LogToS3Extension(stack, "Ext", {
      logsBucket: new LogBucket(stack, "Logs"),
    });

    ext.attachTo(fn);
    ext.attachTo(fn);

    const synthesized: any = Object.values(
      Template.fromStack(stack).findResources("AWS::Lambda::Function"),
    )[0];
    expect(synthesized.Properties.Layers).toHaveLength(1);
  });

  // Regression: the pre-extraction construct granted s3:PutObject on the whole
  // bucket, and against a different bucket than the one the function was told
  // to write to.
  test("grants PutObject scoped to the key prefix and nothing wider", () => {
    const stack = testStack();
    const ext = new LogToS3Extension(stack, "Ext", {
      logsBucket: new LogBucket(stack, "Logs"),
      keyPrefix: "logs/",
    });
    ext.attachTo(testFunction(stack, "Fn"));

    const statements = putObjectStatements(Template.fromStack(stack));
    expect(statements).toHaveLength(1);

    const resource = JSON.stringify(statements[0].Resource);
    expect(resource).toContain("/logs/*");
    expect(resource).not.toContain('","/*"');
  });

  test("the granted bucket is the bucket named in the environment", () => {
    const stack = testStack();
    const bucket = new LogBucket(stack, "Logs");
    const other = new s3.Bucket(stack, "Other");
    new LogToS3Extension(stack, "Ext", { logsBucket: bucket }).attachTo(
      testFunction(stack, "Fn"),
    );

    const template = Template.fromStack(stack);
    const fn: any = Object.values(
      template.findResources("AWS::Lambda::Function"),
    )[0];
    const grantedResource = JSON.stringify(
      putObjectStatements(template)[0].Resource,
    );

    // Both must reference the same logical bucket, which is the bug that made
    // the extension write somewhere other than where it had been granted.
    const bucketRef = bucket.node.defaultChild!.node.id;
    expect(
      JSON.stringify(fn.Properties.Environment.Variables.LOG_TO_S3_BUCKET),
    ).toContain(stack.resolve(bucket.bucketName).Ref);
    expect(grantedResource).toContain(
      stack.resolve(bucket.bucketArn)["Fn::GetAtt"][0],
    );
    expect(grantedResource).not.toContain(
      stack.resolve(other.bucketArn)["Fn::GetAtt"][0],
    );
    expect(bucketRef).toBeDefined();
  });

  test("grantWrite:false skips the grant", () => {
    const stack = testStack();
    const ext = new LogToS3Extension(stack, "Ext", {
      logsBucket: new LogBucket(stack, "Logs"),
    });
    ext.attachTo(testFunction(stack, "Fn"), { grantWrite: false });

    expect(putObjectStatements(Template.fromStack(stack))).toHaveLength(0);
  });

  test("grantWriteLogs grants without attaching a layer", () => {
    const stack = testStack();
    const fn = testFunction(stack, "Fn");
    new LogToS3Extension(stack, "Ext", {
      logsBucket: new LogBucket(stack, "Logs"),
    }).grantWriteLogs(fn);

    const template = Template.fromStack(stack);
    expect(putObjectStatements(template)).toHaveLength(1);

    const synthesized: any = Object.values(
      template.findResources("AWS::Lambda::Function"),
    )[0];
    expect(synthesized.Properties.Layers).toBeUndefined();
  });

  test("rejects an architecture mismatch and names both architectures", () => {
    const stack = testStack();
    const ext = new LogToS3Extension(stack, "Ext", {
      logsBucket: new LogBucket(stack, "Logs"),
      architecture: lambda.Architecture.ARM_64,
    });
    const fn = testFunction(stack, "Fn", lambda.Architecture.X86_64);

    expect(() => ext.attachTo(fn)).toThrow(/arm64.*x86_64|x86_64.*arm64/s);
    expect(() => ext.attachTo(fn)).toThrow(/Architecture\.X86_64/);
  });

  test("rejects an unsupported architecture", () => {
    const stack = testStack();
    expect(
      () =>
        new LogToS3Extension(stack, "Ext", {
          logsBucket: new LogBucket(stack, "Logs"),
          architecture: lambda.Architecture.custom("mips"),
        }),
    ).toThrow(/Unsupported architecture "mips"/);
  });

  test("rejects a sub-second flush interval", () => {
    const stack = testStack();
    expect(
      () =>
        new LogToS3Extension(stack, "Ext", {
          logsBucket: new LogBucket(stack, "Logs"),
          flushInterval: Duration.millis(500),
        }),
    ).toThrow(/at least one second/);
  });

  describe("fromAttributes", () => {
    test("uses the imported ARN and creates no new layer version", () => {
      const stack = testStack();
      const ext = LogToS3Extension.fromAttributes(stack, "Ext", {
        layerVersionArn:
          "arn:aws:lambda:eu-central-1:123456789012:layer:log-to-s3:7",
        logsBucket: s3.Bucket.fromBucketName(stack, "Logs", "existing-logs"),
      });
      ext.attachTo(testFunction(stack, "Fn"));

      const template = Template.fromStack(stack);
      template.resourceCountIs("AWS::Lambda::LayerVersion", 0);
      template.hasResourceProperties("AWS::Lambda::Function", {
        Layers: ["arn:aws:lambda:eu-central-1:123456789012:layer:log-to-s3:7"],
      });
    });

    test("still injects the environment and grants the prefix", () => {
      const stack = testStack();
      const ext = LogToS3Extension.fromAttributes(stack, "Ext", {
        layerVersionArn:
          "arn:aws:lambda:eu-central-1:123456789012:layer:log-to-s3:7",
        logsBucket: s3.Bucket.fromBucketName(stack, "Logs", "existing-logs"),
        keyPrefix: "shared/",
        logLevel: LogLevel.ERROR,
      });
      ext.attachTo(testFunction(stack, "Fn"));

      const template = Template.fromStack(stack);
      template.hasResourceProperties("AWS::Lambda::Function", {
        Environment: {
          Variables: Match.objectLike({
            LOG_TO_S3_BUCKET: "existing-logs",
            LOG_TO_S3_PREFIX: "shared/",
            LOG_TO_S3_LEVEL: "ERROR",
          }),
        },
      });
      expect(
        JSON.stringify(putObjectStatements(template)[0].Resource),
      ).toContain("/shared/*");
    });
  });

  test("two extensions in one stack share the same asset", () => {
    const stack = testStack();
    const bucket = new LogBucket(stack, "Logs");
    new LogToS3Extension(stack, "ExtA", { logsBucket: bucket });
    new LogToS3Extension(stack, "ExtB", { logsBucket: bucket });

    const layers = Object.values(
      Template.fromStack(stack).findResources("AWS::Lambda::LayerVersion"),
    ) as any[];

    expect(layers).toHaveLength(2);
    // Same asset hash, so CDK uploads the zip once no matter how many stacks
    // declare the layer. That is what makes per-stack instantiation cheap.
    expect(layers[0].Properties.Content.S3Key).toEqual(
      layers[1].Properties.Content.S3Key,
    );
  });
});
