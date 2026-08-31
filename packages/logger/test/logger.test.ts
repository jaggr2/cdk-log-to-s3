import {
  LogLevelName,
  StructuredRecord,
  createStructuredLogger,
  getCorrelationId,
  runWithCorrelationId,
  setCorrelationId,
} from '../src';

interface Captured {
  level: LogLevelName;
  record: StructuredRecord;
  line: string;
}

/** Collects emitted lines through the write seam instead of spying on console. */
function collector() {
  const lines: Captured[] = [];
  const write = (level: LogLevelName, line: string) => {
    lines.push({ level, line, record: JSON.parse(line) });
  };
  return { lines, write };
}

describe('createStructuredLogger', () => {
  test('emits the contract the Go parser expects', () => {
    const { lines, write } = collector();
    const log = createStructuredLogger('SyncProcessor', {
      write,
      correlationId: 'cid-1',
      captureCaller: false,
    });

    log.info('Events synced', { count: 5 });

    expect(lines).toHaveLength(1);
    expect(lines[0].record).toEqual({
      __log_level: 'INFO',
      __source: 'SyncProcessor',
      __correlation_id: 'cid-1',
      message: 'Events synced',
      context: { count: 5 },
    });
  });

  test('emits exactly one line with no raw newlines or tabs', () => {
    const { lines, write } = collector();
    const log = createStructuredLogger('Api', { write, captureCaller: false });

    log.error('multi\nline\tmessage', new Error('boom\nwith newline'));

    expect(lines).toHaveLength(1);
    // JSON.stringify escapes both, which is what keeps one record on one line.
    expect(lines[0].line).not.toContain('\n');
    expect(lines[0].line).not.toContain('\t');
  });

  test('__log_level is always present, since it is the parser discriminator', () => {
    const { lines, write } = collector();
    const log = createStructuredLogger('Api', { write, level: 'DEBUG', captureCaller: false });

    log.debug('d');
    log.info('i');
    log.warn('w');
    log.error('e');

    expect(lines.map((l) => l.record.__log_level)).toEqual(['DEBUG', 'INFO', 'WARN', 'ERROR']);
    for (const line of lines) {
      expect(line.record.__log_level).toBeTruthy();
    }
  });

  test('omits context and stackTrace rather than emitting null', () => {
    const { lines, write } = collector();
    const log = createStructuredLogger('Api', { write, captureCaller: false });

    log.info('no context');
    log.info('empty context', {});

    for (const line of lines) {
      expect('context' in line.record).toBe(false);
      expect('stackTrace' in line.record).toBe(false);
    }
  });

  describe('level filtering', () => {
    test('drops records below the threshold before they are written', () => {
      const { lines, write } = collector();
      const log = createStructuredLogger('Api', { write, level: 'WARN', captureCaller: false });

      log.debug('dropped');
      log.info('dropped');
      log.warn('kept');
      log.error('kept');

      expect(lines.map((l) => l.record.message)).toEqual(['kept', 'kept']);
    });

    test('reads the default level from the environment', () => {
      const previous = process.env.LOG_TO_S3_LEVEL;
      process.env.LOG_TO_S3_LEVEL = 'ERROR';
      try {
        const { lines, write } = collector();
        const log = createStructuredLogger('Api', { write, captureCaller: false });
        log.warn('dropped');
        log.error('kept');
        expect(lines).toHaveLength(1);
      } finally {
        if (previous === undefined) delete process.env.LOG_TO_S3_LEVEL;
        else process.env.LOG_TO_S3_LEVEL = previous;
      }
    });

    test('falls back to LOG_LEVEL and then to INFO', () => {
      const previousNew = process.env.LOG_TO_S3_LEVEL;
      const previousOld = process.env.LOG_LEVEL;
      delete process.env.LOG_TO_S3_LEVEL;
      process.env.LOG_LEVEL = 'DEBUG';
      try {
        const { lines, write } = collector();
        createStructuredLogger('Api', { write, captureCaller: false }).debug('kept');
        expect(lines).toHaveLength(1);
      } finally {
        if (previousNew !== undefined) process.env.LOG_TO_S3_LEVEL = previousNew;
        if (previousOld === undefined) delete process.env.LOG_LEVEL;
        else process.env.LOG_LEVEL = previousOld;
      }
    });

    test('an unparseable level falls back to INFO rather than throwing', () => {
      const previous = process.env.LOG_TO_S3_LEVEL;
      process.env.LOG_TO_S3_LEVEL = 'LOUD';
      try {
        const { lines, write } = collector();
        const log = createStructuredLogger('Api', { write, captureCaller: false });
        log.debug('dropped');
        log.info('kept');
        expect(lines.map((l) => l.record.message)).toEqual(['kept']);
      } finally {
        if (previous === undefined) delete process.env.LOG_TO_S3_LEVEL;
        else process.env.LOG_TO_S3_LEVEL = previous;
      }
    });
  });

  describe('error()', () => {
    test('splits an Error into context and a top-level stack trace', () => {
      const { lines, write } = collector();
      const log = createStructuredLogger('Api', { write, captureCaller: false });

      log.error('failed', new TypeError('bad input'), { attempt: 2 });

      const record = lines[0].record;
      expect(record.context).toEqual({
        attempt: 2,
        error: { name: 'TypeError', message: 'bad input' },
      });
      // Its own column in the table, so it stays queryable.
      expect(record.stackTrace).toContain('TypeError: bad input');
    });

    test('handles a non-Error throw', () => {
      const { lines, write } = collector();
      const log = createStructuredLogger('Api', { write, captureCaller: false });

      log.error('failed', 'just a string');

      expect(lines[0].record.context).toEqual({ error: { value: 'just a string' } });
      expect(lines[0].record.stackTrace).toBeUndefined();
    });

    test('works with no error at all', () => {
      const { lines, write } = collector();
      createStructuredLogger('Api', { write, captureCaller: false }).error('failed', undefined, {
        code: 500,
      });

      expect(lines[0].record.context).toEqual({ code: 500 });
      expect(lines[0].record.stackTrace).toBeUndefined();
    });

    test('goes to the ERROR stream', () => {
      const { lines, write } = collector();
      createStructuredLogger('Api', { write, captureCaller: false }).error('failed');
      expect(lines[0].level).toBe('ERROR');
    });
  });

  describe('caller capture', () => {
    test('records the call site by default', () => {
      const { lines, write } = collector();
      createStructuredLogger('Api', { write }).info('hello');

      expect(lines[0].record.__caller).toContain('logger.test.ts');
    });

    test('can be switched off', () => {
      const { lines, write } = collector();
      createStructuredLogger('Api', { write, captureCaller: false }).info('hello');

      expect(lines[0].record.__caller).toBeUndefined();
    });

    test('trims paths at the root marker', () => {
      const { lines, write } = collector();
      createStructuredLogger('Api', { write, callerRootMarker: '/test/' }).info('hello');

      const caller = lines[0].record.__caller!;
      expect(caller.startsWith('test/')).toBe(true);
    });
  });

  describe('child and withCorrelationId', () => {
    test('child changes the source and keeps everything else', () => {
      const { lines, write } = collector();
      const log = createStructuredLogger('Api', {
        write,
        correlationId: 'cid-1',
        captureCaller: false,
        level: 'WARN',
      });

      log.child('Worker').warn('from child');
      log.child('Worker').info('dropped by the inherited threshold');

      expect(lines).toHaveLength(1);
      expect(lines[0].record.__source).toBe('Worker');
      expect(lines[0].record.__correlation_id).toBe('cid-1');
    });

    test('withCorrelationId pins the id', () => {
      const { lines, write } = collector();
      const log = createStructuredLogger('Api', { write, captureCaller: false });

      log.withCorrelationId('cid-9').info('hello');

      expect(lines[0].record.__correlation_id).toBe('cid-9');
    });
  });

  describe('ambient correlation id', () => {
    test('is picked up without threading it through call signatures', () => {
      const { lines, write } = collector();
      const log = createStructuredLogger('Api', { write, captureCaller: false });

      runWithCorrelationId('ambient-1', () => {
        log.info('inside');
      });
      log.info('outside');

      expect(lines[0].record.__correlation_id).toBe('ambient-1');
      expect(lines[1].record.__correlation_id).toBe('');
    });

    test('survives an await', async () => {
      const { lines, write } = collector();
      const log = createStructuredLogger('Api', { write, captureCaller: false });

      await runWithCorrelationId('ambient-2', async () => {
        await new Promise((resolve) => setTimeout(resolve, 1));
        log.info('after await');
      });

      expect(lines[0].record.__correlation_id).toBe('ambient-2');
    });

    test('an explicit id beats the ambient one', () => {
      const { lines, write } = collector();
      const log = createStructuredLogger('Api', {
        write,
        captureCaller: false,
        correlationId: 'explicit',
      });

      runWithCorrelationId('ambient', () => log.info('hello'));

      expect(lines[0].record.__correlation_id).toBe('explicit');
    });

    test('setCorrelationId replaces the id inside a context', () => {
      runWithCorrelationId('first', () => {
        setCorrelationId('second');
        expect(getCorrelationId()).toBe('second');
      });
    });

    test('setCorrelationId outside a context does not leak an ambient store', () => {
      setCorrelationId('orphan');
      expect(getCorrelationId()).toBeUndefined();
    });
  });

  test('the default sink writes to console.error only for ERROR', () => {
    const log = jest.spyOn(console, 'log').mockImplementation(() => {});
    const err = jest.spyOn(console, 'error').mockImplementation(() => {});
    try {
      const logger = createStructuredLogger('Api', { captureCaller: false, level: 'DEBUG' });
      logger.info('to stdout');
      logger.error('to stderr');

      expect(log).toHaveBeenCalledTimes(1);
      expect(err).toHaveBeenCalledTimes(1);
    } finally {
      log.mockRestore();
      err.mockRestore();
    }
  });
});
