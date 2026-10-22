import type { SceneLayout } from "@/components/iso-scene/layout/scene-layout";
import { gridLayer } from "./grid";
import { highlightLayer } from "./highlight";
import { labelsLayer } from "./labels";
import type { PaintContext, PaintLayer } from "./layer";
import { linksLayer } from "./links";
import { markersLayer } from "./markers";
import { planesLayer } from "./planes";
import { titleBlockLayer } from "./title-block";

export const defaultLayers: PaintLayer[] = [
  gridLayer,
  planesLayer,
  highlightLayer,
  linksLayer,
  markersLayer,
  labelsLayer,
  titleBlockLayer,
];

export interface IsoPainter {
  paint: (layout: SceneLayout, context: PaintContext) => void;
  dispose: () => void;
}

interface PendingPaint {
  layout: SceneLayout;
  context: PaintContext;
}

export const createIsoPainter = (
  canvas: HTMLCanvasElement,
  layers: PaintLayer[] = defaultLayers,
): IsoPainter => {
  const ctx = canvas.getContext("2d");
  let pending: PendingPaint | undefined;
  let frameRequest: number | undefined;

  const cancelFrame = () => {
    if (frameRequest !== undefined) {
      cancelAnimationFrame(frameRequest);
      frameRequest = undefined;
    }
  };

  const draw = () => {
    frameRequest = undefined;
    if (!ctx || !pending) {
      return;
    }
    const { layout, context } = pending;
    if (layout.width === 0 || layout.height === 0) {
      return;
    }

    ctx.setTransform(1, 0, 0, 1, 0, 0);
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    for (const layer of layers) {
      ctx.save();
      layer(ctx, layout, context);
      ctx.restore();
    }
  };

  const fitCanvas = ({ width, height }: SceneLayout, dpr: number): boolean => {
    const pixelWidth = Math.round(width * dpr);
    const pixelHeight = Math.round(height * dpr);
    if (canvas.width === pixelWidth && canvas.height === pixelHeight) {
      return false;
    }
    canvas.width = pixelWidth;
    canvas.height = pixelHeight;
    return true;
  };

  return {
    paint: (layout, context) => {
      pending = { layout, context };
      if (fitCanvas(layout, context.dpr)) {
        cancelFrame();
        draw();
      } else {
        frameRequest ??= requestAnimationFrame(draw);
      }
    },
    dispose: cancelFrame,
  };
};
