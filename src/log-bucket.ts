import { Duration, RemovalPolicy } from "aws-cdk-lib";
import * as s3 from "aws-cdk-lib/aws-s3";
import { Construct } from "constructs";
import { DEFAULT_KEY_PREFIX, normalizePrefix } from "./private/env";

export interface LogBucketProps extends s3.BucketProps {
  /**
   * Key prefix the extension writes under. Recorded on the construct so a
   * LogAnalytics table can be pointed at the same place without repeating it.
   *
   * @default 'logs/'
   */
  readonly keyPrefix?: string;

  /**
   * Transition objects to Infrequent Access after this long.
   *
   * @default Duration.days(30)
   */
  readonly infrequentAccessAfter?: Duration;

  /**
   * Transition objects to Glacier after this long.
   *
   * @default Duration.days(90)
   */
  readonly glacierAfter?: Duration;

  /**
   * Delete objects after this long.
   *
   * @default Duration.days(180)
   */
  readonly expireAfter?: Duration;

  /**
   * Abort incomplete multipart uploads after this long.
   *
   * @default Duration.days(7)
   */
  readonly abortIncompleteUploadsAfter?: Duration;
}

/**
 * An S3 bucket with defaults suited to Parquet log storage: private,
 * encrypted, TLS-only, and on a tiered lifecycle.
 *
 * It extends s3.Bucket, so it can be passed anywhere an IBucket is accepted -
 * including LogToS3ExtensionProps.logsBucket and LogAnalyticsProps.logsBucket.
 *
 * The removal policy is RETAIN. Logs usually outlive the stack that produced
 * them, and an accidental `cdk destroy` should not be able to delete an audit
 * trail. Pass `removalPolicy: RemovalPolicy.DESTROY` for throwaway stacks.
 */
export class LogBucket extends s3.Bucket {
  /** Normalised key prefix, e.g. 'logs/'. */
  public readonly keyPrefix: string;

  constructor(scope: Construct, id: string, props: LogBucketProps = {}) {
    if (props.lifecycleRules !== undefined && hasLifecycleShorthand(props)) {
      throw new Error(
        "LogBucket: pass either `lifecycleRules` or the shorthand properties " +
          "(infrequentAccessAfter / glacierAfter / expireAfter / abortIncompleteUploadsAfter), not both. " +
          "Supplying both would silently drop one of them.",
      );
    }

    super(scope, id, {
      encryption: s3.BucketEncryption.S3_MANAGED,
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      enforceSSL: true,
      versioned: false,
      removalPolicy: RemovalPolicy.RETAIN,
      ...props,
      lifecycleRules: props.lifecycleRules ?? buildLifecycleRules(props),
    });

    this.keyPrefix = normalizePrefix(props.keyPrefix ?? DEFAULT_KEY_PREFIX);
  }
}

function hasLifecycleShorthand(props: LogBucketProps): boolean {
  return (
    props.infrequentAccessAfter !== undefined ||
    props.glacierAfter !== undefined ||
    props.expireAfter !== undefined ||
    props.abortIncompleteUploadsAfter !== undefined
  );
}

/**
 * Builds the default lifecycle. A zero duration disables the individual step,
 * which is the escape hatch for "I want expiry but no Glacier transition"
 * without having to hand-write the whole rule set.
 */
function buildLifecycleRules(props: LogBucketProps): s3.LifecycleRule[] {
  const ia = props.infrequentAccessAfter ?? Duration.days(30);
  const glacier = props.glacierAfter ?? Duration.days(90);
  const expire = props.expireAfter ?? Duration.days(180);
  const abort = props.abortIncompleteUploadsAfter ?? Duration.days(7);

  const transitions: s3.Transition[] = [];
  if (ia.toSeconds() > 0) {
    transitions.push({
      storageClass: s3.StorageClass.INFREQUENT_ACCESS,
      transitionAfter: ia,
    });
  }
  if (glacier.toSeconds() > 0) {
    transitions.push({
      storageClass: s3.StorageClass.GLACIER,
      transitionAfter: glacier,
    });
  }

  const rule: s3.LifecycleRule = {
    id: "log-to-s3-lifecycle",
    enabled: true,
    transitions: transitions.length > 0 ? transitions : undefined,
    expiration: expire.toSeconds() > 0 ? expire : undefined,
    abortIncompleteMultipartUploadAfter:
      abort.toSeconds() > 0 ? abort : undefined,
  };

  // An all-zero configuration means "no lifecycle management at all".
  const empty =
    rule.transitions === undefined &&
    rule.expiration === undefined &&
    rule.abortIncompleteMultipartUploadAfter === undefined;

  return empty ? [] : [rule];
}
