<script lang="ts" module>
  type SwitchLayout = "rows" | "columns";

  const view = $state<{ layout: SwitchLayout }>({ layout: "rows" });
</script>

<script lang="ts">
  import ColumnsIcon from "phosphor-svelte/lib/ColumnsIcon";
  import DotsSixVerticalIcon from "phosphor-svelte/lib/DotsSixVerticalIcon";
  import PlusIcon from "phosphor-svelte/lib/PlusIcon";
  import RowsIcon from "phosphor-svelte/lib/RowsIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import { flip } from "svelte/animate";
  import {
    type DndEvent,
    dragHandle,
    dragHandleZone,
    SHADOW_ITEM_MARKER_PROPERTY_NAME,
  } from "svelte-dnd-action";
  import type { AuthSession, Switch, SwitchPanel } from "@/api";
  import dialog from "@/components/dialog";
  import load from "@/components/loading-bar";
  import Spinner from "@/components/spinner/spinner.svelte";
  import Textbox from "@/components/textbox/textbox.svelte";
  import { tooltip } from "@/components/tooltip";
  import {
    addSwitch,
    fetchSwitches,
    moveSwitch,
    removeSwitch,
    watchSwitches,
  } from "@/lib/cache/switches";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import styles from "./switch-panel-editor.module.css";

  interface Props {
    session: AuthSession;
    switchPanel: SwitchPanel;
    onrename: (name: string) => Promise<unknown>;
    oneditswitch: (sw: Switch) => void;
  }

  const { session, switchPanel, onrename, oneditswitch }: Props = $props();

  type DndSwitch = Switch & { [SHADOW_ITEM_MARKER_PROPERTY_NAME]?: boolean };

  const FLIP_DURATION_MS = 150;

  let switches = $state<Switch[]>();
  let isPending = $state(false);
  let newSwitchName = $state("");
  let draftItems = $state<DndSwitch[]>();
  let isConfirmingMove = $state(false);

  const items: DndSwitch[] = $derived(draftItems ?? switches ?? []);

  const canAdd = $derived(newSwitchName.trim() !== "" && !isPending);

  const withSwitchPanel = (sw: Switch) => ({
    ...sw,
    switch_panel: { id: switchPanel.id },
  });

  onMount((signal) => {
    watchSwitches(
      session,
      switchPanel.id,
      (sws) => {
        switches = sws;
      },
      signal,
    );
    fetchSwitches(session, switchPanel.id)
      .then((sws) => {
        if (!signal.aborted) {
          switches = sws;
        }
      })
      .catch((err) => {
        console.error(err);
      });
  });

  const update = (message: string, request: () => Promise<Switch[]>) => {
    isPending = true;
    const loading = load.begin(message);
    return request()
      .then((sws) => {
        switches = sws;
      })
      .finally(() => {
        isPending = false;
        loading.release();
      });
  };

  const add = () => {
    if (!canAdd) {
      return;
    }
    const name = newSwitchName.trim();
    update(t("loading.createSwitch"), () =>
      addSwitch(session, switchPanel.id, name),
    )
      .then(() => {
        newSwitchName = "";
      })
      .catch((err) => {
        console.error(err);
      });
  };

  const onnewswitchkeydown = (event: KeyboardEvent) => {
    if (event.key === "Enter" && !event.isComposing) {
      event.preventDefault();
      add();
    }
  };

  const confirmAction = (text: string, action: string): Promise<void> => {
    const { promise, resolve, reject } = Promise.withResolvers<void>();

    dialog.push({
      text,
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => {
            reject();
            return true;
          },
        },
        {
          text: action,
          type: "",
          onclick: () => {
            resolve();
            return true;
          },
        },
      ],
    });

    return promise;
  };

  const onconsider = (event: CustomEvent<DndEvent<DndSwitch>>) => {
    draftItems = event.detail.items;
  };

  const onfinalize = (event: CustomEvent<DndEvent<DndSwitch>>) => {
    const draggedId = String(event.detail.info.id);
    const finalItems = event.detail.items;
    const current = switches ?? [];

    const fromIndex = current.findIndex((sw) => String(sw.id) === draggedId);
    const toIndex = finalItems.findIndex((sw) => String(sw.id) === draggedId);

    if (fromIndex === -1 || toIndex === -1 || fromIndex === toIndex) {
      draftItems = undefined;
      return;
    }

    const sw = current[fromIndex];
    const target = current[toIndex];
    draftItems = finalItems;
    isConfirmingMove = true;

    confirmAction(
      t("switches.confirmMove", { name: sw.name, position: toIndex + 1 }),
      t("common.move"),
    )
      .then(() => {
        update(t("loading.moveSwitch"), () =>
          moveSwitch(session, {
            ...withSwitchPanel(sw),
            display_order: target.display_order,
          }),
        ).catch((err) => {
          console.error(err);
        });
      })
      .catch(() => {})
      .finally(() => {
        draftItems = undefined;
        isConfirmingMove = false;
      });
  };

  const remove = (sw: Switch) => {
    dialog.push({
      text: t("switches.confirmDelete", { name: sw.name }),
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => true,
        },
        {
          text: t("common.delete"),
          type: "danger",
          onclick: () => {
            update(t("loading.removeSwitch"), () =>
              removeSwitch(session, withSwitchPanel(sw)),
            ).catch((err) => {
              console.error(err);
            });
            return true;
          },
        },
      ],
    });
  };
