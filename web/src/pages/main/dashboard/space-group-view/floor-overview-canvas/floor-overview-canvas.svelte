<script lang="ts">
  import { untrack } from "svelte";
  import type { AuthSession, Space, SpaceGroup } from "@/api";
  import IsoSceneView from "@/components/iso-scene/iso-scene.svelte";
  import { paletteForTone } from "@/components/iso-scene/paint/palette";
  import load from "@/components/loading-bar";
  import { t } from "@/lib/i18n";
  import { createFloorSnapshot } from "@/lib/iso-scene/floor-snapshot";
  import type { IsoPlane, IsoSnapshot } from "@/lib/iso-scene/snapshot";
  import type { FloorOverview } from "@/pages/main/dashboard/space-group-view/floor-overview.svelte";
  import styles from "./floor-overview-canvas.module.css";

  interface Props {
    session: AuthSession;
    spaceGroup: SpaceGroup;
    spaces?: Space[];
    overview: FloorOverview;
  }

  const { session, spaceGroup, spaces, overview }: Props = $props();

  let snapshot = $state.raw<IsoSnapshot>();

  const palette = $derived(paletteForTone(overview.tone));

  const oninspect = (plane: IsoPlane) => {
    overview.inspectedSpaceId = spaces?.find(({ id }) => id === plane.id)?.id;
  };

  $effect(() => {
    if (spaces === undefined) {
      return;
    }

    let cancelled = false;
    const isFirstLoad = untrack(() => snapshot === undefined);
    const loading = isFirstLoad
      ? load.begin(t("loading.renderFloors"))
      : undefined;

    createFloorSnapshot({ session, spaceGroup, spaces, tone: overview.tone })
      .then((next) => {
        if (!cancelled) {
          snapshot = next;
        }
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading?.release();
      });

    return () => {
      cancelled = true;
    };
  });
</script>

{#if snapshot === undefined}
  <div class={styles.placeholder} role="status">
    {t("floorOverview.loading")}
  </div>
{:else}
  <IsoSceneView
    bind:this={overview.view}
    {snapshot}
    tilt={overview.tilt}
    {palette}
    highlightId={overview.inspectedSpaceId}
    {oninspect}
  />
{/if}
