import type { IsoSnapshot } from "@/lib/iso-scene/snapshot";
import { fitViewport, retilt } from "./fit";
import { toRadians, type ViewportState } from "./viewport";

export const resolveViewport = (
  snapshot: IsoSnapshot,
  tiltDegrees: number,
  width: number,
  height: number,
  override?: ViewportState,
): ViewportState => {
  const tilt = toRadians(tiltDegrees);
  if (!override) {
    return fitViewport(snapshot, width, height, tilt);
  }
  if (override.tilt !== tilt) {
    return retilt(override, snapshot, width, height, tilt);
  }
  return override;
};
