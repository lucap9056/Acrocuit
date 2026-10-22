import { setScreenTransform, tracePolygon } from "./canvas";
import type { PaintLayer } from "./layer";

const GLOW_BLUR_PX = 12;

export const highlightLayer: PaintLayer = (
  ctx,
  { planes },
  { dpr, highlightId, palette: { highlight } },
) => {
  const highlighted = planes.find(({ plane }) => plane.id === highlightId);
  if (!highlighted) {
    return;
  }

  setScreenTransform(ctx, dpr);
  ctx.lineWidth = 2;
  ctx.strokeStyle = highlight.line;
  if (highlight.glow) {
    ctx.shadowColor = highlight.glow;
    ctx.shadowBlur = GLOW_BLUR_PX;
  }
  tracePolygon(ctx, highlighted.corners);
  ctx.stroke();
};
