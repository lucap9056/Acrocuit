<script lang="ts">
  import CaretRightIcon from "phosphor-svelte/lib/CaretRightIcon";
  import type { Snippet } from "svelte";
  import { prefersReducedMotion } from "svelte/motion";
  import { slide } from "svelte/transition";
  import styles from "./collapsible-section.module.css";

  interface Props {
    title: string;
    open?: boolean;
    children: Snippet;
  }

  let { title, open = $bindable(false), children }: Props = $props();

  const bodyId = $props.id();

  const duration = $derived(prefersReducedMotion.current ? 0 : 150);
</script>

<section class={[styles.section, open && styles.open]}>
  <h3 class={styles.heading}>
    <button
      type="button"
      class={styles.summary}
      aria-expanded={open}
      aria-controls={bodyId}
      onclick={() => (open = !open)}
    >
      <CaretRightIcon class={styles.caret} aria-hidden="true" />
      {title}
    </button>
  </h3>
  {#if open}
    <div id={bodyId} class={styles.body} transition:slide={{ duration }}>
      {@render children()}
    </div>
  {/if}
</section>
