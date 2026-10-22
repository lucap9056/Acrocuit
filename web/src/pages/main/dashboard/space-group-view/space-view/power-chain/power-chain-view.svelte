<script lang="ts" module>
  import type { ViewportState } from "@/components/iso-scene/camera/viewport";
  import type { FloorPlanTone } from "@/lib/iso-scene/floor-plan-image";

  type ViewMode = "tree" | "scene";

  interface SavedView {
    mode: ViewMode;
    tilt: number;
    tone: FloorPlanTone;
    sceneViewport?: ViewportState;
  }

  const savedViews = new Map<string, SavedView>();
</script>

<script lang="ts">
  import CornersOutIcon from "phosphor-svelte/lib/CornersOutIcon";
  import CubeIcon from "phosphor-svelte/lib/CubeIcon";
  import TreeStructureIcon from "phosphor-svelte/lib/TreeStructureIcon";
  import type { Snippet } from "svelte";
  import type { AuthSession, Space, SpaceGroup, SpaceId } from "@/api";
  import {
    groupColor,
    paletteForTone,
  } from "@/components/iso-scene/paint/palette";
  import back from "@/components/back-button";
  import headerTitle from "@/components/header-title";
  import load from "@/components/loading-bar";
  import { tooltip } from "@/components/tooltip";
  import { fetchSpaces } from "@/lib/cache/spaces";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import {
    type ChainDetail,
    type ChainNode,
    type ChainPlace,
    countPlacesBySpace,
    describeNode,
    describePlace,
    detailPlaceKey,
    detailSpaceId,
    floorGroups,
    indexChain,
    type PowerChain,
    spannedSpaces,
  } from "@/lib/power-chain";
  import type { Workspace } from "@/lib/workspace";
  import FloorInfo from "@/pages/main/dashboard/space-group-view/floor-info/floor-info.svelte";
  import { FloorOverview } from "@/pages/main/dashboard/space-group-view/floor-overview.svelte";
  import FloorOverviewToolbar from "@/pages/main/dashboard/space-group-view/floor-overview-toolbar/floor-overview-toolbar.svelte";
  import ChainFloorList from "./chain-floor-list/chain-floor-list.svelte";
  import ChainNodeInfo from "./chain-node-info/chain-node-info.svelte";
  import ChainSceneCanvas from "./chain-scene-canvas/chain-scene-canvas.svelte";
  import ChainTree from "./chain-tree/chain-tree.svelte";
  import styles from "./power-chain-view.module.css";

  interface Props {
    session: AuthSession;
    workspace: Workspace;
    spaceGroup: SpaceGroup;
    title: Snippet;
    focusIcon: Snippet;
    emptyText: string;
    load: () => Promise<PowerChain>;
    onback: () => void;
    onopennode: (node: ChainNode) => void;
    onopenplace: (place: ChainPlace) => void;
  }

  const {
    session,
    workspace,
    spaceGroup,
    title,
    focusIcon,
    emptyText,
    load: loadChain,
    onback,
    onopennode,
    onopenplace,
  }: Props = $props();

  let chain = $state.raw<PowerChain>();
  let spaces = $state.raw<Space[]>();
  let mode = $state<ViewMode>("tree");
  let tree = $state.raw<ReturnType<typeof ChainTree>>();
  let inspected = $state.raw<ChainDetail>();
  let sceneViewport = $state.raw<ViewportState>();

  const overview = new FloorOverview();

  const index = $derived(chain ? indexChain(chain) : undefined);
  const places = $derived(
    new Map((chain?.places ?? []).map((place) => [place.key, place])),
  );
  const inspectedPlaceKey = $derived(
    inspected ? detailPlaceKey(inspected) : undefined,
  );

  const inspectSpace = (id: SpaceId) => {
    overview.inspectedSpaceId = id;
    if (inspected && detailSpaceId(inspected) !== id) {
      inspected = undefined;
    }
  };

  const inspectDetail = (detail: ChainDetail) => {
    inspected = detail;
    const spaceId = detailSpaceId(detail);
    if (spaceId !== undefined) {
      overview.inspectedSpaceId = spaceId;
    }
  };

  const inspectNode = (node: ChainNode) => {
    if (index) {
      inspectDetail(describeNode(index, node));
    }
  };

  const inspectPlace = (key: string) => {
    const place = places.get(key);
    if (index && place) {
      inspectDetail(describePlace(index, place));
    }
  };

  const openPlace = (key: string) => {
    const place = places.get(key);
    if (place) {
      onopenplace(place);
    }
  };

  const restoreView = (id: string) => {
    const saved = savedViews.get(id);
    if (saved) {
      mode = saved.mode;
      overview.tilt = saved.tilt;
      overview.tone = saved.tone;
      sceneViewport = saved.sceneViewport;
    }
  };

  const floors = $derived(
    chain && spaces ? spannedSpaces(chain, spaces) : undefined,
  );
  const placeCounts = $derived(
    chain ? countPlacesBySpace(chain) : new Map<SpaceId, number>(),
  );
  const floorColors = $derived.by(() => {
    if (mode !== "scene" || !chain || !floors) {
      return undefined;
    }
    const palette = paletteForTone(overview.tone);
    return new Map(
      [...floorGroups(chain, floors)].flatMap(([id, group]) => {
        const color = groupColor(palette, group);
        return color ? [[id, color] as const] : [];
      }),
    );
  });

  onMount((signal) => {
    workspace.push(
      {
        leftpart: floorList,
        centerpart: content,
        toolbarpart: toolbar,
        rightpart: floorInfo,
      },
      signal,
    );
    back.push(onback, signal);
    headerTitle.push(title, signal);

    const loading = load.begin(t("loading.fetchPowerChain"));
    Promise.all([loadChain(), fetchSpaces(session, spaceGroup.id)])
      .then(([c, s]) => {
        if (!signal.aborted) {
          restoreView(c.id);
          chain = c;
          spaces = s;
        }
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
      });

    return () => {
      if (chain) {
        savedViews.set(chain.id, {
          mode,
          tilt: overview.tilt,
          tone: overview.tone,
          sceneViewport,
        });
      }
    };
  });
