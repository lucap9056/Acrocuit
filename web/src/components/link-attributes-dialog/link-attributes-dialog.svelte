<script module lang="ts">
  import type { Attributes } from "@/api";
  import AttributesForm from "@/components/attributes-form/attributes-form.svelte";
  import dialog from "@/components/dialog";
  import {
    describeAttributesIssue,
    type AttributesForm as Form,
    fromAttributesForm,
    LINK_ATTRIBUTE_FIELDS,
    toAttributesForm,
  } from "@/lib/attributes";
  import { t } from "@/lib/i18n";
  import styles from "./link-attributes-dialog.module.css";

  let form = $state<Form>(toAttributesForm(LINK_ATTRIBUTE_FIELDS, null));
  let error = $state<string>();
  let isReadonly = $state(false);

  const clearError = () => {
    error = undefined;
  };

  export const editLinkAttributes = (
    attributes: Attributes | null,
    onsave: (attributes: Attributes) => Promise<unknown>,
  ) => {
    form = toAttributesForm(LINK_ATTRIBUTE_FIELDS, attributes);
    error = undefined;
    isReadonly = false;

    dialog.push({
      text: t("attributes.edit"),
      content: editor,
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => true,
        },
        {
          text: t("common.save"),
          type: "primary",
          onclick: async () => {
            const result = fromAttributesForm(LINK_ATTRIBUTE_FIELDS, form);
            if (!result.ok) {
              error = describeAttributesIssue(result.issue);
              return false;
            }
            try {
              await onsave(result.attributes);
              return true;
            } catch (err) {
              console.error(err);
              error = (err as Error).message;
              return false;
            }
          },
        },
      ],
    });
  };

  export const viewLinkAttributes = (attributes: Attributes | null) => {
    form = toAttributesForm(LINK_ATTRIBUTE_FIELDS, attributes);
    error = undefined;
    isReadonly = true;

    dialog.push({
      text: t("attributes.view"),
      content: editor,
      buttons: [
        {
          text: t("common.close"),
          type: "",
          onclick: () => true,
        },
      ],
    });
  };
</script>

{#snippet editor()}
  <div class={styles.scroll}>
    <AttributesForm
      fields={LINK_ATTRIBUTE_FIELDS}
      bind:form
      readonly={isReadonly}
      oninput={clearError}
    />
  </div>
  {#if error}
    <p class={styles.error} role="alert">{error}</p>
  {/if}
{/snippet}
