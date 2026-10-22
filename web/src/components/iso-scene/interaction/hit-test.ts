import {
  type Point,
  unproject,
} from "@/components/iso-scene/camera/projection";
import type { LabelLayout } from "@/components/iso-scene/layout/label";
import type { SceneLayout } from "@/components/iso-scene/layout/scene-layout";
import { rectContains, rectElevation } from "@/lib/iso-scene/rect";
import type { IsoMarker, IsoPlane } from "@/lib/iso-scene/snapshot";

const MARKER_HIT_RADIUS_PX = 10;

export type SceneHit =
  | { kind: "marker"; marker: IsoMarker }
  | { kind: "plane"; plane: IsoPlane };

const labelContains = (
  { left, top, bottom, anchor }: LabelLayout,
  { x, y }: Point,
) => x >= left && x <= anchor.x && y >= top && y <= bottom;

export const hitTest = (
  { markers, planes, viewport }: SceneLayout,
  point: Point,
): SceneHit | undefined => {
  const marker = markers.findLast(
    (candidate) =>
      candidate.marker.id !== undefined &&
      Math.hypot(candidate.point.x - point.x, candidate.point.y - point.y) <=
        MARKER_HIT_RADIUS_PX,
  )?.marker;
  if (marker) {
    return { kind: "marker", marker };
  }

  const plane = (
    planes.findLast(({ label }) => label && labelContains(label, point)) ??
    planes.findLast(({ plane }) =>
      rectContains(
        plane.rect,
        unproject(point, rectElevation(plane.rect), viewport),
      ),
    )
  )?.plane;
  return plane ? { kind: "plane", plane } : undefined;
};
