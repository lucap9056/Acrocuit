import type { PlaneRect } from "./snapshot";

export interface WorldPoint {
  x: number;
  y: number;
  z: number;
}

export const rectElevation = (rect: PlaneRect): number => rect[2];

export const rectArea = ([, , , width, depth]: PlaneRect): number =>
  width * depth;

export const rectCorners = ([
  x,
  y,
  z,
  width,
  depth,
]: PlaneRect): WorldPoint[] => [
  { x, y, z },
  { x: x + width, y, z },
  { x: x + width, y: y + depth, z },
  { x, y: y + depth, z },
];

export const insetRect = (
  [x, y, z, width, depth]: PlaneRect,
  inset: number,
): PlaneRect => [x + inset, y + inset, z, width - inset * 2, depth - inset * 2];

export const rectContains = (
  [x, y, , width, depth]: PlaneRect,
  point: { x: number; y: number },
): boolean =>
  point.x >= x && point.x <= x + width && point.y >= y && point.y <= y + depth;
