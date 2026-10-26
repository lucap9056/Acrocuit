<script lang="ts" module>
  import type { ViewportState } from "@/components/iso-scene/camera/viewport";
  import type { FloorPlanTone } from "@/lib/iso-scene/floor-plan-image";

  type ViewMode = "tree" | "scene";

  interface SavedView {
    mode: ViewMode;
    tilt: number;
    tone: FloorPlanTone;
    sceneViewport?: ViewportState;
    disabledNodeKeys: string[];
  }

  const savedViews = new Map<string, SavedView>();
</script>

<script lang="ts">
  import CornersOutIcon from "phosphor-svelte/lib/CornersOutIcon";
  import CubeIcon from "phosphor-svelte/lib/CubeIcon";
  import TreeStructureIcon from "phosphor-svelte/lib/TreeStructureIcon";
  import type { Snippet } from "svelte";
  import { SvelteSet } from "svelte/reactivity";
  import type {
    Attributes,
    AuthSession,
    DeviceId,
    Space,
    SpaceGroup,
    SpaceId,
    Switch,
    SwitchPanelId,
  } from "@/api";
  import back from "@/components/back-button";
  import headerTitle from "@/components/header-title";
  import {
    groupColor,
    paletteForTone,
  } from "@/components/iso-scene/paint/palette";
  import { editLinkAttributes } from "@/components/link-attributes-dialog/link-attributes-dialog.svelte";
  import load from "@/components/loading-bar";
  import { tooltip } from "@/components/tooltip";
  import { setDeviceSwitchAttributes } from "@/lib/cache/device-switches";
  import { fetchSpaces } from "@/lib/cache/spaces";
  import { setSwitchUpstreamAttributes } from "@/lib/cache/switch-upstreams";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import {
    attachChain,
    type ChainDetail,
    type ChainLink,
    type ChainNode,
    type ChainPlace,
    chainNodeKey,
    countPlacesBySpace,
    describeNode,
    describePlace,
    detailPlaceKey,
    detailSpaceId,
    floorGroups,
    indexChain,
    type PowerChain,
    spannedSpaces,
    withAttachment,
  } from "@/lib/power-chain";
  import { computeChainLoads } from "@/lib/power-chain/load";
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
  const disabledNodeKeys = new SvelteSet<string>();

  const isNodeEnabled = (node: ChainNode) =>
    !disabledNodeKeys.has(chainNodeKey(node));

  const canToggle = $derived(chain?.direction === "downstream");
  const attachment = $derived(
    chain && canToggle ? attachChain(chain.roots, isNodeEnabled) : undefined,
  );
  const visibleChain = $derived(
    chain && attachment ? withAttachment(chain, attachment) : chain,
  );
  const loads = $derived(
    chain && canToggle
      ? computeChainLoads(chain.roots, chain.supply, isNodeEnabled)
      : undefined,
  );

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

  const inspectLink = (link: ChainLink) => {
    inspectDetail({ kind: "link", link });
  };

  const saveLink = (link: ChainLink, attributes: Attributes) => {
    const { downstream, upstream } = link;
    const loading = load.begin(
      t(
        downstream.kind === "switch"
          ? "loading.updateSwitchUpstream"
          : "loading.updateDeviceUpstream",
      ),
    );
    const request =
      downstream.kind === "switch"
        ? setSwitchUpstreamAttributes(
            session,
            downstream.sw.id,
            upstream.sw.id,
            attributes,
          )
        : setDeviceSwitchAttributes(
            session,
            downstream.device.id,
            upstream.sw.id,
            attributes,
          );
    return request
      .then(() => {
        inspected = undefined;
        reload();
      })
      .finally(() => {
        loading.release();
      });
  };

  const openLink = (link: ChainLink) => {
    editLinkAttributes(link.attributes, (attributes) =>
      saveLink(link, attributes),
    );
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
      disabledNodeKeys.clear();
      for (const key of saved.disabledNodeKeys) {
        disabledNodeKeys.add(key);
      }
    }
  };

  const toggleNode = (node: ChainNode, enabled: boolean) => {
    const key = chainNodeKey(node);
    if (enabled) {
      disabledNodeKeys.delete(key);
    } else {
      disabledNodeKeys.add(key);
    }
  };

  const floors = $derived(
    chain && spaces ? spannedSpaces(chain, spaces) : undefined,
  );
  const placeCounts = $derived(
    visibleChain
      ? countPlacesBySpace(visibleChain)
      : new Map<SpaceId, number>(),
  );
  const floorColors = $derived.by(() => {
    if (mode !== "scene" || !chain || !floors) {
      return undefined;
    }
    const palette = paletteForTone(overview.tone);
    return new Map(
      [...floorGroups(visibleChain ?? chain, floors)].flatMap(([id, group]) => {
        const color = groupColor(palette, group);
        return color ? [[id, color] as const] : [];
      }),
    );
  });

  export const locate = {
    device: (id: DeviceId) => index?.devices.get(id)?.device.space?.id,
    switchPanel: (id: SwitchPanelId) => index?.panels.get(id)?.panel.space?.id,
  };

  export const isChainSwitch = (sw: Switch) => {
    const entry = index?.switches.get(sw.id);
    return (
      entry !== undefined &&
      (!attachment ||
        attachment.nodes.has(chainNodeKey({ kind: "switch", ...entry })))
    );
  };

  let lifecycle: AbortSignal | undefined;
  let latestReload = 0;

  export const reload = () => {
    if (!chain || lifecycle?.aborted) {
      return;
    }
    const request = ++latestReload;
    const loading = load.begin(t("loading.fetchPowerChain"));
    loadChain()
      .then((c) => {
        if (!lifecycle?.aborted && request === latestReload) {
          chain = c;
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
    lifecycle = signal;
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
          disabledNodeKeys: [...disabledNodeKeys],
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
      sinkLinks={chain.sinkLinks}
      {loads}
      attachedItems={attachment?.items}
      isEnabled={isNodeEnabled}
      {canToggle}
      ontogglenode={toggleNode}
      oninspectlink={inspectLink}
      onopenlink={openLink}
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
      chain={visibleChain ?? chain}
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
    <ChainNodeInfo detail={inspected} {spaces} {loads} />
  {/if}
{/snippet}
