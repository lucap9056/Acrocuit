import type { SceneLayout } from "@/components/iso-scene/layout/scene-layout";
import type { ScenePalette } from "./palette";

export interface PaintContext {
  dpr: number;
  highlightId?: number;
  palette: ScenePalette;
}

export type PaintLayer = (
  ctx: CanvasRenderingContext2D,
  layout: SceneLayout,
  paint: PaintContext,
) => void;
