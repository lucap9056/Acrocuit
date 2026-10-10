import { getAbortSignal, onMount as svelteOnMount } from "svelte";

type MountCallback = (signal: AbortSignal) => undefined | (() => void);

export function onMount(fn: MountCallback) {
  svelteOnMount(() => fn(getAbortSignal()));
}
