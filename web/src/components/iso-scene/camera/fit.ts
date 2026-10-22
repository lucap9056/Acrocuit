import { rectArea, rectCorners, type WorldPoint } from "@/lib/iso-scene/rect";
import type { IsoPlane, IsoSnapshot } from "@/lib/iso-scene/snapshot";
import { type Point, project, type Viewport } from "./projection";
import type { ViewportState } from "./viewport";

const FIT_PADDING_PX = 48;
const LABEL_MARGIN_PX = 180;

interface Bounds {
  minX: number;
  maxX: number;
  minY: number;
  maxY: number;
}

const worldCorners = ({ planes }: IsoSnapshot): WorldPoint[] =>
  planes.flatMap(({ rect }) => rectCorners(rect));

const largestPlane = ({ planes }: IsoSnapshot): IsoPlane | undefined =>
  planes.reduce<IsoPlane | undefined>(
    (largest, plane) =>
      !largest || rectArea(plane.rect) > rectArea(largest.rect)
        ? plane
        : largest,
    undefined,
  );

const boundsOf = (points: Point[]): Bounds => {
  const xs = points.map(({ x }) => x);
  const ys = points.map(({ y }) => y);
  return {
    minX: Math.min(...xs),
    maxX: Math.max(...xs),
    minY: Math.min(...ys),
    maxY: Math.max(...ys),
  };
};

const midpoint = (values: number[]) =>
  values.length === 0 ? 0 : (Math.min(...values) + Math.max(...values)) / 2;

const worldCenter = (snapshot: IsoSnapshot): WorldPoint => {
  const corners = worldCorners(snapshot);
  return {
    x: midpoint(corners.map(({ x }) => x)),
    y: midpoint(corners.map(({ y }) => y)),
    z: midpoint(corners.map(({ z }) => z)),
  };
};

export const fitViewport = (
  snapshot: IsoSnapshot,
  width: number,
  height: number,
  tilt: number,
): ViewportState => {
  const unit: Viewport = { zoom: 1, panX: 0, panY: 0, tilt };
  const toScreen = (corners: WorldPoint[]) =>
    corners.map((corner) => project(corner, unit));
  const largest = largestPlane(snapshot);

  if (!largest || width <= 0 || height <= 0) {
    return { zoom: 1, fitZoom: 1, panX: width / 2, panY: height / 2, tilt };
  }

  const sizing = boundsOf(toScreen(rectCorners(largest.rect)));
  const scene = boundsOf(toScreen(worldCorners(snapshot)));

  const availableWidth = Math.max(
    width - FIT_PADDING_PX * 2 - LABEL_MARGIN_PX,
    1,
  );
  const availableHeight = Math.max(height - FIT_PADDING_PX * 2, 1);
  const zoom = Math.min(
    availableWidth / Math.max(sizing.maxX - sizing.minX, 1),
    availableHeight / Math.max(sizing.maxY - sizing.minY, 1),
  );

  return {
    zoom,
    fitZoom: zoom,
    tilt,
    panX:
      (width + LABEL_MARGIN_PX) / 2 - ((scene.minX + scene.maxX) / 2) * zoom,
    panY: height / 2 - ((scene.minY + scene.maxY) / 2) * zoom,
  };
};

export const retilt = (
  v: ViewportState,
  snapshot: IsoSnapshot,
  width: number,
  height: number,
  tilt: number,
): ViewportState => {
  const { fitZoom } = fitViewport(snapshot, width, height, tilt);
  const zoom = fitZoom * (v.zoom / v.fitZoom);
  const center = worldCenter(snapshot);
  const before = project(center, v);
  const after = project(center, { zoom, tilt, panX: 0, panY: 0 });

  return {
    zoom,
    fitZoom,
    tilt,
    panX: before.x - after.x,
    panY: before.y - after.y,
  };
};
