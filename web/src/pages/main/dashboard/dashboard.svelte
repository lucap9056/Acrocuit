<script lang="ts">
  import SignOutIcon from "phosphor-svelte/lib/SignOutIcon";
  import api, { type AuthSession, type SpaceGroup } from "@/api";
  import load from "@/components/loading-bar";
  import { tooltip } from "@/components/tooltip";
  import { fetchSpaceGroups } from "@/lib/cache/space-groups";
  import router, {
    type RouteHandler,
    type RouteParams,
  } from "@/lib/hash-routers";
  import { onMount } from "@/lib/lifecycle";
  import { createWorkspace } from "@/lib/workspace";
  import styles from "./dashboard.module.css";
  import SpaceGroupSelector from "./space-group-selector.svelte";
  import WorkspaceView from "@/components/workspace-view/workspace-view.svelte";

  interface Props {
    session: AuthSession;
  }
  const { session }: Props = $props();

  const workspace = $state(createWorkspace());
  let spaceGroup = $state<SpaceGroup>();

  const onSpaceGroupSelected = (
    signal: AbortSignal,
    spaceGroupIdStr: string,
  ) => {
    const spaceGroupId = parseInt(spaceGroupIdStr, 10);
    if (Number.isNaN(spaceGroupId)) {
      router.back();
      return;
    }

    const loading = load.begin();
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
    const routePath = `space-groups/:${paramName}`;

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
    const params = router.on(routePath, onSpaceGroupSelectedRouteHandle);
    onSpaceGroupSelectedRouteHandle(params);

    return () => {
      router.off(routePath, onSpaceGroupSelectedRouteHandle);
    };
  });

  const logout = () => {
    const loading = load.begin();
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
    <div class={styles.center}></div>
    <div class={styles.right}>
      <button class="icon_button" onclick={logout} {@attach tooltip("Logout")}>
        <SignOutIcon />
      </button>
    </div>
  </div>
  {#if spaceGroup}
    <WorkspaceView {workspace} />
  {:else}
    <div class={styles.space_groups}><SpaceGroupSelector {session} /></div>
  {/if}
</div>
