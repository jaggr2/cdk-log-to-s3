import { App, Duration, RemovalPolicy, Stack } from "aws-cdk-lib";
import { Match, Template } from "aws-cdk-lib/assertions";
import * as s3 from "aws-cdk-lib/aws-s3";
import { LogBucket } from "../src";

function synth(fn: (stack: Stack) => void): Template {
  const stack = new Stack(new App(), "Test");
  fn(stack);
  return Template.fromStack(stack);
}

describe("LogBucket", () => {
  test("is private, encrypted and TLS-only by default", () => {
    const template = synth((stack) => {
      new LogBucket(stack, "Logs");
    });

    template.hasResourceProperties("AWS::S3::Bucket", {
      BucketEncryption: {
        ServerSideEncryptionConfiguration: [
          { ServerSideEncryptionByDefault: { SSEAlgorithm: "AES256" } },
        ],
      },
      PublicAccessBlockConfiguration: {
        BlockPublicAcls: true,
        BlockPublicPolicy: true,
        IgnorePublicAcls: true,
        RestrictPublicBuckets: true,
      },
    });

    // enforceSSL adds a deny statement for non-TLS requests.
    template.hasResourceProperties("AWS::S3::BucketPolicy", {
      PolicyDocument: {
        Statement: Match.arrayWith([
          Match.objectLike({
            Effect: "Deny",
            Condition: { Bool: { "aws:SecureTransport": "false" } },
          }),
        ]),
      },
    });
  });

  test("retains the bucket so cdk destroy cannot delete an audit trail", () => {
    const template = synth((stack) => {
      new LogBucket(stack, "Logs");
    });

    template.hasResource("AWS::S3::Bucket", {
      DeletionPolicy: "Retain",
      UpdateReplacePolicy: "Retain",
    });
  });

  test("applies the default tiered lifecycle", () => {
    const template = synth((stack) => {
      new LogBucket(stack, "Logs");
    });

    template.hasResourceProperties("AWS::S3::Bucket", {
      LifecycleConfiguration: {
        Rules: [
          {
            Id: "log-to-s3-lifecycle",
            Status: "Enabled",
            ExpirationInDays: 180,
            AbortIncompleteMultipartUpload: { DaysAfterInitiation: 7 },
            Transitions: [
              { StorageClass: "STANDARD_IA", TransitionInDays: 30 },
              { StorageClass: "GLACIER", TransitionInDays: 90 },
            ],
          },
        ],
      },
    });
  });

  test("honours the lifecycle shorthand", () => {
    const template = synth((stack) => {
      new LogBucket(stack, "Logs", {
        infrequentAccessAfter: Duration.days(10),
        glacierAfter: Duration.days(0), // disabled
        expireAfter: Duration.days(45),
      });
    });

    template.hasResourceProperties("AWS::S3::Bucket", {
      LifecycleConfiguration: {
        Rules: [
          Match.objectLike({
            ExpirationInDays: 45,
            Transitions: [
              { StorageClass: "STANDARD_IA", TransitionInDays: 10 },
            ],
          }),
        ],
      },
    });
  });

  test("lets a caller replace the lifecycle wholesale", () => {
    const template = synth((stack) => {
      new LogBucket(stack, "Logs", {
        lifecycleRules: [{ id: "mine", expiration: Duration.days(3) }],
      });
    });

    template.hasResourceProperties("AWS::S3::Bucket", {
      LifecycleConfiguration: {
        Rules: [Match.objectLike({ Id: "mine", ExpirationInDays: 3 })],
      },
    });
  });

  test("refuses to combine lifecycleRules with the shorthand", () => {
    expect(() =>
      synth((stack) => {
        new LogBucket(stack, "Logs", {
          lifecycleRules: [{ id: "mine", expiration: Duration.days(3) }],
          expireAfter: Duration.days(45),
        });
      }),
    ).toThrow(/either `lifecycleRules` or the shorthand properties/);
  });

  test("normalises the key prefix", () => {
    const stack = new Stack(new App(), "Test");
    expect(new LogBucket(stack, "A").keyPrefix).toBe("logs/");
    expect(
      new LogBucket(stack, "B", { keyPrefix: "/nested/dir" }).keyPrefix,
    ).toBe("nested/dir/");
    expect(new LogBucket(stack, "C", { keyPrefix: "" }).keyPrefix).toBe("");
  });

  test("is usable anywhere an IBucket is expected", () => {
    const stack = new Stack(new App(), "Test");
    const bucket: s3.IBucket = new LogBucket(stack, "Logs");
    expect(bucket.bucketName).toBeDefined();
  });

  test("caller props override the secure defaults when asked", () => {
    const template = synth((stack) => {
      new LogBucket(stack, "Logs", {
        versioned: true,
        removalPolicy: RemovalPolicy.DESTROY,
      });
    });

    template.hasResourceProperties("AWS::S3::Bucket", {
      VersioningConfiguration: { Status: "Enabled" },
    });
    template.hasResource("AWS::S3::Bucket", { DeletionPolicy: "Delete" });
  });
});
