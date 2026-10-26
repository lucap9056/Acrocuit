<script lang="ts">
  import type { AuthSession, Space, SpaceGroup } from "@/api";
  import back from "@/components/back-button";
  import headerTitle from "@/components/header-title";
  import load from "@/components/loading-bar";
  import { fetchSpaces } from "@/lib/cache/spaces";
  import {
    field,
    type HashRouter,
    type RouteHandler,
    type RouteParams,
  } from "@/lib/hash-routers";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import type { Workspace } from "@/lib/workspace";
  import FloorInfo from "./floor-info/floor-info.svelte";
  import { FloorOverview } from "./floor-overview.svelte";
  import FloorOverviewCanvas from "./floor-overview-canvas/floor-overview-canvas.svelte";
  import FloorOverviewToolbar from "./floor-overview-toolbar/floor-overview-toolbar.svelte";
  import styles from "./space-group-view.module.css";
  import SpaceList from "./space-list/space-list.svelte";
  import SpaceView from "./space-view/space-view.svelte";

  interface Props {
    session: AuthSession;
    workspace: Workspace;
    router: HashRouter;
    spaceGroup: SpaceGroup;
  }

  const { session, workspace, router, spaceGroup }: Props = $props();

  let spaces = $state<Space[]>();
  let selectedSpace = $state<Space>();
  let spaceGroupRouter = $state.raw<HashRouter>();

  const overview = new FloorOverview();

  const onspacesupdate = (updatedSpaces: Space[]) => {
    spaces = updatedSpaces;
    selectedSpace = updatedSpaces.find(({ id }) => id === selectedSpace?.id);
  };

  const onSpaceSelected = (
    signal: AbortSignal,
    scopedRouter: HashRouter,
    spaceIdStr: string,
  ) => {
    const spaceId = parseInt(spaceIdStr, 10);
    if (Number.isNaN(spaceId)) {
      scopedRouter.back();
      return;
    }

    const loading = load.begin(t("loading.fetchSpace"));
    fetchSpaces(session, spaceGroup.id)
      .then((s) => {
        if (!signal.aborted) {
          selectedSpace = s.find(({ id }) => id === spaceId);
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
    workspace.push(
      {
        leftpart: spaceList,
        centerpart: floorView,
        toolbarpart: viewToolbar,
        rightpart: spaceInfo,
      },
      signal,
    );
    back.push(() => router.backTo(""), signal);

    headerTitle.push(title, signal);

    const loading = load.begin(t("loading.fetchSpaces"));

    fetchSpaces(session, spaceGroup.id)
      .then((s) => {
        if (!signal.aborted) {
          spaces = s;
        }
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
      });

    const scopedRouter = router.child(`space-groups/${spaceGroup.id}`);
    spaceGroupRouter = scopedRouter;
    const paramName = "spaceId";
    const routePath = `spaces/:${paramName}/*`;

    const onSpaceSelectedRouteHandle: RouteHandler = (params?: RouteParams) => {
      if (params !== undefined) {
        const spaceId = field(params, paramName, "");

        onSpaceSelected(signal, scopedRouter, spaceId);
      } else if (!signal.aborted) {
        selectedSpace = undefined;
      }
    };

    const params = scopedRouter.on(
      routePath,
      onSpaceSelectedRouteHandle,
      signal,
    );
    onSpaceSelectedRouteHandle(params);

    return () => {
      spaceGroupRouter = undefined;
    };
  });
</script>

{#snippet spaceList()}
  {#if spaceGroupRouter}
    <SpaceList
      {session}
      {spaceGroup}
      router={spaceGroupRouter}
      bind:spaces
      onspacehover={(id) => overview.revealSpace(id)}
    />
  {/if}
{/snippet}

{#snippet title()}
  <span class={styles.space_group_name}>{spaceGroup.name}</span>
{/snippet}

{#snippet viewToolbar()}
  <FloorOverviewToolbar {overview} />
{/snippet}

{#snippet spaceInfo()}
  {#if spaces}
    <FloorInfo {session} {spaces} {overview} />
  {/if}
{/snippet}

{#snippet floorView()}
  <FloorOverviewCanvas {session} {spaceGroup} {spaces} {overview} />
{/snippet}

{#if spaceGroupRouter}
  {#each selectedSpace ? [selectedSpace] : [] as space (space.id)}
    <SpaceView
      {session}
      {workspace}
      router={spaceGroupRouter}
      {spaceGroup}
      {space}
      {onspacesupdate}
    />
  {/each}
{/if}
