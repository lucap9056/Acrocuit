import type { WorldPoint } from "./rect";

export type PlaneRect = readonly [
  x: number,
  y: number,
  z: number,
  width: number,
  depth: number,
];

export interface IsoPlane {
  id: string;
  rect: PlaneRect;
  image?: ImageBitmap;
  text?: string;
  group?: number;
}

export type IsoMarkerShape = "circle" | "square";

export interface IsoMarker {
  id?: string;
  point: WorldPoint;
  shape: IsoMarkerShape;
  text?: string;
  group?: number;
}

export interface IsoLink {
  points: WorldPoint[];
  group?: number;
}

export interface IsoSnapshot {
  title: string;
  planes: IsoPlane[];
  markers?: IsoMarker[];
  links?: IsoLink[];
}
