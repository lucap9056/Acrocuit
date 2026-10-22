import type { Viewport } from "./projection";

export interface ViewportState extends Viewport {
  fitZoom: number;
}

export const MIN_TILT_DEG = 10;
export const MAX_TILT_DEG = 80;
export const DEFAULT_TILT_DEG = 30;

export const toRadians = (degrees: number): number => (degrees * Math.PI) / 180;

export const toDegrees = (radians: number): number => (radians * 180) / Math.PI;

export const zoomPercent = (v: ViewportState): number =>
  Math.round((v.zoom / v.fitZoom) * 100);
