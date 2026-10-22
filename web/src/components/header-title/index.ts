import type { Snippet } from "svelte";

export type HeaderTitleListener = (title?: Snippet) => void;

export interface HeaderTitleManager {
  readonly current: Snippet | undefined;
  push: (title: Snippet, signal: AbortSignal) => void;
  subscribe: (listener: HeaderTitleListener) => void;
  unsubscribe: (listener: HeaderTitleListener) => void;
}

interface HeaderTitleEntry {
  title: Snippet;
}

export const createHeaderTitleManager = (): HeaderTitleManager => {
  const listeners: Set<HeaderTitleListener> = new Set();
  const entries: HeaderTitleEntry[] = [];

  const current = () => entries.at(-1)?.title;

  const notify = () => {
    const title = current();
    for (const listener of listeners) {
      listener(title);
    }
  };

  return {
    get current() {
      return current();
    },
    push(title, signal) {
      if (signal.aborted) {
        return;
      }

      const entry: HeaderTitleEntry = { title };
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

export default createHeaderTitleManager();