</script>

{#snippet layoutButton(layout: SwitchLayout, label: string, Icon: typeof RowsIcon)}
  <button
    class={["icon_button", view.layout === layout && styles.active]}
    aria-label={label}
    aria-pressed={view.layout === layout}
    onclick={() => (view.layout = layout)}
    {@attach tooltip(label)}
  >
    <Icon />
  </button>
{/snippet}

<div class={styles.editor}>
  <Textbox
    class={styles.title}
    label={t("switchPanels.name")}
    value={switchPanel.name}
    onsave={onrename}
  />
  <div class={styles.toolbar}>
    <input
      class={styles.input}
      type="text"
      aria-label={t("switches.name")}
      placeholder={t("switches.name")}
      disabled={isPending}
      bind:value={newSwitchName}
      onkeydown={onnewswitchkeydown}
    >
    <button
      class="icon_button {styles.add}"
      aria-label={t("switches.new")}
      disabled={!canAdd}
      onclick={add}
      {@attach tooltip(t("switches.new"))}
    >
      <PlusIcon />
    </button>
    <div class={styles.layout} role="group" aria-label={t("switches.layout")}>
      {@render layoutButton("rows", t("switches.rows"), RowsIcon)}
      {@render layoutButton("columns", t("switches.columns"), ColumnsIcon)}
    </div>
  </div>
  {#if switches === undefined}
    <div class={styles.loading} role="status" aria-label={t("common.loading")}>
      <Spinner />
    </div>
  {:else if switches.length === 0}
    <p class={styles.message}>{t("common.noneYet")}</p>
  {:else}
    <ul
      class={[styles.switches, styles[view.layout]]}
      use:dragHandleZone={{
        items,
        flipDurationMs: FLIP_DURATION_MS,
        dragDisabled: isConfirmingMove || isPending,
        dropTargetStyle: {},
        delayTouchStart: true,
      }}
      {onconsider}
      {onfinalize}
    >
      {#each items as sw (sw.id)}
        <li
          class={[
            styles.switch,
            sw[SHADOW_ITEM_MARKER_PROPERTY_NAME] && styles.shadow,
          ]}
          animate:flip={{ duration: FLIP_DURATION_MS }}
        >
          <button
            class={["icon_button", styles.handle]}
            aria-label={t("switches.dragToReorder")}
            use:dragHandle
          >
            <DotsSixVerticalIcon />
          </button>
          <button class={styles.open} onclick={() => oneditswitch(sw)}>
            <span class={styles.toggle} aria-hidden="true">
              <span class={styles.lever}></span>
            </span>
            <span class={styles.name}>{sw.name}</span>
          </button>
          <button
            class="icon_button {styles.delete}"
            aria-label={t("common.delete")}
            disabled={isPending}
            onclick={() => remove(sw)}
            {@attach tooltip(t("common.delete"))}
          >
            <TrashIcon />
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</div>
