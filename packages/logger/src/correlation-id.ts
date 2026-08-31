import { AsyncLocalStorage } from 'node:async_hooks';

interface Store {
  correlationId: string;
}

const storage = new AsyncLocalStorage<Store>();

/**
 * Runs `fn` with an ambient correlation id that every logger created inside it
 * picks up automatically, including across awaits.
 *
 * Wrap the body of a Lambda handler in this and the id flows into every log
 * record without being threaded through call signatures.
 */
export function runWithCorrelationId<T>(correlationId: string, fn: () => T): T {
  return storage.run({ correlationId }, fn);
}

export function getCorrelationId(): string | undefined {
  return storage.getStore()?.correlationId;
}

/**
 * Replaces the correlation id of the current context.
 *
 * Only has an effect inside runWithCorrelationId; outside one there is no
 * store to mutate and the call is a no-op, which is deliberate - creating an
 * ambient store here would leak it into unrelated async work.
 */
export function setCorrelationId(correlationId: string): void {
  const store = storage.getStore();
  if (store) {
    store.correlationId = correlationId;
  }
}
