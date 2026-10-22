<script lang="ts">
  import PlusIcon from "phosphor-svelte/lib/PlusIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import type { AuthSession, SwitchPanel, Position, Space } from "@/api";
  import dialog from "@/components/dialog";
  import load from "@/components/loading-bar";
  import { tooltip } from "@/components/tooltip";
  import {
    addSwitchPanel,
    removeSwitchPanel,
  } from "@/lib/cache/switch-panels";
  import { t } from "@/lib/i18n";
  import {
    switchPanelKey,
    type SpaceFocus,
  } from "@/pages/main/dashboard/space-group-view/space-view/space-focus.svelte";
  import styles from "./switch-panel-list.module.css";

  interface Props {
    session: AuthSession;
    space: Space;
    switchPanels?: SwitchPanel[];
    focus: SpaceFocus;
    onedit: (switchPanel: SwitchPanel) => void;
  }

  let {
    session,
    space,
    switchPanels = $bindable(),
    focus,
    onedit,
  }: Props = $props();

  const INITIAL_POSITION: Position = { x: 0, y: 0, z: 0 };

  let switchPanelName = $state("");

  const withSpace = (switchPanel: SwitchPanel) => ({
    ...switchPanel,
    space: { id: space.id },
  });

  const showNameDialog = (text: string, name: string = ""): Promise<string> => {
    switchPanelName = name;

    const { promise, resolve, reject } = Promise.withResolvers<string>();

    dialog.push({
      text: "",
      content: setSwitchPanelName,
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
          text,
          type: "",
          onclick: () => {
            resolve(switchPanelName);
            return true;
          },
        },
      ],
    });

    return promise;
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

  const update = (message: string, request: Promise<SwitchPanel[]>) => {
    const loading = load.begin(message);
    request
      .then((bgs) => {
        switchPanels = bgs;
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
      });
  };

  const create = () => {
    showNameDialog(t("common.create")).then(
      (name) => {
        update(
          t("loading.createSwitchPanel"),
          addSwitchPanel(session, space.id, name, INITIAL_POSITION),
        );
      },
      () => {},
    );
  };

  const remove = (switchPanel: SwitchPanel) => {
    confirmAction(
      t("switchPanels.confirmDelete", { name: switchPanel.name }),
      t("common.delete"),
    ).then(
      () => {
        update(
          t("loading.removeSwitchPanel"),
          removeSwitchPanel(session, withSpace(switchPanel)),
        );
      },
      () => {},
    );
  };
</script>

{#snippet setSwitchPanelName()}
  <input type="text" bind:value={switchPanelName}>
{/snippet}

<section class={styles.section}>
  <div class={styles.heading}>
    <h2 class={styles.title}>{t("switchPanels.title")}</h2>
    <button
      class="icon_button"
      aria-label={t("switchPanels.new")}
      onclick={create}
      {@attach tooltip(t("switchPanels.new"))}
    >
      <PlusIcon />
    </button>
  </div>
  {#if switchPanels === undefined}
    <p class={styles.message} role="status">{t("common.loading")}</p>
  {:else if switchPanels.length === 0}
    <p class={styles.message}>{t("common.noneYet")}</p>
  {:else}
    <ul class={styles.list}>
      {#each switchPanels as switchPanel (switchPanel.id)}
        {@const key = switchPanelKey(switchPanel.id)}
        <li
          class={[styles.item, focus.hoveredKey === key && styles.focused]}
          onpointerenter={() => {
            focus.enter(key);
            focus.scheduleReveal(key);
          }}
          onpointerleave={() => {
            focus.leave(key);
            focus.cancelReveal();
          }}
        >
          <button
            class={styles.name}
            onclick={() => onedit(switchPanel)}
            onfocus={() => focus.scheduleReveal(key)}
            onblur={() => focus.cancelReveal()}
          >
            {switchPanel.name}
          </button>
          <div class={styles.actions}>
            <button
              class="icon_button {styles.delete}"
              aria-label={t("common.delete")}
              onclick={() => remove(switchPanel)}
              {@attach tooltip(t("common.delete"))}
            >
              <TrashIcon />
            </button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</section>
