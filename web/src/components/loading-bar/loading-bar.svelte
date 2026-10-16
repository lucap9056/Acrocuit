<script lang="ts">
  import { onMount } from "svelte";
  import loading, { type LoadEventListener } from "./index";

  import styles from "./loading-bar.module.css";

  let messages = $state<string[]>([]);
  let isVisible = $derived(messages.length > 0);

  onMount(() => {
    const listener: LoadEventListener = ({ value }) => {
      messages = value.messages.map((p) => p.payload);
    };
    loading.subscribe(listener);

    messages = loading.current;

    return () => {
      loading.unsubscribe(listener);
    };
  });
</script>

{#if isVisible}
  <div class={styles.load}>
    <div class={styles.loading_bar}></div>
  </div>
{/if}
