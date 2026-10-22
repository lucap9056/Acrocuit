import type { AuthSession } from "@/api";

type Identified = { id: unknown };

export type ListListener<T> = (items: T[]) => void;

export interface ListCache<K, T extends Identified> {
  clear: (key?: K) => void;
  fetch: (s: AuthSession, key: K) => Promise<T[]>;
  set: (key: K, items: T[]) => T[];
  add: (s: AuthSession, key: K, created: Promise<T>) => Promise<T[]>;
  replace: (s: AuthSession, key: K, updated: Promise<T>) => Promise<T[]>;
  remove: (
    s: AuthSession,
    key: K,
    id: T["id"],
    removed: Promise<unknown>,
  ) => Promise<T[]>;
  watch: (
    s: AuthSession,
    key: K,
    listener: ListListener<T>,
    signal: AbortSignal,
  ) => void;
}

interface Watcher<T> {
  session: AuthSession;
  listener: ListListener<T>;
}

export const createListCache = <K, T extends Identified>(
  load: (s: AuthSession, key: K) => Promise<T[]>,
): ListCache<K, T> => {
  const cache: Map<K, Promise<T[]>> = new Map();
  const watchers: Map<K, Set<Watcher<T>>> = new Map();

  const notify = (key: K, items: T[]) => {
    for (const { listener } of watchers.get(key) ?? []) {
      listener(items);
    }
  };

  const set = (key: K, items: T[]) => {
    cache.set(key, Promise.resolve(items));
    notify(key, items);
    return items;
  };

  const fetch = (s: AuthSession, key: K) => {
    const cached = cache.get(key);
    if (cached) {
      return cached;
    }

    const loading = load(s, key);
    cache.set(key, loading);
    loading.then(
      (items) => {
        if (cache.get(key) === loading) {
          notify(key, items);
        }
      },
      () => {
        if (cache.get(key) === loading) {
          cache.delete(key);
        }
      },
    );
    return loading;
  };

  const refetchWatched = (key: K) => {
    const [watcher] = watchers.get(key) ?? [];
    if (watcher) {
      fetch(watcher.session, key).catch((err) => {
        console.error(err);
      });
    }
  };

  return {
    clear: (key?: K) => {
      if (key === undefined) {
        cache.clear();
        for (const watched of watchers.keys()) {
          refetchWatched(watched);
        }
      } else {
        cache.delete(key);
        refetchWatched(key);
      }
    },
    fetch,
    set,
    add: (s, key, created) =>
      Promise.all([fetch(s, key), created]).then(([items, item]) => {
        if (items.some((i) => i.id === item.id)) {
          return items;
        }
        return set(key, [...items, item]);
      }),
    replace: (s, key, updated) =>
      Promise.all([fetch(s, key), updated]).then(([items, item]) =>
        set(
          key,
          items.map((i) => (i.id === item.id ? item : i)),
        ),
      ),
    remove: (s, key, id, removed) =>
      Promise.all([fetch(s, key), removed]).then(([items]) =>
        set(
          key,
          items.filter((i) => i.id !== id),
        ),
      ),
    watch: (s, key, listener, signal) => {
      if (signal.aborted) {
        return;
      }
      const watcher: Watcher<T> = { session: s, listener };
      const keyed = watchers.get(key) ?? new Set();
      keyed.add(watcher);
      watchers.set(key, keyed);
      signal.addEventListener(
        "abort",
        () => {
          keyed.delete(watcher);
          if (keyed.size === 0 && watchers.get(key) === keyed) {
            watchers.delete(key);
          }
        },
        { once: true },
      );
    },
  };
};
