export type PlaneRect = readonly [
  x: number,
  y: number,
  z: number,
  width: number,
  depth: number,
];

export interface IsoPlane {
  id: number;
  rect: PlaneRect;
  image?: ImageBitmap;
  text?: string;
}

export interface IsoSnapshot {
  title: string;
  planes: IsoPlane[];
}
