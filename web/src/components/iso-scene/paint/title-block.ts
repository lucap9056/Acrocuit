import {
  CAPTION_FONT,
  TITLE_FONT,
} from "@/components/iso-scene/layout/typography";
import { setScreenTransform } from "./canvas";
import type { PaintLayer } from "./layer";

const PADDING_PX = 8;

export const titleBlockLayer: PaintLayer = (
  ctx,
  { titleBlock },
  { dpr, palette },
) => {
  if (!titleBlock) {
    return;
  }

  const { x, y, width, height, titleHeight, rowHeight, title, rows } =
    titleBlock;

  setScreenTransform(ctx, dpr);
  ctx.fillStyle = palette.paper;
  ctx.fillRect(x, y, width, height);

  ctx.lineWidth = 1;
  ctx.strokeStyle = palette.line;
  ctx.strokeRect(x + 0.5, y + 0.5, width, height);

  ctx.strokeStyle = palette.lineMuted;
  ctx.beginPath();
  rows.forEach((_, i) => {
    const rowY = Math.round(y + titleHeight + i * rowHeight) + 0.5;
    ctx.moveTo(x, rowY);
    ctx.lineTo(x + width, rowY);
  });
  const dividerX = Math.round(x + width / 2) + 0.5;
  ctx.moveTo(dividerX, y + titleHeight);
  ctx.lineTo(dividerX, y + height);
  ctx.stroke();

  ctx.textBaseline = "middle";
  ctx.textAlign = "left";
  ctx.fillStyle = palette.line;
  ctx.font = TITLE_FONT;
  ctx.fillText(
    title,
    x + PADDING_PX,
    y + titleHeight / 2,
    width - PADDING_PX * 2,
  );

  ctx.font = CAPTION_FONT;
  rows.forEach(([label, value], i) => {
    const rowY = y + titleHeight + i * rowHeight + rowHeight / 2;
    ctx.fillStyle = palette.lineMuted;
    ctx.fillText(label, x + PADDING_PX, rowY);
    ctx.fillStyle = palette.line;
    ctx.fillText(value, dividerX + PADDING_PX, rowY);
  });
};
