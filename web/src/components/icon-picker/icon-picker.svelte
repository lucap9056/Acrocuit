<script lang="ts">
  import CaretRightIcon from "phosphor-svelte/lib/CaretRightIcon";
  import PlusIcon from "phosphor-svelte/lib/PlusIcon";
  import ProhibitIcon from "phosphor-svelte/lib/ProhibitIcon";
  import dialog from "@/components/dialog";
  import PackIcon from "@/components/pack-icon/pack-icon.svelte";
  import { tooltip } from "@/components/tooltip";
  import { t } from "@/lib/i18n";
  import {
    type PackIcon as Icon,
    IconPackError,
    iconLabel,
    iconRegistry,
    localize,
  } from "@/lib/icon-packs/index.svelte";
  import styles from "./icon-picker.module.css";

  interface Props {
    label: string;
    value: string;
    disabled?: boolean;
    onchange: (value: string) => void;
  }

  const { label, value, disabled = false, onchange }: Props = $props();

  let query = $state("");
  let pending = $state("");
  let activePackId = $state<string>();
  let sourceInput = $state("");
  let sourceError = $state<string>();

  const selected = $derived(iconRegistry.get(value || undefined));
  const activePack = $derived(
    iconRegistry.packs.find(({ source }) => source === activePackId) ??
      selected?.pack ??
      iconRegistry.packs[0],
  );

  const matches = (icon: Icon, needle: string) =>
    needle === "" ||
    [iconLabel(icon), icon.name, ...icon.tags].some((text) =>
      text.toLowerCase().includes(needle),
    );

  const groups = $derived.by(() => {
    const pack = activePack;
    if (!pack) {
      return [];
    }
    const needle = query.trim().toLowerCase();
    const byCategory = new Map<string, Icon[]>();
    for (const icon of pack.icons) {
      if (matches(icon, needle)) {
        const key = icon.category ?? "";
        byCategory.set(key, [...(byCategory.get(key) ?? []), icon]);
      }
    }
    return [...byCategory].map(([category, icons]) => ({
      key: `${pack.source}#${category}`,
      title: category
        ? localize(pack.categories[category] ?? {}, category)
        : pack.name,
      icons,
    }));
  });

  const describeSourceError = (err: unknown) =>
    err instanceof IconPackError
      ? t(`iconPicker.errors.${err.kind}`)
      : t("iconPicker.errors.unreachable");

  const openAddSource = () => {
    sourceInput = "";
    sourceError = undefined;
    dialog.push({
      text: t("iconPicker.addSource"),
      content: sourceForm,
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => true,
        },
        {
          text: t("iconPicker.add"),
          type: "primary",
          onclick: async () => {
            try {
              const pack = await iconRegistry.addSource(sourceInput);
              activePackId = pack.source;
              query = "";
              return true;
            } catch (err) {
              console.error(err);
              sourceError = describeSourceError(err);
              return false;
            }
          },
        },
      ],
    });
  };

  const focus = (input: HTMLInputElement) => {
    input.focus();
  };

  const open = () => {
    query = "";
    pending = value;
    activePackId = undefined;
    iconRegistry.load().catch((err) => {
      console.error(err);
    });
    dialog.push({
      text: label,
      content: picker,
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => true,
        },
        {
          text: t("common.confirm"),
          type: "primary",
          onclick: () => {
            if (pending !== value) {
              onchange(pending);
            }
            return true;
          },
        },
      ],
    });
  };
</script>

{#snippet sourceForm()}
  <div class={styles.source}>
    <input
      class={styles.search}
      type="url"
      inputmode="url"
      aria-label={t("iconPicker.sourceUrl")}
      placeholder="https://example.com/icon-pack/manifest.json"
      bind:value={sourceInput}
      oninput={() => (sourceError = undefined)}
      {@attach focus}
    >
    <p class={styles.hint}>{t("iconPicker.sourceHint")}</p>
    {#if sourceError}
      <p class={styles.error} role="alert">{sourceError}</p>
    {/if}
  </div>
{/snippet}

{#snippet picker()}
  <div class={styles.content}>
    <div class={styles.packs}>
      <div class={styles.tabs} role="group" aria-label={t("iconPicker.packs")}>
        {#each iconRegistry.packs as pack (pack.source)}
          <button
            type="button"
            class={[
              styles.tab,
              pack.source === activePack?.source && styles.active,
            ]}
            aria-pressed={pack.source === activePack?.source}
            onclick={() => (activePackId = pack.source)}
            {@attach tooltip(pack.name)}
          >
            {pack.name}
          </button>
        {/each}
      </div>
      <button
        type="button"
        class="icon_button {styles.add}"
        aria-label={t("iconPicker.addSource")}
        onclick={openAddSource}
        {@attach tooltip(t("iconPicker.addSource"))}
      >
        <PlusIcon />
      </button>
    </div>
    <input
      class={styles.search}
      type="search"
      aria-label={t("iconPicker.search")}
      placeholder={t("iconPicker.search")}
      bind:value={query}
      {@attach focus}
    >
    <div class={styles.grid}>
      <button
        type="button"
        class={[styles.option, styles.none, pending === "" && styles.selected]}
        aria-label={t("common.none")}
        aria-pressed={pending === ""}
        onclick={() => (pending = "")}
        {@attach tooltip(t("common.none"))}
      >
        <ProhibitIcon size="1.75rem" />
      </button>
    </div>
    {#each groups as group (group.key)}
      <h4 class={styles.heading}>{group.title}</h4>
      <div class={styles.grid} role="group" aria-label={group.title}>
        {#each group.icons as icon (icon.id)}
          <button
            type="button"
            class={[styles.option, icon.id === pending && styles.selected]}
            aria-label={iconLabel(icon)}
            aria-pressed={icon.id === pending}
            onclick={() => (pending = icon.id)}
            {@attach tooltip(iconLabel(icon))}
          >
            <PackIcon id={icon.id} size="1.75rem" />
          </button>
        {/each}
      </div>
    {:else}
      <p class={styles.empty}>
        {iconRegistry.packs.length === 0
          ? t("common.loading")
          : t("iconPicker.empty")}
      </p>
    {/each}
  </div>
{/snippet}

<button
  type="button"
  class={styles.trigger}
  aria-label={label}
  aria-haspopup="dialog"
  {disabled}
  onclick={open}
>
  <span class={styles.preview}>
    {#if selected}
      <PackIcon id={value} size="1.25rem" />
    {/if}
  </span>
  <span class={styles.value}>
    {selected ? iconLabel(selected.icon) : value || t("common.none")}
  </span>
  <CaretRightIcon class={styles.caret} aria-hidden="true" />
</button>
