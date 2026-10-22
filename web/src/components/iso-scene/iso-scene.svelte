<script lang="ts">
  import { untrack } from "svelte";
  import type {
    IsoMarker,
    IsoPlane,
    IsoSnapshot,
  } from "@/lib/iso-scene/snapshot";
  import { onMount } from "@/lib/lifecycle";
  import { revealPoints, zoomAt } from "./camera/navigate";
  import type { Point } from "./camera/projection";
  import { resolveViewport } from "./camera/resolve";
  import { type ViewportState, zoomPercent } from "./camera/viewport";
  import { createGestures } from "./interaction/gestures";
  import HitLayer from "./interaction/hit-layer.svelte";
  import styles from "./iso-scene.module.css";
  import { layoutScene } from "./layout/scene-layout";
  import { createIsoPainter, type IsoPainter } from "./paint/painter";
  import { BLUEPRINT_PALETTE, type ScenePalette } from "./paint/palette";

  interface Props {
    snapshot: IsoSnapshot;
    tilt: number;
    palette?: ScenePalette;
    highlightId?: number;
    highlightMarkerId?: string;
    initialViewport?: ViewportState;
    oninspect?: (plane: IsoPlane) => void;
    oninspectmarker?: (marker: IsoMarker) => void;
    onopenmarker?: (marker: IsoMarker) => void;
    onviewportchange?: (viewport: ViewportState | undefined) => void;
  }

  const {
    snapshot,
    tilt,
    palette = BLUEPRINT_PALETTE,
    highlightId,
    highlightMarkerId,
    initialViewport,
    oninspect,
    oninspectmarker,
    onopenmarker,
    onviewportchange,
  }: Props = $props();

  const BUTTON_ZOOM_FACTOR = 1.25;
  const REVEAL_PADDING_PX = 32;
  const REVEAL_DURATION_MS = 280;

  let container = $state<HTMLDivElement>();
  let canvas = $state<HTMLCanvasElement>();
  let painter = $state.raw<IsoPainter>();
  let width = $state(0);
  let height = $state(0);
  let dpr = $state(1);
  let viewportOverride = $state.raw<ViewportState | undefined>(
    untrack(() => initialViewport),
  );
  let revealFrame: number | undefined;
  let isPanning = $state(false);

  const viewport = $derived(
    resolveViewport(snapshot, tilt, width, height, viewportOverride),
  );
  const layout = $derived(layoutScene(snapshot, viewport, width, height));
  const zoomLevel = $derived(zoomPercent(viewport));

  const toLocal = (event: MouseEvent): Point => {
    const rect = canvas?.getBoundingClientRect();
    return {
      x: event.clientX - (rect?.left ?? 0),
      y: event.clientY - (rect?.top ?? 0),
    };
  };

  const inspect = (plane: IsoPlane) => {
    if (plane.id !== highlightId) {
      oninspect?.(plane);
    }
  };

  const inspectMarker = (marker: IsoMarker) => {
    if (marker.id !== highlightMarkerId) {
      oninspectmarker?.(marker);
    }
  };

  const inspectElementAt = (clientX: number, clientY: number) => {
    const target = document
      .elementFromPoint(clientX, clientY)
      ?.closest<HTMLElement>("[data-plane-index], [data-marker-index]");
    const marker = layout.markers[Number(target?.dataset.markerIndex)]?.marker;
    const plane = layout.planes[Number(target?.dataset.planeIndex)]?.plane;
    if (marker) {
      inspectMarker(marker);
    } else if (plane) {
      inspect(plane);
    }
  };

  const onwheel = (event: WheelEvent) => {
    gestures.onwheel(event);
    requestAnimationFrame(() => inspectElementAt(event.clientX, event.clientY));
  };

  const cancelReveal = () => {
    if (revealFrame !== undefined) {
      cancelAnimationFrame(revealFrame);
      revealFrame = undefined;
    }
  };

  const setViewport = (next: ViewportState | undefined) => {
    cancelReveal();
    viewportOverride = next;
  };

  const gestures = createGestures({
    toLocal,
    getViewport: () => viewport,
    onviewportchange: setViewport,
    onpanningchange: (panning) => {
      isPanning = panning;
    },
  });

  const zoomFromCenter = (factor: number) => {
    setViewport(
      zoomAt(viewport, factor, {
        x: width / 2,
        y: height / 2,
      }),
    );
  };

  export const zoomIn = () => zoomFromCenter(BUTTON_ZOOM_FACTOR);
  export const zoomOut = () => zoomFromCenter(1 / BUTTON_ZOOM_FACTOR);
  export const fit = () => setViewport(undefined);

  export const reveal = (id: number) => {
    const target = layout.planes.find(({ plane }) => plane.id === id);
    if (!target) {
      return;
    }
    const from = viewport;
    const to = revealPoints(
      from,
      target.corners,
      width,
      height,
      REVEAL_PADDING_PX,
    );
    if (to === from) {
      return;
    }

    cancelReveal();
    const startedAt = performance.now();
    const step = (now: number) => {
      const progress = Math.min((now - startedAt) / REVEAL_DURATION_MS, 1);
      const eased = 1 - (1 - progress) ** 3;
      viewportOverride = {
        ...from,
        panX: from.panX + (to.panX - from.panX) * eased,
        panY: from.panY + (to.panY - from.panY) * eased,
      };
      revealFrame = progress < 1 ? requestAnimationFrame(step) : undefined;
    };
    revealFrame = requestAnimationFrame(step);
  };

  export const getZoomLevel = () => zoomLevel;

  $effect(() => {
    painter?.paint(layout, { dpr, highlightId, highlightMarkerId, palette });
  });

  $effect(() => {
    onviewportchange?.(viewportOverride);
  });

  onMount((signal) => {
    if (!canvas || !container) {
      return;
    }

    const isoPainter = createIsoPainter(canvas);
    painter = isoPainter;

    container.addEventListener("wheel", onwheel, {
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
      cancelReveal();
      observer.disconnect();
      isoPainter.dispose();
      painter = undefined;
    };
  });
</script>

<div
  class={[styles.iso_scene, isPanning && styles.panning]}
  role="group"
  aria-label={snapshot.title}
  bind:this={container}
  onpointerdown={gestures.onpointerdown}
  onpointermove={gestures.onpointermove}
  onpointerup={gestures.onpointerup}
  onpointercancel={gestures.onpointercancel}
>
  <canvas class={styles.canvas} bind:this={canvas}></canvas>
  <HitLayer
    {layout}
    {highlightMarkerId}
    oninspect={inspect}
    oninspectmarker={inspectMarker}
    onopenmarker={(marker) => onopenmarker?.(marker)}
  />
</div>
