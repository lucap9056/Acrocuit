<script lang="ts">
  import { untrack } from "svelte";
  import type { Attachment } from "svelte/attachments";
  import type { SwitchPanel, Device, Position } from "@/api";
  import switchPanelSymbol from "@/assets/symbols/switch-panel.svg";
  import deviceSymbol from "@/assets/symbols/device.svg";
  import type { Point } from "@/components/iso-scene/camera/projection";
  import type { ViewportState } from "@/components/iso-scene/camera/viewport";
  import { createGestures } from "@/components/iso-scene/interaction/gestures";
  import { layoutGrid } from "@/components/iso-scene/layout/grid";
  import { paintGrid } from "@/components/iso-scene/paint/grid";
  import { GRAY_PALETTE } from "@/components/iso-scene/paint/palette";
  import { t } from "@/lib/i18n";
  import type { FloorPlanImage } from "@/lib/iso-scene/floor-plan-image";
  import { onMount } from "@/lib/lifecycle";
  import type {
    PositionDraft,
    PositionDrafts,
  } from "@/pages/main/dashboard/space-group-view/space-view/position-drafts.svelte";
  import {
    switchPanelKey,
    deviceKey,
    type FocusKey,
    type SpaceFocus,
  } from "@/pages/main/dashboard/space-group-view/space-view/space-focus.svelte";
  import styles from "./floor-plan-canvas.module.css";

  interface Props {
    image: FloorPlanImage;
    switchPanels?: SwitchPanel[];
    devices?: Device[];
    focus: SpaceFocus;
    drafts: PositionDrafts;
    onsavepositions: () => void;
    oneditswitchpanel: (switchPanel: SwitchPanel) => void;
    oneditdevice: (device: Device) => void;
  }

  type Marker = PositionDraft & {
    key: FocusKey;
    symbol: string;
    name: string;
  };

  interface MarkerDrag {
    pointerId: number;
    start: Point;
    origin: Position;
    marker: Marker;
    hasMoved: boolean;
  }

  const {
    image,
    switchPanels = [],
    devices = [],
    focus,
    drafts,
    onsavepositions,
    oneditswitchpanel,
    oneditdevice,
  }: Props = $props();

  const FIT_PADDING_PX = 48;
  const DRAG_THRESHOLD_PX = 4;
  const REVEAL_DURATION_MS = 280;
  const REVEAL_PADDING_PX = 64;

  let container = $state<HTMLDivElement>();
  let canvas = $state<HTMLCanvasElement>();
  let ctx = $state.raw<CanvasRenderingContext2D | null>(null);
  let width = $state(0);
  let height = $state(0);
  let dpr = $state(1);
  let viewportOverride = $state.raw<ViewportState>();
  let isPanning = $state(false);
  let draggingKey = $state<FocusKey>();
  let markerDrag: MarkerDrag | undefined;
  let revealFrame: number | undefined;
  let handledRevealRequest = untrack(() => focus.revealRequest);

  const fitViewport = (): ViewportState => {
    const zoom = Math.max(
      Math.min(
        (width - FIT_PADDING_PX * 2) / image.width,
        (height - FIT_PADDING_PX * 2) / image.height,
      ),
      Number.EPSILON,
    );
    return {
      zoom,
      fitZoom: zoom,
      panX: (width - image.width * zoom) / 2,
      panY: (height - image.height * zoom) / 2,
      tilt: Math.PI / 2,
    };
  };

  const viewport = $derived(viewportOverride ?? fitViewport());

  const markers: Marker[] = $derived([
    ...switchPanels.flatMap((target) =>
      target.position
        ? [
            {
              key: switchPanelKey(target.id),
              kind: "switch_panel" as const,
              target,
              symbol: switchPanelSymbol,
              name: target.name,
              position: target.position,
            },
          ]
        : [],
    ),
    ...devices.flatMap((target) =>
      target.position
        ? [
            {
              key: deviceKey(target.id),
              kind: "device" as const,
              target,
              symbol: deviceSymbol,
              name: target.name,
              position: target.position,
            },
          ]
        : [],
    ),
  ]);

  const toScreen = ({ x, y }: Position): Point => ({
    x: viewport.panX + (image.width / 2 + x) * viewport.zoom,
    y: viewport.panY + (image.height / 2 + y) * viewport.zoom,
  });

  const toLocal = (event: MouseEvent): Point => {
    const rect = canvas?.getBoundingClientRect();
    return {
      x: event.clientX - (rect?.left ?? 0),
      y: event.clientY - (rect?.top ?? 0),
    };
  };

  const displayedPosition = (marker: Marker): Position =>
    drafts.get(marker.key)?.position ?? marker.position;

  const onmarkerpointerdown = (event: PointerEvent, marker: Marker) => {
    if (
      drafts.isSaving ||
      markerDrag ||
      (event.pointerType === "mouse" && event.button !== 0)
    ) {
      return;
    }
    if (event.currentTarget instanceof Element) {
      event.currentTarget.setPointerCapture(event.pointerId);
    }
    markerDrag = {
      pointerId: event.pointerId,
      start: toLocal(event),
      origin: displayedPosition(marker),
      marker,
      hasMoved: false,
    };
  };

  const onmarkerpointermove = (event: PointerEvent) => {
    if (markerDrag?.pointerId !== event.pointerId) {
      return;
    }
    const { start, origin, marker } = markerDrag;
    const current = toLocal(event);
    if (
      !markerDrag.hasMoved &&
      Math.hypot(current.x - start.x, current.y - start.y) < DRAG_THRESHOLD_PX
    ) {
      return;
    }
    markerDrag.hasMoved = true;
    draggingKey = marker.key;
    drafts.set(marker.key, {
      ...marker,
      position: {
        ...origin,
        x: Math.round(origin.x + (current.x - start.x) / viewport.zoom),
        y: Math.round(origin.y + (current.y - start.y) / viewport.zoom),
      },
    });
  };

  const onmarkerpointerend = (event: PointerEvent) => {
    if (markerDrag?.pointerId === event.pointerId) {
      markerDrag = undefined;
      draggingKey = undefined;
    }
  };

  const cancelReveal = () => {
    if (revealFrame !== undefined) {
      cancelAnimationFrame(revealFrame);
      revealFrame = undefined;
    }
  };

  const shiftIntoView = (value: number, size: number) => {
    if (size <= REVEAL_PADDING_PX * 2) {
      return size / 2 - value;
    }
    if (value < REVEAL_PADDING_PX) {
      return REVEAL_PADDING_PX - value;
    }
    if (value > size - REVEAL_PADDING_PX) {
      return size - REVEAL_PADDING_PX - value;
    }
    return 0;
  };

  const reveal = (key: FocusKey) => {
    const marker = markers.find((m) => m.key === key);
    if (!marker) {
      return;
    }
    const from = viewport;
    const point = toScreen(displayedPosition(marker));
    const dx = shiftIntoView(point.x, width);
    const dy = shiftIntoView(point.y, height);
    if (dx === 0 && dy === 0) {
      return;
    }
    const toPanX = from.panX + dx;
    const toPanY = from.panY + dy;

    cancelReveal();
    const startedAt = performance.now();
    const step = (now: number) => {
      const progress = Math.min((now - startedAt) / REVEAL_DURATION_MS, 1);
      const eased = 1 - (1 - progress) ** 3;
      viewportOverride = {
        ...from,
        panX: from.panX + (toPanX - from.panX) * eased,
        panY: from.panY + (toPanY - from.panY) * eased,
      };
      revealFrame = progress < 1 ? requestAnimationFrame(step) : undefined;
    };
    revealFrame = requestAnimationFrame(step);
  };

  $effect(() => {
    const request = focus.revealRequest;
    if (request && request !== handledRevealRequest) {
      handledRevealRequest = request;
      untrack(() => reveal(request.key));
    }
  });

  const onmarkerdblclick = (marker: Marker) => {
    if (marker.kind === "switch_panel") {
      oneditswitchpanel(marker.target);
    } else {
      oneditdevice(marker.target);
    }
  };

  const gestures = createGestures({
    toLocal,
    getViewport: () => viewport,
    onviewportchange: (next) => {
      cancelReveal();
      viewportOverride = next;
    },
    onpanningchange: (panning) => {
      isPanning = panning;
    },
  });

  const paint = (context: CanvasRenderingContext2D, v: ViewportState) => {
    const { zoom, panX, panY } = v;
    context.setTransform(1, 0, 0, 1, 0, 0);
    context.clearRect(0, 0, context.canvas.width, context.canvas.height);
    paintGrid(context, layoutGrid(v), width, height, dpr, GRAY_PALETTE);

    context.setTransform(zoom * dpr, 0, 0, zoom * dpr, panX * dpr, panY * dpr);
    context.drawImage(image.bitmap, 0, 0, image.width, image.height);

    context.lineWidth = 1 / zoom;
    context.strokeStyle = GRAY_PALETTE.lineFaint;
    context.strokeRect(0, 0, image.width, image.height);
  };

  $effect(() => {
    if (canvas) {
      canvas.width = Math.round(width * dpr);
      canvas.height = Math.round(height * dpr);
    }
  });

  $effect(() => {
    if (ctx && width > 0 && height > 0) {
      paint(ctx, viewport);
    }
  });

  const forwardWheel: Attachment<HTMLElement> = (element) => {
    element.addEventListener("wheel", gestures.onwheel, { passive: false });
    return () => {
      element.removeEventListener("wheel", gestures.onwheel);
    };
  };

  onMount((signal) => {
    if (!canvas || !container) {
      return;
    }

    ctx = canvas.getContext("2d");
    canvas.addEventListener("wheel", gestures.onwheel, {
      passive: false,
      signal,
    });

    const observer = new ResizeObserver(([entry]) => {
      width = entry.contentRect.width;
      height = entry.contentRect.height;
      dpr = window.devicePixelRatio || 1;
    });
    observer.observe(container);

    return () => {
      observer.disconnect();
      cancelReveal();
    };
  });
