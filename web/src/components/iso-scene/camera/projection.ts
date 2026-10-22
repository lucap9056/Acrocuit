import type { WorldPoint } from "@/lib/iso-scene/rect";

export interface Point {
  x: number;
  y: number;
}

export interface Viewport {
  zoom: number;
  panX: number;
  panY: number;
  tilt: number;
}

export type Matrix = readonly [number, number, number, number, number, number];

const ISO_X = Math.SQRT1_2;

const planeY = (tilt: number) => Math.SQRT1_2 * Math.sin(tilt);
const heightY = (tilt: number) => Math.cos(tilt);

export const project = ({ x, y, z }: WorldPoint, v: Viewport): Point => ({
  x: (x - y) * ISO_X * v.zoom + v.panX,
  y: ((x + y) * planeY(v.tilt) - z * heightY(v.tilt)) * v.zoom + v.panY,
});

export const unproject = (p: Point, z: number, v: Viewport): Point => {
  const difference = (p.x - v.panX) / (v.zoom * ISO_X);
  const sum = ((p.y - v.panY) / v.zoom + z * heightY(v.tilt)) / planeY(v.tilt);
  return { x: (sum + difference) / 2, y: (sum - difference) / 2 };
};

export const planeMatrix = (z: number, v: Viewport): Matrix => [
  ISO_X * v.zoom,
  planeY(v.tilt) * v.zoom,
  -ISO_X * v.zoom,
  planeY(v.tilt) * v.zoom,
  v.panX,
  v.panY - z * heightY(v.tilt) * v.zoom,
];
