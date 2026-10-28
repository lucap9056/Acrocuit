<script lang="ts">
  import type { AuthSession, Device, Space, SwitchPanel } from "@/api";
  import load from "@/components/loading-bar";
  import { fetchFloorPlanImage } from "@/lib/cache/space-images";
  import { t } from "@/lib/i18n";
  import type { FloorPlanImage } from "@/lib/iso-scene/floor-plan-image";
  import type { PositionDrafts } from "@/pages/main/dashboard/space-group-view/space-view/position-drafts.svelte";
  import type { SpaceFocus } from "@/pages/main/dashboard/space-group-view/space-view/space-focus.svelte";
  import styles from "./floor-plan-canvas.module.css";
  import FloorPlanView from "./floor-plan-view.svelte";

  interface Props {
    session: AuthSession;
    space: Space;
    switchPanels?: SwitchPanel[];
    devices?: Device[];
    focus: SpaceFocus;
    drafts: PositionDrafts;
    onsavepositions: () => void;
    oneditswitchpanel: (switchPanel: SwitchPanel) => void;
    oneditdevice: (device: Device) => void;
    editable: boolean;
  }

  interface LoadedFloorPlan {
    version: number;
    image?: FloorPlanImage;
  }

  const {
    session,
    space,
    switchPanels,
    devices,
    focus,
    drafts,
    onsavepositions,
    oneditswitchpanel,
    oneditdevice,
    editable,
  }: Props = $props();

  let floorPlan = $state.raw<LoadedFloorPlan>();

  const current = $derived(
    floorPlan?.version === space.background_image_updated_at
      ? floorPlan
      : undefined,
  );

  $effect(() => {
    const target = space;
    const targetVersion = target.background_image_updated_at;
    if (!targetVersion) {
      floorPlan = { version: targetVersion };
      return;
    }

    let cancelled = false;
    const loading = load.begin(t("loading.fetchFloorPlan"));
    fetchFloorPlanImage(session, target, "original")
      .then((image) => {
        if (!cancelled) {
          floorPlan = { version: targetVersion, image };
        }
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
      });

    return () => {
      cancelled = true;
    };
  });
</script>

{#if current === undefined}
  <div class={styles.placeholder} role="status">{t("floorPlan.loading")}</div>
{:else if current.image === undefined}
  <div class={styles.placeholder}>{t("floorPlan.empty")}</div>
{:else}
  {#key current.image}
    <FloorPlanView
      image={current.image}
      {switchPanels}
      {devices}
      {focus}
      {drafts}
      {editable}
      {onsavepositions}
      {oneditswitchpanel}
      {oneditdevice}
    />
  {/key}
{/if}
