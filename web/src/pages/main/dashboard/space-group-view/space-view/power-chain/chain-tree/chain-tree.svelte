<script lang="ts" module>
  import type { ViewportState } from "@/components/iso-scene/camera/viewport";

  const savedViewports = new Map<string, ViewportState>();
</script>

<script lang="ts">
  import { untrack } from "svelte";
  import type { Attributes, Space, SpaceId, SwitchId } from "@/api";
  import type { Point } from "@/components/iso-scene/camera/projection";
  import { createGestures } from "@/components/iso-scene/interaction/gestures";
  import { measureTextWidth } from "@/components/iso-scene/layout/text-metrics";
  import {
    CAPTION_FONT,
    LABEL_FONT,
  } from "@/components/iso-scene/layout/typography";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import {
    type ChainLink,
    type ChainNode,
    type ChainTreeNode,
    chainLink,
    isNodePlaced,
    nodeSpaceId,
  } from "@/lib/power-chain";
  import {
    type ChainLoads,
    formatSwitchLoad,
    isLinkOverloaded,
    isOverloaded,
  } from "@/lib/power-chain/load";
  import styles from "./chain-tree.module.css";
  import {
    layoutTree,
    NODE_HEIGHT,
    NODE_WIDTH,
    type TreeEdge,
  } from "./tree-layout";

  interface Props {
    chainId: string;
    label: string;
    roots: ChainTreeNode[];
    sink?: ChainNode;
    sinkLinks?: Map<SwitchId, Attributes | null>;
    loads?: ChainLoads;
    attachedItems?: ReadonlySet<string>;
    isEnabled: (node: ChainNode) => boolean;
    canToggle: boolean;
    emptyText: string;
    spaces: Space[];
    inspectedSpaceId?: SpaceId;
    oninspectnode: (node: ChainNode) => void;
    onopennode: (node: ChainNode) => void;
    ontogglenode: (node: ChainNode, enabled: boolean) => void;
    oninspectlink: (link: ChainLink) => void;
    onopenlink: (link: ChainLink) => void;
  }

  const {
    chainId,
    label,
    roots,
    sink,
    sinkLinks,
    loads,
    attachedItems,
    isEnabled,
    canToggle,
    emptyText,
    spaces,
    inspectedSpaceId,
    oninspectnode,
    onopennode,
    ontogglenode,
    oninspectlink,
    onopenlink,
  }: Props = $props();

  const FIT_PADDING_PX = 32;
  const MAX_FIT_ZOOM = 1;
  const TEXT_INSET_PX = 12;
  const MAX_TEXT_WIDTH_PX = NODE_WIDTH - TEXT_INSET_PX * 2;
  const CHECKBOX_SIZE_PX = 16;
  const CHECKBOX_RIGHT_PX = 12;
  const CHECKBOX_OVERHANG_PX = 6;
  const CHECKBOX_RING_PX = 2;
  const BADGE_GAP_PX = 8;
  const DEVICE_RADIUS_PX = NODE_HEIGHT / 2;

  let container = $state<HTMLDivElement>();
  let width = $state(0);
  let height = $state(0);
  let viewportOverride = $state.raw<ViewportState | undefined>(
    untrack(() => savedViewports.get(chainId)),
  );
  let isPanning = $state(false);
  let activeLinkKey = $state<string>();

  const layout = $derived(layoutTree(roots, sink !== undefined));
  const spaceNames = $derived(
    new Map(spaces.map(({ id, name }) => [id, name])),
  );

  const fitViewport = (): ViewportState => {
    const zoom = Math.max(
      Math.min(
        (width - FIT_PADDING_PX * 2) / layout.width,
        (height - FIT_PADDING_PX * 2) / layout.height,
        MAX_FIT_ZOOM,
      ),
      Number.EPSILON,
    );
    return {
      zoom,
      fitZoom: zoom,
      panX: (width - layout.width * zoom) / 2,
      panY: Math.max((height - layout.height * zoom) / 2, FIT_PADDING_PX),
      tilt: Math.PI / 2,
    };
  };

  const viewport = $derived(viewportOverride ?? fitViewport());

  export const fit = () => {
    viewportOverride = undefined;
  };

  const fitText = (
    text: string,
    font: string,
    maxWidth = MAX_TEXT_WIDTH_PX,
  ) => {
    if (measureTextWidth(text, font) <= maxWidth) {
      return text;
    }
    let end = text.length;
    while (
      end > 0 &&
      measureTextWidth(`${text.slice(0, end)}…`, font) > maxWidth
    ) {
      end--;
    }
    return `${text.slice(0, end)}…`;
  };

  const nodeTitle = (node: ChainNode) =>
    node.kind === "switch" ? node.sw.name : node.device.name;

  const nodeCaption = (node: ChainNode) => {
    const spaceId = nodeSpaceId(node);
    const floor = spaceId === undefined ? undefined : spaceNames.get(spaceId);
    const panel = node.kind === "switch" ? node.panel.name : undefined;
    const unplaced = isNodePlaced(node) ? undefined : t("powerChain.unplaced");
    return [panel, floor, unplaced].filter(Boolean).join(" · ");
  };

  const switchLoad = (node: ChainNode) =>
    node.kind === "switch" ? loads?.switches.get(node.sw.id) : undefined;

  const isNodeOverloaded = (item: ChainTreeNode) => {
    const load = switchLoad(item.node);
    return load !== undefined && isOverloaded(load);
  };

  const edgeLink = ({ parent, child }: TreeEdge) => {
    if (child) {
      return chainLink(parent.node, child.node, child.link);
    }
    if (sink && parent.node.kind === "switch") {
      return chainLink(
        parent.node,
        sink,
        sinkLinks?.get(parent.node.sw.id) ?? null,
      );
    }
    return undefined;
  };

  const isLinkLoadOverloaded = (link: ChainLink | undefined) => {
    const linkLoad = link && loads?.links.get(link.key);
    return linkLoad !== undefined && isLinkOverloaded(linkLoad);
  };

  const inspectLink = (link: ChainLink) => {
    activeLinkKey = link.key;
    oninspectlink(link);
  };

  const leaveLink = () => {
    activeLinkKey = undefined;
  };

  const onlinkkeydown = (event: KeyboardEvent, link: ChainLink) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      onopenlink(link);
    }
  };

  const isAttached = (key: string) => attachedItems?.has(key) ?? true;

  const hasToggle = (item: ChainTreeNode) => canToggle && !item.isFocus;

  const titleWidth = (loadText: string | undefined) =>
    loadText === undefined
      ? MAX_TEXT_WIDTH_PX
      : MAX_TEXT_WIDTH_PX -
        measureTextWidth(loadText, CAPTION_FONT) -
        BADGE_GAP_PX;

  const captionWidth = (item: ChainTreeNode) =>
    hasToggle(item)
      ? NODE_WIDTH -
        CHECKBOX_RIGHT_PX -
        CHECKBOX_SIZE_PX -
        BADGE_GAP_PX -
        TEXT_INSET_PX
      : MAX_TEXT_WIDTH_PX;

  const stopPropagation = (event: Event) => {
    event.stopPropagation();
  };

  const inspect = (node: ChainNode) => {
    oninspectnode(node);
  };

  const open = ({ node }: ChainTreeNode) => {
    onopennode(node);
  };

  const onnodekeydown = (event: KeyboardEvent, item: ChainTreeNode) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      open(item);
    }
  };

  const toLocal = (event: MouseEvent): Point => {
    const rect = container?.getBoundingClientRect();
    return {
      x: event.clientX - (rect?.left ?? 0),
      y: event.clientY - (rect?.top ?? 0),
    };
  };

  const gestures = createGestures({
    toLocal,
    getViewport: () => viewport,
    onviewportchange: (next) => {
      viewportOverride = next;
    },
    onpanningchange: (panning) => {
      isPanning = panning;
    },
  });

  onMount((signal) => {
    if (!container) {
      return;
    }

    container.addEventListener("wheel", gestures.onwheel, {
      passive: false,
      signal,
    });

    const observer = new ResizeObserver(([entry]) => {
      width = entry.contentRect.width;
      height = entry.contentRect.height;
    });
    observer.observe(container);

    return () => {
      observer.disconnect();
      if (viewportOverride) {
        savedViewports.set(chainId, viewportOverride);
      } else {
        savedViewports.delete(chainId);
      }
    };
  });
