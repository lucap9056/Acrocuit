<script lang="ts">
  import DotsSixVerticalIcon from "phosphor-svelte/lib/DotsSixVerticalIcon";
  import PencilSimpleIcon from "phosphor-svelte/lib/PencilSimpleIcon";
  import PlusIcon from "phosphor-svelte/lib/PlusIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import { flip } from "svelte/animate";
  import {
    type DndEvent,
    dragHandle,
    dragHandleZone,
    SHADOW_ITEM_MARKER_PROPERTY_NAME,
  } from "svelte-dnd-action";
  import type { AuthSession, Space, SpaceGroup, SpaceId } from "@/api";
  import dialog from "@/components/dialog";
  import load from "@/components/loading-bar";
  import { tooltip } from "@/components/tooltip";
  import {
    addSpace,
    moveSpace,
    removeSpace,
    renameSpace,
  } from "@/lib/cache/spaces";
  import type { HashRouter } from "@/lib/hash-routers";
  import { t } from "@/lib/i18n";
  import styles from "./space-list.module.css";

  interface Props {
    session: AuthSession;
    spaceGroup: SpaceGroup;
    spaces?: Space[];
    router: HashRouter;
    onspacehover?: (id: SpaceId) => void;
  }

  let {
    session,
    spaceGroup,
    router,
    spaces = $bindable(),
    onspacehover,
  }: Props = $props();

  type DndSpace = Space & { [SHADOW_ITEM_MARKER_PROPERTY_NAME]?: boolean };

  const FLIP_DURATION_MS = 150;

  let spaceName = $state("");
  let draftItems = $state<DndSpace[]>();
  let isConfirmingMove = $state(false);

  const items: DndSpace[] = $derived(draftItems ?? spaces ?? []);

  const hover = (space: Space) => {
    if (draftItems === undefined) {
      onspacehover?.(space.id);
    }
  };

  const withSpaceGroup = (space: Space) => ({
    ...space,
    space_group: { id: spaceGroup.id },
  });

  const showNameDialog = (text: string, name: string = ""): Promise<string> => {
    spaceName = name;

    const { promise, resolve, reject } = Promise.withResolvers<string>();

    dialog.push({
      text: "",
      content: setSpaceName,
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
            resolve(spaceName);
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

  const create = () => {
    showNameDialog(t("common.create")).then((name) => {
      const loading = load.begin(t("loading.createSpace"));
      addSpace(session, spaceGroup.id, name)
        .then((s) => {
          spaces = s;
        })
        .catch((err) => {
          console.error(err);
        })
        .finally(() => {
          loading.release();
        });
    });
  };

  const rename = (space: Space) => {
    showNameDialog(t("common.rename"), space.name).then((name) => {
      const loading = load.begin(t("loading.renameSpace"));
      renameSpace(session, withSpaceGroup(space), name)
        .then((s) => {
          spaces = s;
        })
        .catch((err) => {
          console.error(err);
        })
        .finally(() => {
          loading.release();
        });
    });
  };

  const remove = (space: Space) => {
    confirmAction("", t("common.remove")).then(() => {
      const loading = load.begin(t("loading.removeSpace"));
      removeSpace(session, withSpaceGroup(space))
        .then((s) => {
          spaces = s;
        })
        .catch((err) => {
          console.error(err);
        })
        .finally(() => {
          loading.release();
        });
    });
  };

  const onconsider = (event: CustomEvent<DndEvent<DndSpace>>) => {
    draftItems = [...event.detail.items].reverse();
  };

  const onfinalize = (event: CustomEvent<DndEvent<DndSpace>>) => {
    const draggedId = String(event.detail.info.id);
    const finalItems = [...event.detail.items].reverse();
    const current = spaces ?? [];

    const fromIndex = current.findIndex((sp) => String(sp.id) === draggedId);
    const toIndex = finalItems.findIndex((sp) => String(sp.id) === draggedId);

    if (fromIndex === -1 || toIndex === -1 || fromIndex === toIndex) {
      draftItems = undefined;
      return;
    }

    const space = current[fromIndex];
    const target = current[toIndex];
    draftItems = finalItems;
    isConfirmingMove = true;

    confirmAction(
      t("spaces.confirmMove", { name: space.name, position: toIndex + 1 }),
      t("common.move"),
    )
      .then(() => {
        const loading = load.begin(t("loading.moveSpace"));
        moveSpace(session, {
          ...withSpaceGroup(space),
          display_order: target.display_order,
        })
          .then((s) => {
            spaces = s;
          })
          .catch((err) => {
            console.error(err);
          })
          .finally(() => {
            loading.release();
          });
      })
      .catch(() => {})
      .finally(() => {
        draftItems = undefined;
        isConfirmingMove = false;
      });
  };

  let reversedItems = $derived([...items].reverse());
</script>

{#snippet setSpaceName()}
  <input type="text" bind:value={spaceName}>
{/snippet}

<div class={styles.space_list}>
  <div class={styles.list_heading}>
    <h2 class={styles.list_title}>{t("spaces.title")}</h2>
    <button
      class="icon_button"
      aria-label={t("spaces.new")}
      onclick={create}
      {@attach tooltip(t("spaces.new"))}
    >
      <PlusIcon />
    </button>
  </div>

  {#if spaces === undefined}
    <p class={styles.message} role="status">{t("spaces.loading")}</p>
  {:else if items.length === 0}
    <p class={styles.message}>{t("spaces.empty")}</p>
  {:else}
    <ul
      class={styles.list}
      use:dragHandleZone={{
        items: reversedItems,
        flipDurationMs: FLIP_DURATION_MS,
        dragDisabled: isConfirmingMove,
        dropTargetStyle: {},
        delayTouchStart: true,
      }}
      {onconsider}
      {onfinalize}
    >
      {#each reversedItems as space (space.id)}
        <li
          class={[
            styles.item,
            space[SHADOW_ITEM_MARKER_PROPERTY_NAME] && styles.shadow,
          ]}
          animate:flip={{ duration: FLIP_DURATION_MS }}
          onpointerenter={() => hover(space)}
          onfocusin={() => hover(space)}
        >
          <button
            class={["icon_button", styles.handle]}
            aria-label={t("spaces.dragToReorder")}
            use:dragHandle
          >
            <DotsSixVerticalIcon />
          </button>
          <button
            class={styles.select}
            onclick={() => router.to(`space/${space.id}`)}
          >
            <span class={styles.name}>{space.name}</span>
          </button>
          <div class={styles.actions}>
            <button
              class="icon_button"
              aria-label={t("common.rename")}
              onclick={() => rename(space)}
              {@attach tooltip(t("common.rename"))}
            >
              <PencilSimpleIcon />
            </button>
            <button
              class="icon_button {styles.delete}"
              aria-label={t("common.delete")}
              onclick={() => remove(space)}
              {@attach tooltip(t("common.delete"))}
            >
              <TrashIcon />
            </button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</div>
