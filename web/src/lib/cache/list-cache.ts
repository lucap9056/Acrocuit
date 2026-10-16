import type { AuthSession } from "@/api";

type Identified = { id: unknown };

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
}

export const createListCache = <K, T extends Identified>(
  load: (s: AuthSession, key: K) => Promise<T[]>,
): ListCache<K, T> => {
  const cache: Map<K, Promise<T[]>> = new Map();

  const set = (key: K, items: T[]) => {
    cache.set(key, Promise.resolve(items));
    return items;
  };

  const fetch = (s: AuthSession, key: K) => {
    const cached = cache.get(key);
    if (cached) {
      return cached;
    }

    const loading = load(s, key);
    cache.set(key, loading);
    loading.catch(() => {
      if (cache.get(key) === loading) {
        cache.delete(key);
      }
    });
    return loading;
  };

  return {
    clear: (key?: K) => {
      if (key === undefined) {
        cache.clear();
      } else {
        cache.delete(key);
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
  };
};
