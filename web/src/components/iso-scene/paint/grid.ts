import type { GridLayout, GridLines } from "@/components/iso-scene/layout/grid";
import { setScreenTransform } from "./canvas";
import type { PaintLayer } from "./layer";
import type { ScenePalette } from "./palette";

const drawLines = (
  ctx: CanvasRenderingContext2D,
  { spacing, offsetX, offsetY }: GridLines,
  width: number,
  height: number,
) => {
  ctx.beginPath();
  for (let x = offsetX; x <= width; x += spacing) {
    ctx.moveTo(Math.round(x) + 0.5, 0);
    ctx.lineTo(Math.round(x) + 0.5, height);
  }
  for (let y = offsetY; y <= height; y += spacing) {
    ctx.moveTo(0, Math.round(y) + 0.5);
    ctx.lineTo(width, Math.round(y) + 0.5);
  }
  ctx.stroke();
};

export const paintGrid = (
  ctx: CanvasRenderingContext2D,
  grid: GridLayout,
  width: number,
  height: number,
  dpr: number,
  palette: ScenePalette,
) => {
  setScreenTransform(ctx, dpr);
  ctx.fillStyle = palette.paper;
  ctx.fillRect(0, 0, width, height);

  ctx.lineWidth = 1;
  ctx.strokeStyle = palette.gridMinor;
  drawLines(ctx, grid.minor, width, height);
  ctx.strokeStyle = palette.gridMajor;
  drawLines(ctx, grid.major, width, height);
};

export const gridLayer: PaintLayer = (
  ctx,
  { grid, width, height },
  { dpr, palette },
) => {
  paintGrid(ctx, grid, width, height, dpr, palette);
};
