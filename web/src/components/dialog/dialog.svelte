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
    {#each messages as message (message.token)}
      {@const isTop = message.token === lastMessage.token}
      <div class={styles.dialog_message} hidden={!isTop} inert={!isTop}>
        {#if message.payload.text}
          <div class={styles.dialog_text}>{message.payload.text}</div>
        {/if}
        {#if message.payload.content}
          <div class={styles.dialog_content}>
            {@render message.payload.content({
              dismiss: () => dialog.dismiss(message.token),
            })}
          </div>
        {/if}
        {#if message.payload.buttons.length > 0}
          <div class={styles.dialog_buttons}>
            {#each message.payload.buttons as button, i}
              <button
                data-type={button.type}
                data-align={button.align}
                disabled={pendingToken === message.token}
                onclick={() => onclick(message, i)}
              >
                {button.text}
              </button>
            {/each}
          </div>
        {/if}
      </div>
    {/each}
  </div>
{/if}
