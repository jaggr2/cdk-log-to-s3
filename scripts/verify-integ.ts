/**
 * Verifies a deployed integ stack end to end.
 *
 * This is the only check that proves partition projection actually resolves.
 * The synth tests can assert that storage.location.template is the string we
 * meant to write; only Athena can say whether it finds the files.
 *
 * Needs real AWS credentials. Run after `npx projen integ:deploy`.
 */
import {
  AthenaClient,
  GetQueryExecutionCommand,
  GetQueryResultsCommand,
  StartQueryExecutionCommand,
} from '@aws-sdk/client-athena';
import { CloudFormationClient, DescribeStacksCommand } from '@aws-sdk/client-cloudformation';
import { InvokeCommand, LambdaClient } from '@aws-sdk/client-lambda';
import { GetObjectCommand, ListObjectsV2Command, S3Client } from '@aws-sdk/client-s3';

const STACK_NAME = process.env.INTEG_STACK_NAME ?? 'LogToS3Integ';
const OBJECT_TIMEOUT_MS = 90_000;
const QUERY_TIMEOUT_MS = 120_000;

/** Keys must be flat inside the hour partition, or projection cannot see them. */
const KEY_PATTERN = /^logs\/year=\d{4}\/month=\d{2}\/day=\d{2}\/hour=\d{2}\/[A-Za-z0-9._-]+\.parquet$/;

const cfn = new CloudFormationClient({});
const lambda = new LambdaClient({});
const s3 = new S3Client({});
const athena = new AthenaClient({});

const failures: string[] = [];

function check(condition: boolean, description: string, detail?: string): void {
  if (condition) {
    console.log(`  ok    ${description}`);
  } else {
    console.error(`  FAIL  ${description}${detail ? `\n        ${detail}` : ''}`);
    failures.push(description);
  }
}

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

async function outputs(): Promise<Record<string, string>> {
  const result = await cfn.send(new DescribeStacksCommand({ StackName: STACK_NAME }));
  const stack = result.Stacks?.[0];
  if (!stack) throw new Error(`Stack ${STACK_NAME} not found. Run: npx projen integ:deploy`);

  const map: Record<string, string> = {};
  for (const o of stack.Outputs ?? []) {
    if (o.OutputKey && o.OutputValue) map[o.OutputKey] = o.OutputValue;
  }
  return map;
}

async function runQuery(sql: string, workgroup: string): Promise<string[][]> {
  const started = await athena.send(
    new StartQueryExecutionCommand({ QueryString: sql, WorkGroup: workgroup }),
  );
  const id = started.QueryExecutionId!;

  const deadline = Date.now() + QUERY_TIMEOUT_MS;
  for (;;) {
    const execution = await athena.send(new GetQueryExecutionCommand({ QueryExecutionId: id }));
    const state = execution.QueryExecution?.Status?.State;

    if (state === 'SUCCEEDED') break;
    if (state === 'FAILED' || state === 'CANCELLED') {
      throw new Error(
        `Athena query ${state}: ${execution.QueryExecution?.Status?.StateChangeReason ?? 'no reason given'}`,
      );
    }
    if (Date.now() > deadline) throw new Error(`Athena query timed out after ${QUERY_TIMEOUT_MS} ms`);
    await sleep(2000);
  }

  const results = await athena.send(new GetQueryResultsCommand({ QueryExecutionId: id }));
  // Row 0 is the header.
  return (results.ResultSet?.Rows ?? [])
    .slice(1)
    .map((row) => (row.Data ?? []).map((d) => d.VarCharValue ?? ''));
}

async function main(): Promise<void> {
  const out = await outputs();
  const marker = `m${Date.now().toString(36)}`;
  console.log(`Stack ${STACK_NAME}, marker ${marker}\n`);

  // 1. Invoke, with a marker unique to this run.
  console.log('1. invoke the function');
  const invoked = await lambda.send(
    new InvokeCommand({
      FunctionName: out.FunctionName,
      Payload: Buffer.from(JSON.stringify({ marker })),
    }),
  );
  check(invoked.StatusCode === 200 && !invoked.FunctionError, 'function invoked without error');

  // 2. Wait for a Parquet object and check the key layout.
  console.log('\n2. wait for a Parquet object in S3');
  const prefix = out.KeyPrefix || 'logs/';
  const deadline = Date.now() + OBJECT_TIMEOUT_MS;
  let keys: string[] = [];
  while (Date.now() < deadline) {
    const listed = await s3.send(
      new ListObjectsV2Command({ Bucket: out.LogsBucketName, Prefix: prefix }),
    );
    keys = (listed.Contents ?? []).map((o) => o.Key!).filter(Boolean);
    if (keys.length > 0) break;
    await sleep(3000);
  }

  check(keys.length > 0, 'at least one object was written', `prefix ${prefix}`);
  if (keys.length > 0) {
    const bad = keys.filter((k) => !KEY_PATTERN.test(k));
    check(bad.length === 0, 'every key is flat inside its hour partition', bad.slice(0, 3).join('\n        '));
    check(
      !keys.some((k) => k.includes('/extension/')),
      'no extension/ segment below the partitions',
    );
  }

  // 3. The object is a real Parquet file.
  console.log('\n3. inspect the object');
  if (keys.length > 0) {
    const object = await s3.send(new GetObjectCommand({ Bucket: out.LogsBucketName, Key: keys[0] }));
    const body = Buffer.from(await object.Body!.transformToByteArray());
    check(
      body.subarray(0, 4).toString('latin1') === 'PAR1' &&
        body.subarray(-4).toString('latin1') === 'PAR1',
      'the object is a Parquet file',
    );
  }

  // 4. Athena resolves the partitions and returns the rows.
  console.log('\n4. query through partition projection');
  const now = new Date();
  const y = now.getUTCFullYear();
  const m = String(now.getUTCMonth() + 1).padStart(2, '0');
  const d = String(now.getUTCDate()).padStart(2, '0');

  const sql = `SELECT level, source, message, request_id, correlation_id, context, stack_trace, caller
    FROM "${out.DatabaseName}"."${out.TableName}"
    WHERE year = '${y}' AND month = '${m}' AND day = '${d}'
      AND message LIKE '%${marker}%'
    LIMIT 50`;

  const rows = await runQuery(sql, out.WorkgroupName);
  check(rows.length > 0, 'Athena returned rows for this run', 'partition projection did not resolve');

  if (rows.length > 0) {
    const levels = new Set(rows.map((r) => r[0]));
    check(levels.has('INFO') && levels.has('ERROR'), 'structured levels survived the round trip');
    check(
      rows.some((r) => r[3] !== ''),
      'request_id was recovered from the runtime log prefix',
    );
    check(
      rows.some((r) => r[4].startsWith('integ-')),
      'correlation_id survived the round trip',
    );
    check(
      rows.some((r) => r[6] !== ''),
      'stack_trace landed in its own column',
    );
    check(
      rows.some((r) => r[1] === '' && r[2].includes('plain text line')),
      'a non-structured line was kept as plain text',
    );
  }

  console.log('');
  if (failures.length > 0) {
    console.error(`${failures.length} check(s) failed.`);
    process.exit(1);
  }
  console.log('All checks passed.');
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
