export type BackHandler = () => void;

export type BackListener = (onback?: BackHandler) => void;

export interface BackManager {
  readonly current: BackHandler | undefined;
  push: (onback: BackHandler, signal: AbortSignal) => void;
  subscribe: (listener: BackListener) => void;
  unsubscribe: (listener: BackListener) => void;
}

interface BackEntry {
  onback: BackHandler;
}

export const createBackManager = (): BackManager => {
  const listeners: Set<BackListener> = new Set();
  const entries: BackEntry[] = [];

  const current = () => entries.at(-1)?.onback;

  const notify = () => {
    const onback = current();
    for (const listener of listeners) {
      listener(onback);
    }
  };

  return {
    get current() {
      return current();
    },
    push(onback, signal) {
      if (signal.aborted) {
        return;
      }

      const entry: BackEntry = { onback };
      entries.push(entry);
      signal.addEventListener(
        "abort",
        () => {
          const index = entries.indexOf(entry);
          if (index === -1) {
            return;
          }
          const wasTop = index === entries.length - 1;
          entries.splice(index, 1);
          if (wasTop) {
            notify();
          }
        },
        { once: true },
      );
      notify();
    },
    subscribe(listener) {
      listeners.add(listener);
    },
    unsubscribe(listener) {
      listeners.delete(listener);
    },
  };
};

export default createBackManager();
