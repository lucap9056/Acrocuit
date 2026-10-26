<script lang="ts">
  import { iconRegistry } from "@/lib/icon-packs/index.svelte";
  import { onMount } from "@/lib/lifecycle";

  interface Props {
    id: string | undefined;
    size?: string;
    class?: string;
  }

  const { id, size = "1em", class: className }: Props = $props();

  const entry = $derived(iconRegistry.get(id));

  onMount(() => {
    iconRegistry.load().catch((err) => {
      console.error(err);
    });
  });
</script>

{#if entry}
  {@const { pack, icon } = entry}
  <svg
    class={className}
    xmlns="http://www.w3.org/2000/svg"
    viewBox={pack.style.viewBox}
    width={size}
    height={size}
    aria-hidden="true"
  >
    {#each icon.fills as d, index (index)}
      <path {d} fill="currentColor" />
    {/each}
    {#each icon.strokes as d, index (index)}
      <path
        {d}
        fill="none"
        stroke="currentColor"
        stroke-width={pack.style.strokeWidth}
        stroke-linecap={pack.style.lineCap}
        stroke-linejoin={pack.style.lineJoin}
      />
    {/each}
  </svg>
{/if}
