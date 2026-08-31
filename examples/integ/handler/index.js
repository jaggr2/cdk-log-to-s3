/**
 * Emits one of each record shape the extension has to handle.
 *
 * Written as plain JS rather than importing @jaggr2/log-to-s3-logger so the
 * integ app needs no bundling step. The exact bytes the logger produces are
 * already pinned by docs/contract-fixtures.json, which both packages test
 * against; what this handler exercises is the delivery path.
 */
function emit(level, source, correlationId, message, context, stackTrace) {
  const record = {
    __log_level: level,
    __source: source,
    __correlation_id: correlationId,
    __caller: 'examples/integ/handler/index.js',
    message,
  };
  if (context) record.context = context;
  if (stackTrace) record.stackTrace = stackTrace;

  const line = JSON.stringify(record);
  if (level === 'ERROR') console.error(line);
  else console.log(line);
}

exports.handler = async (event) => {
  const marker = (event && event.marker) || 'no-marker';
  const correlationId = `integ-${marker}`;

  emit('INFO', 'IntegHandler', correlationId, `structured info ${marker}`, { marker, count: 5 });
  emit('DEBUG', 'IntegHandler', correlationId, `structured debug ${marker}`, { marker });
  emit(
    'ERROR',
    'IntegHandler',
    correlationId,
    `structured error ${marker}`,
    { marker, error: { name: 'TypeError', message: 'bad input' } },
    'TypeError: bad input\n    at handler (index.js:1:1)',
  );

  // A line that is not one of our records: it must still land in the table as
  // a plain-text row rather than being dropped.
  console.log(`plain text line ${marker}`);

  return { ok: true, marker, correlationId };
};
