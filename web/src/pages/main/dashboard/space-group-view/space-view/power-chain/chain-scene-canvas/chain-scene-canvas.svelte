<script lang="ts">
  import { untrack } from "svelte";
  import type { AuthSession, Space, SpaceGroup, SpaceId } from "@/api";
  import type { ViewportState } from "@/components/iso-scene/camera/viewport";
  import IsoSceneView from "@/components/iso-scene/iso-scene.svelte";
  import { paletteForTone } from "@/components/iso-scene/paint/palette";
  import load from "@/components/loading-bar";
  import { t } from "@/lib/i18n";
  import { createChainSnapshot } from "@/lib/iso-scene/chain-snapshot";
  import type {
    IsoMarker,
    IsoPlane,
    IsoSnapshot,
  } from "@/lib/iso-scene/snapshot";
  import type { PowerChain } from "@/lib/power-chain";
  import type { FloorOverview } from "@/pages/main/dashboard/space-group-view/floor-overview.svelte";
  import styles from "./chain-scene-canvas.module.css";

  interface Props {
    session: AuthSession;
    spaceGroup: SpaceGroup;
    floors: Space[];
    chain: PowerChain;
    overview: FloorOverview;
    highlightMarkerId?: string;
    initialViewport?: ViewportState;
    oninspectspace: (id: SpaceId) => void;
    oninspectplace: (key: string) => void;
    onopenplace: (key: string) => void;
    onviewportchange: (viewport: ViewportState | undefined) => void;
  }

  const {
    session,
    spaceGroup,
    floors,
    chain,
    overview,
    highlightMarkerId,
    initialViewport,
    oninspectspace,
    oninspectplace,
    onopenplace,
    onviewportchange,
  }: Props = $props();

  let snapshot = $state.raw<IsoSnapshot>();

  const palette = $derived(paletteForTone(overview.tone));

  const oninspect = (plane: IsoPlane) => {
    const floor = floors.find(({ id }) => id === plane.id);
    if (floor) {
      oninspectspace(floor.id);
    }
  };

  const oninspectmarker = ({ id }: IsoMarker) => {
    if (id !== undefined) {
      oninspectplace(id);
    }
  };

  const onopenmarker = ({ id }: IsoMarker) => {
    if (id !== undefined) {
      onopenplace(id);
    }
  };

  $effect(() => {
    let cancelled = false;
    const isFirstLoad = untrack(() => snapshot === undefined);
    const loading = isFirstLoad
      ? load.begin(t("loading.renderPowerChain"))
      : undefined;

    createChainSnapshot({
      session,
      spaceGroup,
      spaces: floors,
      chain,
      tone: overview.tone,
    })
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
    {highlightMarkerId}
    {initialViewport}
    {oninspect}
    {oninspectmarker}
    {onopenmarker}
    {onviewportchange}
  />
{/if}