</script>

<div class={styles.floor_plan} bind:this={container}>
  <canvas
    class={[styles.canvas, isPanning && styles.panning]}
    bind:this={canvas}
    onpointerdown={gestures.onpointerdown}
    onpointermove={gestures.onpointermove}
    onpointerup={gestures.onpointerup}
    onpointercancel={gestures.onpointercancel}
  ></canvas>
  <div
    class={styles.markers}
    style:--marker-highlight={GRAY_PALETTE.highlight.line}
    style:--marker-paper={GRAY_PALETTE.paper}
    style:--marker-line={GRAY_PALETTE.line}
    style:--marker-leader={GRAY_PALETTE.lineMuted}
  >
    {#each markers as marker (marker.key)}
      {@const point = toScreen(displayedPosition(marker))}
      <div
        class={[
          styles.marker,
          styles[marker.kind],
          focus.hoveredKey === marker.key && styles.focused,
          draggingKey === marker.key && styles.dragging,
        ]}
        style:transform={`translate(${point.x}px, ${point.y}px)`}
      >
        <span
          class={styles.hit_area}
          role="img"
          aria-label={marker.name}
          onpointerenter={() => focus.enter(marker.key)}
          onpointerleave={() => focus.leave(marker.key)}
          onpointerdown={(event) => onmarkerpointerdown(event, marker)}
          onpointermove={onmarkerpointermove}
          onpointerup={onmarkerpointerend}
          onpointercancel={onmarkerpointerend}
          ondblclick={() => onmarkerdblclick(marker)}
          {@attach forwardWheel}
        >
          <span
            class={styles.symbol}
            style:--symbol={`url("${marker.symbol}")`}
          ></span>
        </span>
        <span class={styles.label}>{marker.name}</span>
      </div>
    {/each}
  </div>
  {#if drafts.size > 0}
    <div class={styles.draft_actions}>
      <button disabled={drafts.isSaving} onclick={() => drafts.clear()}>
        {t("common.cancel")}
      </button>
      <button
        class={styles.save}
        disabled={drafts.isSaving}
        onclick={onsavepositions}
      >
        {t("common.save")}
      </button>
    </div>
  {/if}
</div>
