<script lang="ts">
  import type { SceneLayout } from "@/components/iso-scene/layout/scene-layout";
  import type { IsoPlane } from "@/lib/iso-scene/snapshot";
  import styles from "./hit-layer.module.css";

  interface Props {
    layout: SceneLayout;
    oninspect: (plane: IsoPlane) => void;
  }

  const { layout, oninspect }: Props = $props();

  const onpointerenter = (event: PointerEvent, plane: IsoPlane) => {
    if (event.pointerType !== "touch") {
      oninspect(plane);
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
</div>
