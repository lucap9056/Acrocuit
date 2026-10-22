import type { SpaceId } from "@/api";
import { DEFAULT_TILT_DEG } from "@/components/iso-scene/camera/viewport";
import type IsoSceneView from "@/components/iso-scene/iso-scene.svelte";
import type { FloorPlanTone } from "@/lib/iso-scene/floor-plan-image";

export class FloorOverview {
  tilt = $state(DEFAULT_TILT_DEG);
  tone = $state<FloorPlanTone>("original");
  inspectedSpaceId = $state<SpaceId>();
  view = $state.raw<ReturnType<typeof IsoSceneView>>();

  revealSpace(id: SpaceId) {
    this.inspectedSpaceId = id;
    this.view?.reveal(id);
  }

  get zoomLevel(): number | undefined {
    return this.view?.getZoomLevel();
  }
}
