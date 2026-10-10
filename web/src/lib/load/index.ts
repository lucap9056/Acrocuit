import { SvelteSet } from "svelte/reactivity";

export type Release = { release: () => boolean };

export interface LoadManager {
  readonly isLoading: boolean;
  begin: () => Release;
}

const pending = new SvelteSet<symbol>();

const loadManager: LoadManager = {
  get isLoading() {
    return pending.size > 0;
  },
  begin() {
    const token = Symbol();
    pending.add(token);
    return {
      release: () => pending.delete(token),
    };
  },
};

export default loadManager;
