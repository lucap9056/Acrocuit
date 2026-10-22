<script lang="ts">
  import FloppyDiskIcon from "phosphor-svelte/lib/FloppyDiskIcon";
  import type { ClassValue } from "svelte/elements";
  import type { Attachment } from "svelte/attachments";
  import { tooltip } from "@/components/tooltip";
  import { t } from "@/lib/i18n";
  import styles from "./textbox.module.css";

  interface Props {
    label: string;
    value: string;
    placeholder?: string;
    disabled?: boolean;
    class?: ClassValue;
    onsave: (value: string) => unknown;
  }

  const {
    label,
    value,
    placeholder = "",
    disabled = false,
    class: className,
    onsave,
  }: Props = $props();

  let isEditing = $state(false);
  let isSaving = $state(false);
  let draft = $state("");

  const isDirty = $derived(isEditing && draft !== value);

  const focusOnMount: Attachment<HTMLInputElement> = (input) => {
    input.focus();
    input.select();
  };

  const edit = () => {
    if (disabled) {
      return;
    }
    draft = value;
    isEditing = true;
  };

  const cancel = () => {
    isEditing = false;
  };

  const save = async () => {
    if (!isDirty || isSaving) {
      return;
    }
    isSaving = true;
    try {
      await onsave(draft);
      isEditing = false;
    } catch (err) {
      console.error(err);
    } finally {
      isSaving = false;
    }
  };

  const onkeydown = (event: KeyboardEvent) => {
    if (event.isComposing) {
      return;
    }
    if (event.key === "Enter") {
      event.preventDefault();
      save();
    } else if (event.key === "Escape") {
      event.preventDefault();
      cancel();
    }
  };

  const onfocusout = (event: FocusEvent) => {
    const container = event.currentTarget as HTMLElement;
    if (!isDirty && !container.contains(event.relatedTarget as Node | null)) {
      cancel();
    }
  };
</script>

<div class={[styles.textbox, className]} {onfocusout}>
  {#if isEditing}
    <input
      class={styles.input}
      type="text"
      aria-label={label}
      {placeholder}
      disabled={isSaving}
      bind:value={draft}
      {onkeydown}
      {@attach focusOnMount}
    >
    {#if isDirty}
      <button
        class="icon_button {styles.save}"
        aria-label={t("common.save")}
        disabled={isSaving}
        onclick={save}
        {@attach tooltip(t("common.save"))}
      >
        <FloppyDiskIcon />
      </button>
    {/if}
  {:else}
    <button
      type="button"
      class={[styles.display, value === "" && styles.placeholder]}
      aria-label={label}
      {disabled}
      onclick={edit}
    >
      {value === "" ? placeholder : value}
    </button>
  {/if}
</div>
