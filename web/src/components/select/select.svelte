<script lang="ts" module>
  export interface SelectOption<V> {
    value: V;
    label: string;
  }
</script>

<script lang="ts" generics="T">
  import CaretDownIcon from "phosphor-svelte/lib/CaretDownIcon";
  import Spinner from "@/components/spinner/spinner.svelte";
  import { t } from "@/lib/i18n";
  import styles from "./select.module.css";

  interface Props {
    label: string;
    options?: SelectOption<T>[];
    value: T;
    disabled?: boolean;
    onchange?: (value: T) => void;
  }

  let {
    label,
    options,
    value = $bindable(),
    disabled = false,
    onchange,
  }: Props = $props();

  const id = $props.id();

  let root = $state<HTMLDivElement>();
  let list = $state<HTMLUListElement>();
  let isOpen = $state(false);
  let activeIndex = $state(-1);

  const isLoading = $derived(options === undefined);
  const isDisabled = $derived(disabled || isLoading);
  const selectedIndex = $derived(
    options?.findIndex((option) => option.value === value) ?? -1,
  );
  const selectedLabel = $derived(options?.[selectedIndex]?.label ?? "");

  const optionId = (index: number) => `${id}-option-${index}`;

  const open = () => {
    if (isDisabled) {
      return;
    }
    isOpen = true;
    activeIndex = Math.max(selectedIndex, 0);
  };

  const close = () => {
    isOpen = false;
    activeIndex = -1;
  };

  const choose = (index: number) => {
    const option = options?.[index];
    close();
    if (option === undefined || option.value === value) {
      return;
    }
    value = option.value;
    onchange?.(option.value);
  };

  const move = (offset: number) => {
    const count = options?.length ?? 0;
    if (count === 0) {
      return;
    }
    activeIndex = (activeIndex + offset + count) % count;
  };

  const onkeydown = (event: KeyboardEvent) => {
    if (!isOpen) {
      if (["ArrowDown", "ArrowUp", "Enter", " "].includes(event.key)) {
        event.preventDefault();
        open();
      }
      return;
    }

    switch (event.key) {
      case "ArrowDown":
        event.preventDefault();
        move(1);
        break;
      case "ArrowUp":
        event.preventDefault();
        move(-1);
        break;
      case "Home":
        event.preventDefault();
        activeIndex = 0;
        break;
      case "End":
        event.preventDefault();
        activeIndex = (options?.length ?? 0) - 1;
        break;
      case "Enter":
      case " ":
        event.preventDefault();
        choose(activeIndex);
        break;
      case "Escape":
        event.preventDefault();
        close();
        break;
      case "Tab":
        close();
        break;
    }
  };

  const onlistclick = (event: MouseEvent) => {
    const option = (event.target as Element).closest<HTMLElement>("[data-index]");
    if (option) {
      choose(Number(option.dataset.index));
    }
  };

  const onwindowpointerdown = (event: PointerEvent) => {
    if (isOpen && !root?.contains(event.target as Node)) {
      close();
    }
  };

  $effect(() => {
    if (isOpen && activeIndex >= 0) {
      list
        ?.querySelector(`#${CSS.escape(optionId(activeIndex))}`)
        ?.scrollIntoView({ block: "nearest" });
    }
  });

  $effect(() => {
    if (isDisabled) {
      close();
    }
  });
</script>

<svelte:window onpointerdown={onwindowpointerdown} />

<div class={styles.select} bind:this={root}>
  <button
    type="button"
    class={styles.trigger}
    role="combobox"
    aria-label={label}
    aria-haspopup="listbox"
    aria-expanded={isOpen}
    aria-controls={`${id}-listbox`}
    aria-activedescendant={isOpen && activeIndex >= 0
      ? optionId(activeIndex)
      : undefined}
    aria-busy={isLoading}
    disabled={isDisabled}
    onclick={() => (isOpen ? close() : open())}
    {onkeydown}
  >
    {#if isLoading}
      <span class={styles.loading} role="status" aria-label={t("common.loading")}>
        <Spinner />
      </span>
    {:else}
      <span class={styles.value}>{selectedLabel}</span>
    {/if}
    <CaretDownIcon class={styles.caret} aria-hidden="true" />
  </button>
  {#if isOpen && options}
    <ul
      id={`${id}-listbox`}
      class={styles.listbox}
      role="listbox"
      aria-label={label}
      tabindex="-1"
      bind:this={list}
      onpointerdown={(event) => event.preventDefault()}
      onclick={onlistclick}
      {onkeydown}
    >
      {#each options as option, index}
        <li
          id={optionId(index)}
          class={[styles.option, index === activeIndex && styles.active]}
          role="option"
          aria-selected={index === selectedIndex}
          data-index={index}
          onpointermove={() => (activeIndex = index)}
        >
          {option.label}
        </li>
      {/each}
    </ul>
  {/if}
</div>
