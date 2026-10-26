<script lang="ts">
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import UploadSimpleIcon from "phosphor-svelte/lib/UploadSimpleIcon";
  import type { Device, Space, SwitchPanel } from "@/api";
  import { t } from "@/lib/i18n";
  import styles from "./space-info.module.css";

  interface Props {
    space: Space;
    switchPanels?: SwitchPanel[];
    devices?: Device[];
    isUpdatingFloorPlan: boolean;
    onuploadfloorplan: (image: File) => void;
    onremovefloorplan: () => void;
  }

  const {
    space,
    switchPanels,
    devices,
    isUpdatingFloorPlan,
    onuploadfloorplan,
    onremovefloorplan,
  }: Props = $props();

  const FLOOR_PLAN_TYPES = "image/png,image/jpeg,image/webp,image/gif";

  let fileInput = $state<HTMLInputElement>();

  const floorPlanUpdatedAt = $derived(
    space.background_image_updated_at
      ? new Date(space.background_image_updated_at).toLocaleString()
      : undefined,
  );

  const onfilechange = (event: Event & { currentTarget: HTMLInputElement }) => {
    const image = event.currentTarget.files?.[0];
    event.currentTarget.value = "";
    if (image) {
      onuploadfloorplan(image);
    }
  };
</script>

<div class={styles.space_info}>
  <h2 class={styles.name}>{space.name}</h2>
  <dl class={styles.details}>
    <dt>{t("switchPanels.title")}</dt>
    <dd>{switchPanels?.length ?? "—"}</dd>
    <dt>{t("devices.title")}</dt>
    <dd>{devices?.length ?? "—"}</dd>
    <dt>{t("floorPlan.title")}</dt>
    <dd>{floorPlanUpdatedAt ?? t("common.none")}</dd>
  </dl>
  <div class={styles.actions}>
    {#if floorPlanUpdatedAt === undefined}
      <button
        class={styles.action}
        disabled={isUpdatingFloorPlan}
        onclick={() => fileInput?.click()}
      >
        <UploadSimpleIcon />
        {t("floorPlan.upload")}
      </button>
    {:else}
      <button
        class={styles.action}
        disabled={isUpdatingFloorPlan}
        onclick={onremovefloorplan}
      >
        <TrashIcon />
        {t("floorPlan.remove")}
      </button>
    {/if}
    <input
      class={styles.file_input}
      type="file"
      accept={FLOOR_PLAN_TYPES}
      bind:this={fileInput}
      onchange={onfilechange}
    >
  </div>
</div>
