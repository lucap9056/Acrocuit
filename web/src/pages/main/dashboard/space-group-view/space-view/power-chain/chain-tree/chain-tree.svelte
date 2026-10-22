<script lang="ts" module>
  import type { ViewportState } from "@/components/iso-scene/camera/viewport";

  const savedViewports = new Map<string, ViewportState>();
</script>

<script lang="ts">
  import { untrack } from "svelte";
  import type { Space, SpaceId } from "@/api";
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
    type ChainNode,
    type ChainTreeNode,
    isNodePlaced,
    nodeSpaceId,
  } from "@/lib/power-chain";
  import styles from "./chain-tree.module.css";
  import { layoutTree, NODE_HEIGHT, NODE_WIDTH } from "./tree-layout";

  interface Props {
    chainId: string;
    label: string;
    roots: ChainTreeNode[];
    sink?: ChainNode;
    emptyText: string;
    spaces: Space[];
    inspectedSpaceId?: SpaceId;
    oninspectnode: (node: ChainNode) => void;
    onopennode: (node: ChainNode) => void;
  }

  const {
    chainId,
    label,
    roots,
    sink,
    emptyText,
    spaces,
    inspectedSpaceId,
    oninspectnode,
    onopennode,
  }: Props = $props();

  const FIT_PADDING_PX = 32;
  const MAX_FIT_ZOOM = 1;
  const TEXT_INSET_PX = 12;
  const MAX_TEXT_WIDTH_PX = NODE_WIDTH - TEXT_INSET_PX * 2;

  let container = $state<HTMLDivElement>();
  let width = $state(0);
  let height = $state(0);
  let viewportOverride = $state.raw<ViewportState | undefined>(
    untrack(() => savedViewports.get(chainId)),
  );
  let isPanning = $state(false);

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

  const fitText = (text: string, font: string) => {
    if (measureTextWidth(text, font) <= MAX_TEXT_WIDTH_PX) {
      return text;
    }
    let end = text.length;
    while (
      end > 0 &&
      measureTextWidth(`${text.slice(0, end)}…`, font) > MAX_TEXT_WIDTH_PX
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

  const inspect = (node: ChainNode) => {
    oninspectnode(node);
  };

  const open = ({ node, isFocus }: ChainTreeNode) => {
    if (!isFocus) {
      onopennode(node);
    }
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
        {#each layout.edges as path, index (index)}
          <path class={styles.edge} d={path} />
        {/each}
        {#each layout.sinkEdges as path, index (index)}
          <path class={[styles.edge, styles.feed]} d={path} />
        {/each}
        {#each layout.nodes as { item, x, y } (item.key)}
          {@const node = item.node}
          <g
            class={[
              styles.node,
              node.kind === "device" && styles.leaf,
              item.isFocus && styles.focus,
              nodeSpaceId(node) === inspectedSpaceId && styles.inspected,
              !isNodePlaced(node) && styles.unplaced,
            ]}
            transform={`translate(${x} ${y})`}
            role="treeitem"
            aria-selected="false"
            aria-disabled={item.isFocus}
            tabindex={item.isFocus ? -1 : 0}
            onpointerenter={() => inspect(node)}
            onfocus={() => inspect(node)}
            onclick={() => open(item)}
            onkeydown={(event) => onnodekeydown(event, item)}
          >
            <rect
              width={NODE_WIDTH}
              height={NODE_HEIGHT}
              rx={node.kind === "device" ? NODE_HEIGHT / 2 : 6}
            />
            <text class={styles.name} x={TEXT_INSET_PX} y="19">
              {fitText(nodeTitle(node), LABEL_FONT)}
            </text>
            <text class={styles.caption} x={TEXT_INSET_PX} y="36">
              {fitText(nodeCaption(node), CAPTION_FONT)}
            </text>
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
              rx={NODE_HEIGHT / 2}
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
