import type { Matrix, Point } from "@/components/iso-scene/camera/projection";

export const setPlaneTransform = (
  ctx: CanvasRenderingContext2D,
  [a, b, c, d, e, f]: Matrix,
  dpr: number,
) => {
  ctx.setTransform(a * dpr, b * dpr, c * dpr, d * dpr, e * dpr, f * dpr);
};

export const setScreenTransform = (
  ctx: CanvasRenderingContext2D,
  dpr: number,
) => {
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
};

export const tracePolygon = (
  ctx: CanvasRenderingContext2D,
  points: Point[],
) => {
  ctx.beginPath();
  points.forEach((p, i) => {
    if (i === 0) {
      ctx.moveTo(p.x, p.y);
    } else {
      ctx.lineTo(p.x, p.y);
    }
  });
  ctx.closePath();
};
