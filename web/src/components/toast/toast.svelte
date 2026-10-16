<script lang="ts">
  import { onMount } from "svelte";
  import toast, { type ToastEvent } from "./index";

  import styles from "./toast.module.css";

  let messages = $state<string[]>([]);

  onMount(() => {
    const updateMessages = ({ value }: ToastEvent) => {
      messages = value.messages.map((p) => p.payload).reverse();
    };

    toast.subscribe(updateMessages);

    return () => {
      toast.unsubscribe(updateMessages);
    };
  });
</script>

{#if messages.length > 0}
  <div class={styles.toast}>
    {#each messages as message}
      <div class={styles.toast_item}>{message}</div>
    {/each}
  </div>
{/if}
