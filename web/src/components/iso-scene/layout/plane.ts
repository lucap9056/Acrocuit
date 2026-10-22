import {
  type Matrix,
  type Point,
  planeMatrix,
  project,
  type Viewport,
} from "@/components/iso-scene/camera/projection";
import {
  insetRect,
  rectCorners,
  rectElevation,
  type WorldPoint,
} from "@/lib/iso-scene/rect";
import type { IsoPlane } from "@/lib/iso-scene/snapshot";
import { type LabelLayout, layoutLabel } from "./label";

const INNER_BORDER_RATIO = 0.016;

export interface PlaneLayout {
  plane: IsoPlane;
  matrix: Matrix;
  corners: Point[];
  innerCorners: Point[];
  label?: LabelLayout;
}

export const layoutPlane = (
  plane: IsoPlane,
  viewport: Viewport,
): PlaneLayout => {
  const toScreen = (corners: WorldPoint[]) =>
    corners.map((corner) => project(corner, viewport));
  const corners = toScreen(rectCorners(plane.rect));
  const frontLeft = corners[3];
  const [, , , width, depth] = plane.rect;
  const innerBorder = Math.min(width, depth) * INNER_BORDER_RATIO;

  return {
    plane,
    matrix: planeMatrix(rectElevation(plane.rect), viewport),
    corners,
    innerCorners: toScreen(rectCorners(insetRect(plane.rect, innerBorder))),
    label:
      plane.text === undefined || !frontLeft
        ? undefined
        : layoutLabel(plane.text, frontLeft),
  };
};
