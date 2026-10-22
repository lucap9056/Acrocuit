<script lang="ts">
  import CornersOutIcon from "phosphor-svelte/lib/CornersOutIcon";
  import MagnifyingGlassMinusIcon from "phosphor-svelte/lib/MagnifyingGlassMinusIcon";
  import MagnifyingGlassPlusIcon from "phosphor-svelte/lib/MagnifyingGlassPlusIcon";
  import PaletteIcon from "phosphor-svelte/lib/PaletteIcon";
  import {
    MAX_TILT_DEG,
    MIN_TILT_DEG,
  } from "@/components/iso-scene/camera/viewport";
  import { tooltip } from "@/components/tooltip";
  import { t } from "@/lib/i18n";
  import type { FloorOverview } from "@/pages/main/dashboard/space-group-view/floor-overview.svelte";
  import styles from "./floor-overview-toolbar.module.css";

  interface Props {
    overview: FloorOverview;
  }

  const { overview }: Props = $props();

  const isBlueprint = $derived(overview.tone === "blueprint");

  const toggleTone = () => {
    overview.tone = isBlueprint ? "original" : "blueprint";
  };
</script>

<div class={styles.toolbar}>
  <label class={styles.tilt_control}>
    <span>{t("floorOverview.tilt")}</span>
    <input
      type="range"
      min={MIN_TILT_DEG}
      max={MAX_TILT_DEG}
      step="1"
      bind:value={overview.tilt}
    >
    <span class={styles.tilt_value}>{overview.tilt}°</span>
  </label>
  <button
    class={["icon_button", isBlueprint && styles.active]}
    aria-label={t("floorOverview.blueprintColors")}
    aria-pressed={isBlueprint}
    onclick={toggleTone}
    {@attach tooltip(t("floorOverview.blueprintColors"))}
  >
    <PaletteIcon />
  </button>
  <div class={styles.zoom_controls}>
    <button
      class="icon_button"
      aria-label={t("floorOverview.zoomOut")}
      onclick={() => overview.view?.zoomOut()}
      {@attach tooltip(t("floorOverview.zoomOut"))}
    >
      <MagnifyingGlassMinusIcon />
    </button>
    <span class={styles.zoom_level}
      >{overview.zoomLevel === undefined ? "—" : `${overview.zoomLevel}%`}</span
    >
    <button
      class="icon_button"
      aria-label={t("floorOverview.zoomIn")}
      onclick={() => overview.view?.zoomIn()}
      {@attach tooltip(t("floorOverview.zoomIn"))}
    >
      <MagnifyingGlassPlusIcon />
    </button>
    <button
      class="icon_button"
      aria-label={t("floorOverview.fitToView")}
      onclick={() => overview.view?.fit()}
      {@attach tooltip(t("floorOverview.fitToView"))}
    >
      <CornersOutIcon />
    </button>
  </div>
</div>
