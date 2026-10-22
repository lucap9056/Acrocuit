<script lang="ts">
  import type { Snippet } from "svelte";
  import { onMount } from "@/lib/lifecycle";
  import styles from "./header-title.module.css";
  import headerTitle, { type HeaderTitleListener } from "./index";

  let title = $state<Snippet>();

  onMount(() => {
    const listener: HeaderTitleListener = (next) => {
      title = next;
    };
    headerTitle.subscribe(listener);

    title = headerTitle.current;

    return () => {
      headerTitle.unsubscribe(listener);
    };
  });
</script>

{#if title}
  <div class={styles.title}>{@render title()}</div>
{/if}
