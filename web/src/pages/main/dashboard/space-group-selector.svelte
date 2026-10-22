<script lang="ts">
  import PencilSimpleIcon from "phosphor-svelte/lib/PencilSimpleIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import type { AuthSession, SpaceGroup } from "@/api";
  import dialog from "@/components/dialog";
  import load from "@/components/loading-bar";
  import { tooltip } from "@/components/tooltip";
  import {
    addSpaceGroup,
    fetchSpaceGroups,
    removeSpaceGroup,
    updateSpaceGroup,
  } from "@/lib/cache/space-groups";
  import type { HashRouter } from "@/lib/hash-routers";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";

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

  const rename = async (sg: SpaceGroup) => {
    showDialog(t("common.rename"), sg.name).then((name) => {
      const loading = load.begin(t("loading.renameSpaceGroup"));

      updateSpaceGroup(session, { ...sg, name })
        .then((sgs) => {
          spaceGroups = sgs;
        })
        .catch((err) => {
          console.error(err);
        })
        .finally(() => {
          loading.release();
        });
    });
  };

  const remove = async (sg: SpaceGroup) => {
    new Promise((resolve, reject) => {
      dialog.push({
        text: "",
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
            text: t("common.remove"),
            type: "",
            onclick: () => {
              resolve(true);
              return true;
            },
          },
        ],
      });
    }).then(() => {
      const loading = load.begin(t("loading.removeSpaceGroup"));

      removeSpaceGroup(session, sg)
        .then((sgs) => {
          spaceGroups = sgs;
        })
        .catch((err) => {
          console.error(err);
        })
        .finally(() => {
          loading.release();
        });
    });
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
          </button>
          <div class={styles.actions}>
            <button
              class="icon_button"
              onclick={() => rename(sg)}
              {@attach tooltip(t("common.rename"))}
            >
              <PencilSimpleIcon />
            </button>
            <button
              class="icon_button {styles.delete}"
              onclick={() => remove(sg)}
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
