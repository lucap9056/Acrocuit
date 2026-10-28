<script lang="ts">
  import DoorOpenIcon from "phosphor-svelte/lib/DoorOpenIcon";
  import PencilSimpleIcon from "phosphor-svelte/lib/PencilSimpleIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import type { AuthSession, SpaceGroup } from "@/api";
  import dialog from "@/components/dialog";
  import load from "@/components/loading-bar";
  import { tooltip } from "@/components/tooltip";
  import {
    addSpaceGroup,
    fetchSpaceGroups,
    leaveSpaceGroup,
    removeSpaceGroup,
    transferSpaceGroupOwner,
    updateSpaceGroup,
  } from "@/lib/cache/space-groups";
  import type { HashRouter } from "@/lib/hash-routers";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import { canEdit, canManage } from "@/lib/permissions";

  import SpaceGroupEditor from "./space-group-editor/space-group-editor.svelte";
  import styles from "./space-group-selector.module.css";

  interface Props {
    session: AuthSession;
    router: HashRouter;
  }

  const { session, router }: Props = $props();
  let spaceGroups = $state<SpaceGroup[]>([]);
  let loadError = $state<string>();
  let isLoaded = $state(false);

  let spaceGroupName = $state("");
  let editingSpaceGroup = $state<SpaceGroup>();

  const showSpaceGroups = (sgs: SpaceGroup[]) => {
    spaceGroups = sgs;
  };

  const loadSpaceGroups = (s: AuthSession, signal: AbortSignal) => {
    const loading = load.begin(t("loading.fetchSpaceGroups"));
    fetchSpaceGroups(s)
      .then((sgs) => {
        if (!signal.aborted) {
          showSpaceGroups(sgs);
          loadError = undefined;
        }
      })
      .catch((err) => {
        if (!signal.aborted) {
          console.log(err);
        }
      })
      .finally(() => {
        if (!signal.aborted) {
          isLoaded = true;
        }
        loading.release();
      });
  };

  onMount((signal) => {
    loadSpaceGroups(session, signal);
  });

  const showDialog = (text: string, name: string = ""): Promise<string> => {
    spaceGroupName = name;

    const { promise, resolve, reject } = Promise.withResolvers<string>();

    dialog.push({
      text: "",
      content: setSpaceGroupName,
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
            resolve(spaceGroupName);
            return true;
          },
        },
      ],
    });

    return promise;
  };

  const create = () => {
    showDialog(t("common.create")).then((name) => {
      const loading = load.begin(t("loading.createSpaceGroup"));
      addSpaceGroup(session, name)
        .then(showSpaceGroups)
        .catch((err) => {
          console.error(err);
        })
        .finally(() => {
          loading.release();
        });
    });
  };

  const rename = (sg: SpaceGroup, name: string) => {
    const loading = load.begin(t("loading.renameSpaceGroup"));
    return updateSpaceGroup(session, { ...sg, name })
      .then((sgs) => {
        spaceGroups = sgs;
        editingSpaceGroup = sgs.find(({ id }) => id === sg.id);
      })
      .finally(() => {
        loading.release();
      });
  };

  const transfer = (sg: SpaceGroup, email: string) => {
    const loading = load.begin(t("loading.transferSpaceGroupOwner"));
    return transferSpaceGroupOwner(session, sg, email)
      .then((sgs) => {
        spaceGroups = sgs;
        editingSpaceGroup = sgs.find(({ id }) => id === sg.id);
      })
      .finally(() => {
        loading.release();
      });
  };

  const edit = (sg: SpaceGroup) => {
    editingSpaceGroup = sg;
    dialog.push({
      content: spaceGroupEditor,
      buttons: [
        {
          text: t("common.close"),
          type: "",
          onclick: () => {
            editingSpaceGroup = undefined;
            return true;
          },
        },
      ],
    });
  };

  const confirmDrop = (
    text: string,
    action: string,
    loadingText: string,
    request: () => Promise<SpaceGroup[]>,
  ) => {
    dialog.push({
      text,
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => true,
        },
        {
          text: action,
          type: "danger",
          onclick: () => {
            const loading = load.begin(loadingText);
            request()
              .then((sgs) => {
                spaceGroups = sgs;
              })
              .catch((err) => {
                console.error(err);
              })
              .finally(() => {
                loading.release();
              });
            return true;
          },
        },
      ],
    });
  };

  const remove = (sg: SpaceGroup) => {
    confirmDrop(
      t("spaceGroups.confirmRemove", { name: sg.name }),
      t("common.delete"),
      t("loading.removeSpaceGroup"),
      () => removeSpaceGroup(session, sg),
    );
  };

  const leave = (sg: SpaceGroup) => {
    confirmDrop(
      t("spaceGroups.confirmLeave", { name: sg.name }),
      t("spaceGroups.leave"),
      t("loading.leaveSpaceGroup"),
      () => leaveSpaceGroup(session, sg),
    );
  };

  const onselected = (spaceGroupId: number) => {
    const spaceGroupIdStr = spaceGroupId.toString();
    const routePath = `space-groups/${spaceGroupIdStr}`;
    router.to(routePath);
  };
</script>

{#snippet setSpaceGroupName()}
  <input type="text" bind:value={spaceGroupName}>
{/snippet}

{#snippet spaceGroupEditor()}
  {#if editingSpaceGroup}
    {@const sg = editingSpaceGroup}
    <SpaceGroupEditor
      {session}
      spaceGroup={sg}
      onrename={(name) => rename(sg, name)}
      ontransfer={(email) => transfer(sg, email)}
    />
  {/if}
{/snippet}

<div class="space_groups {styles.space_groups}">
  <div class="heading {styles.heading}">
    <h1 class={styles.title}>{t("spaceGroups.title")}</h1>
    <button class={styles.create} onclick={create}>
      {t("spaceGroups.new")}
    </button>
  </div>

  {#if loadError}
    <p class={styles.message}>{loadError}</p>
  {:else if isLoaded && spaceGroups.length === 0}
    <div>
      <p class={styles.message}>
        {t("spaceGroups.empty")}
      </p>
    </div>
  {:else}
    <ul class={styles.list}>
      {#each spaceGroups as sg (sg.id)}
        <li class={styles.item}>
          <button class={styles.select} onclick={() => onselected(sg.id)}>
            <span class={styles.name}>{sg.name}</span>
            <span class={styles.role} data-role={sg.role}>
              {t(`spaceGroups.role.${sg.role}`)}
            </span>
          </button>
          <div class={styles.actions}>
            {#if canEdit(sg)}
              <button
                class="icon_button"
                onclick={() => edit(sg)}
                {@attach tooltip(t("spaceGroups.edit"))}
              >
                <PencilSimpleIcon />
              </button>
            {/if}
            {#if canManage(sg)}
              <button
                class="icon_button {styles.delete}"
                onclick={() => remove(sg)}
                {@attach tooltip(t("common.delete"))}
              >
                <TrashIcon />
              </button>
            {:else}
              <button
                class="icon_button {styles.delete}"
                onclick={() => leave(sg)}
                {@attach tooltip(t("spaceGroups.leave"))}
              >
                <DoorOpenIcon />
              </button>
            {/if}
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</div>
