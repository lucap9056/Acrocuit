import {
  toDegrees,
  type ViewportState,
  zoomPercent,
} from "@/components/iso-scene/camera/viewport";
import type { IsoSnapshot } from "@/lib/iso-scene/snapshot";

const MARGIN_PX = 16;
const WIDTH_PX = 220;
const ROW_HEIGHT_PX = 20;
const TITLE_HEIGHT_PX = 30;

export interface TitleBlockLayout {
  x: number;
  y: number;
  width: number;
  height: number;
  titleHeight: number;
  rowHeight: number;
  title: string;
  rows: [string, string][];
}

export const layoutTitleBlock = (
  { title, planes }: IsoSnapshot,
  viewport: ViewportState,
  width: number,
  height: number,
): TitleBlockLayout | undefined => {
  const rows: [string, string][] = [
    ["FLOORS", planes.length.toString()],
    ["SCALE", `${zoomPercent(viewport)}%`],
    ["TILT", `${Math.round(toDegrees(viewport.tilt))}°`],
  ];
  const blockHeight = TITLE_HEIGHT_PX + rows.length * ROW_HEIGHT_PX;
  const x = width - MARGIN_PX - WIDTH_PX;
  const y = height - MARGIN_PX - blockHeight;

  if (x < MARGIN_PX || y < MARGIN_PX) {
    return undefined;
  }

  return {
    x,
    y,
    width: WIDTH_PX,
    height: blockHeight,
    titleHeight: TITLE_HEIGHT_PX,
    rowHeight: ROW_HEIGHT_PX,
    title: title.toUpperCase(),
    rows,
  };
};
