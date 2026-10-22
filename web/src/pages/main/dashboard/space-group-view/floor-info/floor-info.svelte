<script lang="ts">
  import type { AuthSession, Space } from "@/api";
  import { fetchSwitchPanels } from "@/lib/cache/switch-panels";
  import { fetchDevices } from "@/lib/cache/devices";
  import { fetchFloorPlanImage } from "@/lib/cache/space-images";
  import { t } from "@/lib/i18n";
  import type { FloorOverview } from "@/pages/main/dashboard/space-group-view/floor-overview.svelte";
  import styles from "./floor-info.module.css";

  interface Props {
    session: AuthSession;
    spaces: Space[];
    overview: FloorOverview;
  }

  interface SpaceStats {
    switchPanels: number;
    devices: number;
    hasFloorPlan: boolean;
  }

  const { session, spaces, overview }: Props = $props();

  let stats = $state<SpaceStats>();

  const index = $derived(
    spaces.findIndex(({ id }) => id === overview.inspectedSpaceId),
  );
  const space = $derived<Space | undefined>(spaces[index]);

  $effect(() => {
    const target = space;
    stats = undefined;
    if (target === undefined) {
      return;
    }

    let cancelled = false;
    Promise.all([
      fetchSwitchPanels(session, target.id),
      fetchDevices(session, target.id),
      fetchFloorPlanImage(session, target, overview.tone),
    ])
      .then(([switchPanels, devices, floorPlan]) => {
        if (!cancelled) {
          stats = {
            switchPanels: switchPanels.length,
            devices: devices.length,
            hasFloorPlan: floorPlan !== undefined,
          };
        }
      })
      .catch((err) => {
        console.error(err);
      });

    return () => {
      cancelled = true;
    };
  });
</script>

<div class={styles.floor_info}>
  {#if space}
    <h2 class={styles.name}>{space.name}</h2>
    <dl class={styles.details}>
      <dt>{t("floorInfo.floor")}</dt>
      <dd>{index + 1} / {spaces.length}</dd>
      <dt>{t("switchPanels.title")}</dt>
      <dd>{stats?.switchPanels ?? "—"}</dd>
      <dt>{t("devices.title")}</dt>
      <dd>{stats?.devices ?? "—"}</dd>
      <dt>{t("floorPlan.title")}</dt>
      <dd>
        {stats === undefined
          ? "—"
          : stats.hasFloorPlan
            ? t("floorPlan.uploaded")
            : t("common.none")}
      </dd>
    </dl>
  {:else}
    <p class={styles.message}>
      {t("floorInfo.hint")}
    </p>
  {/if}
</div>
