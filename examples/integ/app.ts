/**
 * A throwaway stack used to verify the whole path for real: a function writes
 * structured logs, the extension turns them into Parquet in S3, and Athena
 * reads them back through partition projection.
 *
 * Partition projection cannot be verified by a synth test - the assertions can
 * only check that the location template is the string we intended, not that
 * Athena resolves it. This app plus scripts/verify-integ.ts is what actually
 * proves it works.
 *
 *   npx projen integ:deploy
 *   npx projen integ:verify
 *   npx projen integ:destroy
 */
import { App, CfnOutput, Duration, RemovalPolicy, Stack, StackProps } from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as logs from 'aws-cdk-lib/aws-logs';
import { Construct } from 'constructs';
import { LogAnalytics, LogBucket, LogLevel, LogToS3Extension } from '../../src';

class IntegStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    // DESTROY throughout: this stack exists to be torn down again.
    const logsBucket = new LogBucket(this, 'Logs', {
      removalPolicy: RemovalPolicy.DESTROY,
      autoDeleteObjects: true,
      expireAfter: Duration.days(1),
      infrequentAccessAfter: Duration.days(0),
      glacierAfter: Duration.days(0),
    });

    const extension = new LogToS3Extension(this, 'Extension', {
      logsBucket,
      logLevel: LogLevel.DEBUG,
      // Short enough that verification does not have to wait 15 seconds.
      flushInterval: Duration.seconds(5),
      extensionDebug: true,
      removalPolicy: RemovalPolicy.DESTROY,
    });

    const fn = new lambda.Function(this, 'Emitter', {
      runtime: lambda.Runtime.NODEJS_20_X,
      architecture: lambda.Architecture.ARM_64,
      handler: 'index.handler',
      timeout: Duration.seconds(30),
      code: lambda.Code.fromAsset(`${__dirname}/handler`),
      // An explicit group rather than logRetention, which provisions a
      // custom-resource Lambda just to set one property.
      logGroup: new logs.LogGroup(this, 'EmitterLogs', {
        retention: logs.RetentionDays.ONE_DAY,
        removalPolicy: RemovalPolicy.DESTROY,
      }),
    });
    extension.attachTo(fn);

    const analytics = LogAnalytics.fromExtension(this, 'Analytics', extension, {
      databaseName: 'log_to_s3_integ',
      tableName: 'app_logs',
      removalPolicy: RemovalPolicy.DESTROY,
      projectionYearRange: { start: new Date().getUTCFullYear(), end: new Date().getUTCFullYear() + 1 },
    });

    new CfnOutput(this, 'FunctionName', { value: fn.functionName });
    new CfnOutput(this, 'LogsBucketName', { value: logsBucket.bucketName });
    new CfnOutput(this, 'KeyPrefix', { value: extension.keyPrefix });
    new CfnOutput(this, 'DatabaseName', { value: analytics.databaseName });
    new CfnOutput(this, 'TableName', { value: analytics.tableName });
    new CfnOutput(this, 'WorkgroupName', { value: analytics.workgroup!.name });
  }
}

const app = new App();
new IntegStack(app, 'LogToS3Integ');
