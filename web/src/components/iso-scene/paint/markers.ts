import { CAPTION_FONT } from "@/components/iso-scene/layout/typography";
import { setScreenTransform } from "./canvas";
import type { PaintLayer } from "./layer";
import { groupColor } from "./palette";

const MARKER_RADIUS_PX = 4.5;
const HIGHLIGHT_RADIUS_PX = 6.5;
const HIGHLIGHT_RING_GAP_PX = 3.5;
const GLOW_BLUR_PX = 10;
const TEXT_GAP_PX = 6;

export const markersLayer: PaintLayer = (
  ctx,
  { markers },
  { dpr, highlightMarkerId, palette },
) => {
  if (markers.length === 0) {
    return;
  }

  setScreenTransform(ctx, dpr);
  ctx.font = CAPTION_FONT;
  ctx.textAlign = "left";
  ctx.textBaseline = "middle";
  ctx.lineJoin = "round";

  for (const { marker, point } of markers) {
    const isHighlighted =
      marker.id !== undefined && marker.id === highlightMarkerId;
    const radius = isHighlighted ? HIGHLIGHT_RADIUS_PX : MARKER_RADIUS_PX;

    if (isHighlighted) {
      ctx.save();
      ctx.lineWidth = 2;
      ctx.strokeStyle = palette.highlight.line;
      if (palette.highlight.glow) {
        ctx.shadowColor = palette.highlight.glow;
        ctx.shadowBlur = GLOW_BLUR_PX;
      }
      ctx.beginPath();
      ctx.arc(point.x, point.y, radius + HIGHLIGHT_RING_GAP_PX, 0, Math.PI * 2);
      ctx.stroke();
      ctx.restore();
    }

    ctx.beginPath();
    if (marker.shape === "circle") {
      ctx.arc(point.x, point.y, radius, 0, Math.PI * 2);
    } else {
      ctx.rect(point.x - radius, point.y - radius, radius * 2, radius * 2);
    }
    ctx.fillStyle = groupColor(palette, marker.group) ?? palette.trace.line;
    ctx.fill();
    ctx.lineWidth = 2;
    ctx.strokeStyle = palette.paper;
    ctx.stroke();

    if (marker.text) {
      const textX = point.x + radius + TEXT_GAP_PX;
      ctx.lineWidth = 3;
      ctx.strokeText(marker.text, textX, point.y);
      ctx.fillStyle = isHighlighted ? palette.highlight.line : palette.line;
      ctx.fillText(marker.text, textX, point.y);
    }
  }
};
