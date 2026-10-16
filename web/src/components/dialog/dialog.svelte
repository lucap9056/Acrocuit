<script lang="ts">
  import { onMount } from "svelte";
  import type { Message } from "@/lib/message-queue";
  import styles from "./dialog.module.css";
  import type { DialogEvent, DialogMessage } from "./index";
  import dialog from "./index";

  let messages = $state<Message<DialogMessage>[]>([]);
  let pendingToken = $state<symbol>();
  const lastMessage = $derived(
    messages.length > 0 ? messages[messages.length - 1] : undefined,
  );

  onMount(() => {
    const updateMessages = ({ value }: DialogEvent) => {
      messages = value.messages;
    };
    dialog.subscribe(updateMessages);
    messages = dialog.items;

    return () => {
      dialog.unsubscribe(updateMessages);
    };
  });

  const onclick = async (msg: Message<DialogMessage>, index: number) => {
    const button = msg.payload.buttons[index];
    pendingToken = msg.token;
    try {
      const shouldDismiss = await button.onclick();
      if (shouldDismiss !== false) {
        dialog.dismiss(msg.token);
      }
    } finally {
      pendingToken = undefined;
    }
  };
</script>

{#if lastMessage}
  <div class={styles.dialog}>
    <div class={styles.dialog_message}>
      {#key lastMessage.token}
        {#if lastMessage.payload.text}
          <div class={styles.dialog_text}>{lastMessage.payload.text}</div>
        {/if}
        {#if lastMessage.payload.content}
          <div class={styles.dialog_content}>
            {@render lastMessage.payload.content({
              dismiss: () => dialog.dismiss(lastMessage.token),
            })}
          </div>
        {/if}
      {/key}
      {#if lastMessage.payload.buttons.length > 0}
        <div class={styles.dialog_buttons}>
          {#each lastMessage.payload.buttons as button, i}
            <button
              data-type={button.type}
              disabled={pendingToken === lastMessage.token}
              onclick={() => onclick(lastMessage, i)}
            >
              {button.text}
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>
{/if}