</script>

{#snippet floorList()}
  <ChainFloorList
    {floors}
    focusSpaceId={chain?.focusSpaceId}
    {focusIcon}
    {placeCounts}
    {floorColors}
    inspectedSpaceId={overview.inspectedSpaceId}
    onfloorhover={(id) => {
      inspectSpace(id);
      overview.revealSpace(id);
    }}
  />
{/snippet}

{#snippet toolbar()}
  <div class={styles.toolbar}>
    <div class={styles.modes}>
      <button
        class={["icon_button", mode === "tree" && styles.active]}
        aria-label={t("powerChain.tree")}
        aria-pressed={mode === "tree"}
        onclick={() => (mode = "tree")}
        {@attach tooltip(t("powerChain.tree"))}
      >
        <TreeStructureIcon />
      </button>
      <button
        class={["icon_button", mode === "scene" && styles.active]}
        aria-label={t("powerChain.scene")}
        aria-pressed={mode === "scene"}
        disabled={floors === undefined}
        onclick={() => (mode = "scene")}
        {@attach tooltip(t("powerChain.scene"))}
      >
        <CubeIcon />
      </button>
    </div>
    {#if mode === "scene"}
      <FloorOverviewToolbar {overview} />
    {:else}
      <button
        class={["icon_button", styles.fit]}
        aria-label={t("floorOverview.fitToView")}
        disabled={tree === undefined}
        onclick={() => tree?.fit()}
        {@attach tooltip(t("floorOverview.fitToView"))}
      >
        <CornersOutIcon />
      </button>
    {/if}
  </div>
{/snippet}

{#snippet content()}
  {#if chain === undefined || floors === undefined}
    <div class={styles.placeholder} role="status">
      {t("powerChain.loading")}
    </div>
  {:else if mode === "tree"}
    <ChainTree
      bind:this={tree}
      chainId={chain.id}
      label={chain.title}
      roots={chain.roots}
      sink={chain.sink}
      {emptyText}
      spaces={spaces ?? []}
      inspectedSpaceId={overview.inspectedSpaceId}
      oninspectnode={inspectNode}
      {onopennode}
    />
  {:else}
    <ChainSceneCanvas
      {session}
      {spaceGroup}
      {floors}
      {chain}
      {overview}
      highlightMarkerId={inspectedPlaceKey}
      initialViewport={sceneViewport}
      oninspectspace={inspectSpace}
      oninspectplace={inspectPlace}
      onopenplace={openPlace}
      onviewportchange={(viewport) => {
        sceneViewport = viewport;
      }}
    />
  {/if}
{/snippet}

{#snippet floorInfo()}
  {#if spaces}
    <FloorInfo {session} {spaces} {overview} />
    <ChainNodeInfo detail={inspected} {spaces} />
  {/if}
{/snippet}
