<script lang="ts">
  import type { SceneLayout } from "@/components/iso-scene/layout/scene-layout";
  import type { IsoMarker, IsoPlane } from "@/lib/iso-scene/snapshot";
  import styles from "./hit-layer.module.css";

  interface Props {
    layout: SceneLayout;
    highlightMarkerId?: string;
    oninspect: (plane: IsoPlane) => void;
    oninspectmarker: (marker: IsoMarker) => void;
    onopenmarker: (marker: IsoMarker) => void;
  }

  const {
    layout,
    highlightMarkerId,
    oninspect,
    oninspectmarker,
    onopenmarker,
  }: Props = $props();

  const markerOrder = $derived(
    layout.markers
      .map((_, index) => index)
      .filter((index) => layout.markers[index].marker.id !== undefined)
      .sort(
        (a, b) =>
          layout.markers[a].marker.point.z - layout.markers[b].marker.point.z,
      ),
  );

  const onpointerenter = (event: PointerEvent, plane: IsoPlane) => {
    if (event.pointerType !== "touch") {
      oninspect(plane);
    }
  };

  const onmarkerpointerenter = (event: PointerEvent, marker: IsoMarker) => {
    if (event.pointerType !== "touch") {
      oninspectmarker(marker);
    }
  };

  const onmarkerclick = (event: MouseEvent, marker: IsoMarker) => {
    const isTouch =
      event instanceof PointerEvent && event.pointerType === "touch";
    if (isTouch && marker.id !== highlightMarkerId) {
      oninspectmarker(marker);
    } else {
      onopenmarker(marker);
    }
  };
</script>

<div class={styles.hit_layer}>
  {#each layout.planes as { plane, matrix }, index (plane.id)}
    {@const [x, y, , width, depth] = plane.rect}
    <button
      class={styles.plane}
      style:width={`${width}px`}
      style:height={`${depth}px`}
      style:transform={`matrix(${matrix.join(",")}) translate(${x}px, ${y}px)`}
      aria-label={plane.text}
      data-plane-index={index}
      onpointerenter={(event) => onpointerenter(event, plane)}
      onfocus={() => oninspect(plane)}
      onclick={() => oninspect(plane)}
    ></button>
  {/each}
  {#each layout.planes as { plane, label }, index (plane.id)}
    {#if label}
      <button
        class={styles.label}
        style:left={`${label.left}px`}
        style:top={`${label.top}px`}
        style:width={`${label.anchor.x - label.left}px`}
        style:height={`${label.bottom - label.top}px`}
        tabindex="-1"
        aria-hidden="true"
        data-plane-index={index}
        onpointerenter={(event) => onpointerenter(event, plane)}
        onclick={() => oninspect(plane)}
      ></button>
    {/if}
  {/each}
  {#each markerOrder as index (layout.markers[index].marker.id)}
    {@const { marker, point } = layout.markers[index]}
    <button
      class={styles.marker}
      style:transform={`translate(${point.x}px, ${point.y}px)`}
      aria-label={marker.text}
      data-marker-index={index}
      onpointerenter={(event) => onmarkerpointerenter(event, marker)}
      onfocus={() => oninspectmarker(marker)}
      onclick={(event) => onmarkerclick(event, marker)}
    ></button>
  {/each}
</div>