</script>

<div
  class={[styles.tree, isPanning && styles.panning]}
  role="group"
  aria-label={t("powerChain.tree")}
  bind:this={container}
  onpointerdown={gestures.onpointerdown}
  onpointermove={gestures.onpointermove}
  onpointerup={gestures.onpointerup}
  onpointercancel={gestures.onpointercancel}
>
  {#if roots.length === 0}
    <p class={styles.message}>{emptyText}</p>
  {:else}
    <svg class={styles.canvas} role="tree" aria-label={label}>
      <g
        transform={`translate(${viewport.panX} ${viewport.panY}) scale(${viewport.zoom})`}
      >
        {#each [...layout.edges, ...layout.sinkEdges] as edge, index (index)}
          {@const link = edgeLink(edge)}
          <path
            class={[
              styles.edge,
              !edge.child && styles.feed,
              edge.child && !isAttached(edge.child.key) && styles.detached,
              isLinkLoadOverloaded(link) && styles.overloaded,
              link !== undefined && link.key === activeLinkKey && styles.active,
            ]}
            d={edge.path}
          />
        {/each}
        {#each [...layout.edges, ...layout.sinkEdges] as edge, index (index)}
          {@const link = edgeLink(edge)}
          {#if link}
            <path
              class={styles.hit}
              d={edge.hitPath}
              role="button"
              tabindex="0"
              aria-label={t("powerChain.linkLabel", {
                from: link.upstream.sw.name,
                to: nodeTitle(link.downstream),
              })}
              onpointerenter={() => inspectLink(link)}
              onpointerleave={leaveLink}
              onfocus={() => inspectLink(link)}
              onblur={leaveLink}
              onclick={() => onopenlink(link)}
              onkeydown={(event) => onlinkkeydown(event, link)}
            />
          {/if}
        {/each}
        {#each layout.nodes as { item, x, y } (item.key)}
          {@const node = item.node}
          {@const load = switchLoad(node)}
          {@const loadText = load && formatSwitchLoad(load)}
          <g
            class={[
              styles.node,
              node.kind === "device" && styles.leaf,
              !isAttached(item.key) && styles.disabled,
              isNodeOverloaded(item) && styles.overloaded,
              item.isFocus && styles.focus,
              nodeSpaceId(node) === inspectedSpaceId && styles.inspected,
              !isNodePlaced(node) && styles.unplaced,
            ]}
            transform={`translate(${x} ${y})`}
            role="treeitem"
            aria-selected="false"
            tabindex="0"
            onpointerenter={() => inspect(node)}
            onfocus={() => inspect(node)}
            onclick={() => open(item)}
            onkeydown={(event) => onnodekeydown(event, item)}
          >
            <rect
              width={NODE_WIDTH}
              height={NODE_HEIGHT}
              rx={node.kind === "device" ? DEVICE_RADIUS_PX : 0}
            />
            <text class={styles.name} x={TEXT_INSET_PX} y="19">
              {fitText(nodeTitle(node), LABEL_FONT, titleWidth(loadText))}
            </text>
            <text class={styles.caption} x={TEXT_INSET_PX} y="36">
              {fitText(nodeCaption(node), CAPTION_FONT, captionWidth(item))}
            </text>
            {#if loadText}
              <text
                class={styles.load}
                x={NODE_WIDTH - TEXT_INSET_PX}
                y="19"
                text-anchor="end"
              >
                {loadText}
              </text>
            {/if}
            {#if hasToggle(item)}
              <foreignObject
                x={NODE_WIDTH -
                  CHECKBOX_RIGHT_PX -
                  CHECKBOX_SIZE_PX -
                  CHECKBOX_RING_PX}
                y={NODE_HEIGHT -
                  CHECKBOX_SIZE_PX +
                  CHECKBOX_OVERHANG_PX -
                  CHECKBOX_RING_PX}
                width={CHECKBOX_SIZE_PX + CHECKBOX_RING_PX * 2}
                height={CHECKBOX_SIZE_PX + CHECKBOX_RING_PX * 2}
              >
                <input
                  class={styles.toggle}
                  type="checkbox"
                  aria-label={t("powerChain.enableNode")}
                  checked={isEnabled(node)}
                  onpointerdown={stopPropagation}
                  onclick={stopPropagation}
                  onkeydown={stopPropagation}
                  onchange={(event) =>
                    ontogglenode(node, event.currentTarget.checked)}
                >
              </foreignObject>
            {/if}
          </g>
        {/each}
        {#if layout.sink && sink}
          <g
            class={[styles.node, styles.sink]}
            transform={`translate(${layout.sink.x} ${layout.sink.y})`}
          >
            <rect
              width={NODE_WIDTH}
              height={NODE_HEIGHT}
              rx={DEVICE_RADIUS_PX}
            />
            <text
              class={styles.name}
              x={NODE_WIDTH / 2}
              y={NODE_HEIGHT / 2}
              text-anchor="middle"
              dominant-baseline="central"
            >
              {fitText(nodeTitle(sink), LABEL_FONT)}
            </text>
          </g>
        {/if}
      </g>
    </svg>
  {/if}
</div>
