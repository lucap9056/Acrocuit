<script lang="ts">
  import PlusIcon from "phosphor-svelte/lib/PlusIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import Select, { type SelectOption } from "@/components/select/select.svelte";
  import { tooltip } from "@/components/tooltip";
  import type {
    AttributeField,
    AttributesForm,
    CustomAttribute,
  } from "@/lib/attributes";
  import { t } from "@/lib/i18n";
  import styles from "./attributes-form.module.css";

  interface Props {
    fields: readonly AttributeField[];
    form: AttributesForm;
    disabled?: boolean;
    oninput?: () => void;
  }

  let {
    fields,
    form = $bindable(),
    disabled = false,
    oninput,
  }: Props = $props();

  const addCustom = () => {
    form.custom.push({ key: "", value: "" });
    oninput?.();
  };

  const removeCustom = (row: CustomAttribute) => {
    form.custom = form.custom.filter((r) => r !== row);
    oninput?.();
  };

  const inputMode = (field: AttributeField) => {
    switch (field.type) {
      case "integer":
        return "numeric";
      case "number":
        return "decimal";
      default:
        return "text";
    }
  };

  const enumOptions = (
    field: Extract<AttributeField, { type: "enum" }>,
  ): SelectOption<string>[] => [
    { value: "", label: t("common.none") },
    ...field.options.map(({ value, label }) => ({ value, label: label() })),
  ];

  const choose = (field: AttributeField, value: string) => {
    form.fixed[field.key] = value;
    oninput?.();
  };
</script>

<div class={styles.form}>
  <div class={styles.fields}>
    {#each fields as field (field.key)}
      {@const label = t(`attributes.fields.${field.key}`)}
      {#if field.type === "enum"}
        <div class={styles.field}>
          <span class={styles.label}>{label}</span>
          <span class={styles.control}>
            <span class={styles.select}>
              <Select
                {label}
                options={enumOptions(field)}
                value={form.fixed[field.key] ?? ""}
                {disabled}
                onchange={(value) => choose(field, value)}
              />
            </span>
          </span>
        </div>
      {:else}
        <label class={styles.field}>
          <span class={styles.label}>{label}</span>
          <span class={styles.control}>
            <input
              class={styles.input}
              type="text"
              inputmode={inputMode(field)}
              placeholder={field.estimate?.(form.fixed)}
              {disabled}
              bind:value={form.fixed[field.key]}
              {oninput}
            >
            {#if field.unit}
              <span class={styles.unit}>{field.unit}</span>
            {/if}
          </span>
        </label>
      {/if}
    {/each}
  </div>
  <div class={styles.custom}>
    <h4 class={styles.heading}>{t("attributes.custom")}</h4>
    {#each form.custom as row (row)}
      <div class={styles.custom_row}>
        <input
          class={styles.input}
          type="text"
          aria-label={t("attributes.key")}
          placeholder={t("attributes.key")}
          {disabled}
          bind:value={row.key}
          {oninput}
        >
        <input
          class={styles.input}
          type="text"
          aria-label={t("attributes.value")}
          placeholder={t("attributes.value")}
          {disabled}
          bind:value={row.value}
          {oninput}
        >
        <button
          class="icon_button {styles.remove}"
          aria-label={t("common.remove")}
          {disabled}
          onclick={() => removeCustom(row)}
          {@attach tooltip(t("common.remove"))}
        >
          <TrashIcon />
        </button>
      </div>
    {/each}
    <button class={styles.add} {disabled} onclick={addCustom}>
      <PlusIcon />
      {t("attributes.addCustom")}
    </button>
  </div>
</div>
