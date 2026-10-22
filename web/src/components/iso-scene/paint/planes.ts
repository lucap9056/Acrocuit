import type { Point } from "@/components/iso-scene/camera/projection";
import { setPlaneTransform, setScreenTransform, tracePolygon } from "./canvas";
import type { PaintLayer } from "./layer";
import type { ScenePalette } from "./palette";

const drawStackConnectors = (
  ctx: CanvasRenderingContext2D,
  below: Point[],
  above: Point[],
  palette: ScenePalette,
) => {
  ctx.setLineDash([4, 4]);
  ctx.lineWidth = 1;
  ctx.strokeStyle = palette.lineFaint;
  ctx.beginPath();
  above.forEach((corner, i) => {
    const start = below[i];
    if (start) {
      ctx.moveTo(start.x, start.y);
      ctx.lineTo(corner.x, corner.y);
    }
  });
  ctx.stroke();
  ctx.setLineDash([]);
};

export const planesLayer: PaintLayer = (ctx, { planes }, { dpr, palette }) => {
  planes.forEach(({ plane, matrix, corners, innerCorners }, index) => {
    const below = planes[index - 1];
    if (below) {
      setScreenTransform(ctx, dpr);
      drawStackConnectors(ctx, below.corners, corners, palette);
    }

    const [x, y, , width, depth] = plane.rect;
    setPlaneTransform(ctx, matrix, dpr);
    ctx.fillStyle = palette.plane;
    ctx.fillRect(x, y, width, depth);
    if (plane.image) {
      ctx.drawImage(plane.image, x, y, width, depth);
    }

    setScreenTransform(ctx, dpr);
    ctx.lineWidth = 1.5;
    ctx.strokeStyle = palette.line;
    tracePolygon(ctx, corners);
    ctx.stroke();

    ctx.lineWidth = 0.75;
    ctx.strokeStyle = palette.lineMuted;
    tracePolygon(ctx, innerCorners);
    ctx.stroke();
  });
};
