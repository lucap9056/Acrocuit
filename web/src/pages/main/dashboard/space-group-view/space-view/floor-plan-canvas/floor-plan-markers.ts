import type { Point } from "@/components/iso-scene/camera/projection";
import { LABEL_FONT } from "@/components/iso-scene/layout/typography";
import { setScreenTransform } from "@/components/iso-scene/paint/canvas";
import type { ScenePalette } from "@/components/iso-scene/paint/palette";

const SYMBOL_SIZE_PX = 27;
const SYMBOL_COLOR = "#2f6fd6";
const DEVICE_RADIUS_PX = 12;
const DEVICE_STROKE_PX = 3;
const PANEL_HALF_WIDTH_PX = 12.4;
const PANEL_HALF_HEIGHT_PX = 6.4;
const PANEL_WEDGE_HALF_HEIGHT_PX = 5.8;
const PANEL_STROKE_PX = 2.25;
const OUTLINE_PX = 2;
const LABEL_GAP_PX = 4;
const LABEL_PADDING_X_PX = 4;
const LABEL_HEIGHT_PX = 18;

export interface FloorPlanMarkerShape {
  kind: "switch_panel" | "device";
  name: string;
  point: Point;
  isActive: boolean;
}

const inDrawOrder = <T extends FloorPlanMarkerShape>(markers: T[]): T[] => [
  ...markers.filter(({ isActive }) => !isActive),
  ...markers.filter(({ isActive }) => isActive),
];

const drawSymbol = (
  ctx: CanvasRenderingContext2D,
  { kind, point: { x, y } }: FloorPlanMarkerShape,
) => {
  ctx.strokeStyle = SYMBOL_COLOR;
  ctx.fillStyle = SYMBOL_COLOR;
  if (kind === "device") {
    ctx.lineWidth = DEVICE_STROKE_PX;
    ctx.beginPath();
    ctx.arc(x, y, DEVICE_RADIUS_PX, 0, Math.PI * 2);
    ctx.stroke();
    return;
  }

  ctx.lineWidth = PANEL_STROKE_PX;
  ctx.strokeRect(
    x - PANEL_HALF_WIDTH_PX,
    y - PANEL_HALF_HEIGHT_PX,
    PANEL_HALF_WIDTH_PX * 2,
    PANEL_HALF_HEIGHT_PX * 2,
  );
  ctx.beginPath();
  ctx.moveTo(x - PANEL_HALF_WIDTH_PX, y + PANEL_WEDGE_HALF_HEIGHT_PX);
  ctx.lineTo(x + PANEL_HALF_WIDTH_PX, y - PANEL_WEDGE_HALF_HEIGHT_PX);
  ctx.lineTo(x + PANEL_HALF_WIDTH_PX, y + PANEL_WEDGE_HALF_HEIGHT_PX);
  ctx.closePath();
  ctx.fill();
};

const drawOutline = (ctx: CanvasRenderingContext2D, { x, y }: Point) => {
  const half = SYMBOL_SIZE_PX / 2 + OUTLINE_PX / 2;
  ctx.lineWidth = OUTLINE_PX;
  ctx.strokeRect(x - half, y - half, half * 2, half * 2);
};

const drawLabel = (
  ctx: CanvasRenderingContext2D,
  { name, point: { x, y }, isActive }: FloorPlanMarkerShape,
  palette: ScenePalette,
) => {
  const width = ctx.measureText(name).width + LABEL_PADDING_X_PX * 2 + 2;
  const left = x - width / 2;
  const top = y + SYMBOL_SIZE_PX / 2 + LABEL_GAP_PX;

  ctx.fillStyle = palette.paper;
  ctx.fillRect(left, top, width, LABEL_HEIGHT_PX);
  ctx.lineWidth = 1;
  ctx.strokeStyle = isActive ? palette.highlight.line : palette.lineMuted;
  ctx.strokeRect(left + 0.5, top + 0.5, width - 1, LABEL_HEIGHT_PX - 1);
  ctx.fillStyle = isActive ? palette.highlight.line : palette.line;
  ctx.fillText(name, x, top + LABEL_HEIGHT_PX / 2);
};

export const paintFloorPlanMarkers = (
  ctx: CanvasRenderingContext2D,
  markers: FloorPlanMarkerShape[],
  dpr: number,
  palette: ScenePalette,
) => {
  setScreenTransform(ctx, dpr);
  ctx.font = LABEL_FONT;
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";

  for (const marker of inDrawOrder(markers)) {
    drawSymbol(ctx, marker);
    if (marker.isActive) {
      ctx.strokeStyle = palette.highlight.line;
      drawOutline(ctx, marker.point);
    }
    drawLabel(ctx, marker, palette);
  }
};

export const hitTestFloorPlanMarkers = <T extends FloorPlanMarkerShape>(
  markers: T[],
  { x, y }: Point,
): T | undefined =>
  inDrawOrder(markers).findLast(
    ({ point }) =>
      Math.abs(point.x - x) <= SYMBOL_SIZE_PX / 2 &&
      Math.abs(point.y - y) <= SYMBOL_SIZE_PX / 2,
  );
