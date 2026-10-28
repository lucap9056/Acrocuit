<script lang="ts">
  import { untrack } from "svelte";
  import type { Attributes } from "@/api";
  import AttributesForm from "@/components/attributes-form/attributes-form.svelte";
  import CollapsibleSection from "@/components/collapsible-section/collapsible-section.svelte";
  import {
    type AttributeField,
    describeAttributesIssue,
    fromAttributesForm,
    toAttributesForm,
  } from "@/lib/attributes";
  import { t } from "@/lib/i18n";
  import styles from "./attributes-section.module.css";

  interface Props {
    fields: readonly AttributeField[];
    attributes: Attributes | null;
    readonly?: boolean;
    onsave: (attributes: Attributes) => Promise<unknown>;
  }

  const { fields, attributes, readonly = false, onsave }: Props = $props();

  let form = $state(untrack(() => toAttributesForm(fields, attributes)));
  let saved = $state(untrack(() => JSON.stringify(form)));
  let error = $state<string>();
  let isPending = $state(false);

  const isDirty = $derived(JSON.stringify(form) !== saved);

  const clearError = () => {
    error = undefined;
  };

  const reset = () => {
    form = JSON.parse(saved);
    clearError();
  };

  const save = () => {
    const result = fromAttributesForm(fields, form);
    if (!result.ok) {
      error = describeAttributesIssue(result.issue);
      return;
    }

    isPending = true;
    onsave(result.attributes)
      .then(() => {
        form = toAttributesForm(fields, result.attributes);
        saved = JSON.stringify(form);
      })
      .catch((err) => {
        console.error(err);
        error = (err as Error).message;
      })
      .finally(() => {
        isPending = false;
      });
  };
</script>

<CollapsibleSection title={t("attributes.title")}>
  <div class={styles.body}>
    <AttributesForm
      {fields}
      bind:form
      disabled={isPending}
      {readonly}
      oninput={clearError}
    />
    {#if error}
      <p class={styles.error} role="alert">{error}</p>
    {/if}
    {#if isDirty}
      <div class={styles.actions}>
        <button class={styles.reset} disabled={isPending} onclick={reset}>
          {t("attributes.revert")}
        </button>
        <button class={styles.save} disabled={isPending} onclick={save}>
          {t("common.save")}
        </button>
      </div>
    {/if}
  </div>
</CollapsibleSection>
