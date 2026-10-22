import type { Point } from "@/components/iso-scene/camera/projection";
import { setScreenTransform, tracePolyline } from "./canvas";
import type { PaintLayer } from "./layer";
import { groupColor } from "./palette";

const GLOW_BLUR_PX = 8;
const LINE_WIDTH_PX = 2;
const CASING_WIDTH_PX = 4.5;

interface ColoredLink {
  points: Point[];
  color: string;
}

const strokePath = (ctx: CanvasRenderingContext2D, points: Point[]) => {
  tracePolyline(ctx, points);
  ctx.stroke();
};

export const linksLayer: PaintLayer = (ctx, { links }, { dpr, palette }) => {
  if (links.length === 0) {
    return;
  }

  const ungrouped: Point[][] = [];
  const grouped: ColoredLink[] = [];
  for (const { link, points } of links) {
    const color = groupColor(palette, link.group);
    if (color) {
      grouped.push({ points, color });
    } else {
      ungrouped.push(points);
    }
  }

  setScreenTransform(ctx, dpr);
  ctx.lineJoin = "round";
  ctx.lineCap = "round";

  ctx.save();
  ctx.lineWidth = LINE_WIDTH_PX;
  ctx.strokeStyle = palette.trace.line;
  if (palette.trace.glow) {
    ctx.shadowColor = palette.trace.glow;
    ctx.shadowBlur = GLOW_BLUR_PX;
  }
  for (const points of ungrouped) {
    strokePath(ctx, points);
  }
  ctx.restore();

  ctx.lineWidth = CASING_WIDTH_PX;
  ctx.strokeStyle = palette.paper;
  for (const { points } of grouped) {
    strokePath(ctx, points);
  }

  ctx.lineWidth = LINE_WIDTH_PX;
  for (const { points, color } of grouped) {
    ctx.strokeStyle = color;
    strokePath(ctx, points);
  }
};
