import type { Point } from "./projection";
import type { ViewportState } from "./viewport";

const MIN_ZOOM_RATIO = 0.25;
const MAX_ZOOM_RATIO = 16;

export const zoomAt = (
  v: ViewportState,
  factor: number,
  anchor: Point,
): ViewportState => {
  const zoom = Math.min(
    Math.max(v.zoom * factor, v.fitZoom * MIN_ZOOM_RATIO),
    v.fitZoom * MAX_ZOOM_RATIO,
  );
  const ratio = zoom / v.zoom;

  return {
    ...v,
    zoom,
    panX: anchor.x - (anchor.x - v.panX) * ratio,
    panY: anchor.y - (anchor.y - v.panY) * ratio,
  };
};

export const panBy = (
  v: ViewportState,
  dx: number,
  dy: number,
): ViewportState => ({
  ...v,
  panX: v.panX + dx,
  panY: v.panY + dy,
});

const axisShift = (
  min: number,
  max: number,
  size: number,
  padding: number,
): number => {
  if (max - min > size - padding * 2) {
    return size / 2 - (min + max) / 2;
  }
  if (min < padding) {
    return padding - min;
  }
  if (max > size - padding) {
    return size - padding - max;
  }
  return 0;
};

export const revealPoints = (
  v: ViewportState,
  points: Point[],
  width: number,
  height: number,
  padding: number,
): ViewportState => {
  if (points.length === 0) {
    return v;
  }
  const xs = points.map(({ x }) => x);
  const ys = points.map(({ y }) => y);
  const dx = axisShift(Math.min(...xs), Math.max(...xs), width, padding);
  const dy = axisShift(Math.min(...ys), Math.max(...ys), height, padding);
  return dx === 0 && dy === 0 ? v : panBy(v, dx, dy);
};
