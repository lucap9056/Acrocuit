import { LABEL_FONT } from "@/components/iso-scene/layout/typography";
import { setScreenTransform } from "./canvas";
import type { PaintLayer } from "./layer";

export const labelsLayer: PaintLayer = (
  ctx,
  { planes },
  { dpr, highlightId, palette },
) => {
  setScreenTransform(ctx, dpr);
  ctx.textBaseline = "middle";
  ctx.textAlign = "right";
  ctx.font = LABEL_FONT;

  for (const { plane, label } of planes) {
    if (!label) {
      continue;
    }
    const { anchor, leaderEndX, textX, text } = label;
    const isHighlighted = plane.id === highlightId;

    ctx.lineWidth = 1;
    ctx.strokeStyle = isHighlighted
      ? palette.highlight.line
      : palette.lineMuted;
    ctx.beginPath();
    ctx.moveTo(anchor.x, anchor.y);
    ctx.lineTo(leaderEndX, anchor.y);
    ctx.stroke();

    ctx.fillStyle = isHighlighted ? palette.highlight.line : palette.line;
    ctx.beginPath();
    ctx.arc(anchor.x, anchor.y, 2, 0, Math.PI * 2);
    ctx.fill();

    ctx.fillText(text, textX, anchor.y);
  }
};
