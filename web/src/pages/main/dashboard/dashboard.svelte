<script lang="ts">
  import SignOutIcon from "phosphor-svelte/lib/SignOutIcon";
  import api, { type AuthSession, type SpaceGroup } from "@/api";
  import BackButton from "@/components/back-button/back-button.svelte";
  import HeaderTitle from "@/components/header-title/header-title.svelte";
  import load from "@/components/loading-bar";
  import { tooltip } from "@/components/tooltip";
  import WorkspaceView from "@/components/workspace-view/workspace-view.svelte";
  import { fetchSpaceGroups } from "@/lib/cache/space-groups";
  import type {
    HashRouter,
    RouteHandler,
    RouteParams,
  } from "@/lib/hash-routers";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import { createWorkspace } from "@/lib/workspace";
  import styles from "./dashboard.module.css";
  import SpaceGroupSelector from "./space-group-selector.svelte";
  import SpaceGroupView from "./space-group-view/space-group-view.svelte";

  interface Props {
    session: AuthSession;
    router: HashRouter;
  }
  const { session, router }: Props = $props();

  const workspace = $state(createWorkspace());
  let spaceGroup = $state<SpaceGroup>();

  const onSpaceGroupSelected = (
    signal: AbortSignal,
    spaceGroupIdStr: string,
  ) => {
    const spaceGroupId = parseInt(spaceGroupIdStr, 10);
    if (Number.isNaN(spaceGroupId)) {
      router.back("");
      return;
    }

    const loading = load.begin(t("loading.fetchSpaceGroups"));
    fetchSpaceGroups(session)
      .then((spaceGroups) => {
        if (!signal.aborted) {
          spaceGroup = spaceGroups.find((sg) => sg.id === spaceGroupId);
        }
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
      });
  };

  onMount((signal) => {
    const paramName = "spaceGroupId";
    const routePath = `space-groups/:${paramName}/*`;

    const onSpaceGroupSelectedRouteHandle: RouteHandler = (
      params?: RouteParams,
    ) => {
      if (params !== undefined && typeof params[paramName] === "string") {
        const spaceGroupId = params[paramName];
        onSpaceGroupSelected(signal, spaceGroupId);
      } else {
        if (!signal.aborted) {
          spaceGroup = undefined;
        }
      }
    };
    const params = router.on(
      routePath,
      onSpaceGroupSelectedRouteHandle,
      signal,
    );
    onSpaceGroupSelectedRouteHandle(params);
  });

  const logout = () => {
    const loading = load.begin(t("loading.signOut"));
    api.auth
      .logout()
      .then(() => {
        location.assign("/");
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
      });
  };
</script>

<div class={styles.dashboard}>
  <div class={styles.header}>
    <div class={styles.left}>
      <BackButton />
    </div>
    <div class={styles.center}>
      <HeaderTitle />
    </div>
    <div class={styles.right}>
      <button
        class="icon_button"
        onclick={logout}
        {@attach tooltip(t("auth.signOut"))}
      >
        <SignOutIcon />
      </button>
    </div>
  </div>
  {#each spaceGroup
    ? [spaceGroup]
    : [] as currentSpaceGroup (currentSpaceGroup.id)}
    <WorkspaceView {workspace} />
    <SpaceGroupView
      {session}
      {workspace}
      {router}
      spaceGroup={currentSpaceGroup}
    />
  {:else}
    <div class={styles.space_groups}>
      <SpaceGroupSelector {session} {router} />
    </div>
  {/each}
</div>
