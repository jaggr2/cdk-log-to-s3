import * as fs from 'node:fs';
import * as path from 'node:path';
import { LogLevelName, StructuredLogger, createStructuredLogger } from '../src';

/**
 * The other half of the shared contract check. extension-go/contract_test.go
 * asserts the Go parser reads these lines; this asserts the logger still
 * produces them. A change to the wire format now breaks a test on both sides.
 *
 * See docs/CONTRACT.md.
 */
interface Fixture {
  cases: Array<{
    name: string;
    line: string;
    emit?: {
      method: 'debug' | 'info' | 'warn' | 'error';
      source: string;
      correlationId: string;
      message: string;
      context?: Record<string, unknown>;
    };
  }>;
}

const fixturePath = path.join(__dirname, '..', '..', '..', 'docs', 'contract-fixtures.json');
const fixtures: Fixture = JSON.parse(fs.readFileSync(fixturePath, 'utf8'));

describe('shared contract fixtures', () => {
  const emitCases = fixtures.cases.filter((c) => c.emit);

  test('the fixture file has cases the logger can reproduce', () => {
    expect(emitCases.length).toBeGreaterThan(0);
  });

  test.each(emitCases.map((c) => [c.name, c] as const))('%s', (_name, testCase) => {
    const emit = testCase.emit!;
    let emitted = '';

    const log: StructuredLogger = createStructuredLogger(emit.source, {
      correlationId: emit.correlationId,
      // Fixed so the fixture is not tied to this file's line numbers.
      captureCaller: false,
      level: 'DEBUG',
      write: (_level: LogLevelName, line: string) => {
        emitted = line;
      },
    });

    if (emit.method === 'error') {
      log.error(emit.message, undefined, emit.context);
    } else {
      log[emit.method](emit.message, emit.context);
    }

    // Compared as text, not as parsed objects: key order is part of what the
    // Go parser sees, and a byte comparison catches accidental reordering.
    expect(emitted).toBe(testCase.line);
  });

  test('every case is valid JSON carrying the discriminator', () => {
    for (const testCase of fixtures.cases) {
      const record = JSON.parse(testCase.line);
      expect(record.__log_level).toBeTruthy();
      expect(typeof record.message).toBe('string');
      // One record must be one line.
      expect(testCase.line).not.toContain('\n');
    }
  });
});
