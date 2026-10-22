import type { Viewport } from "@/components/iso-scene/camera/projection";

const BASE_SPACING = 25;
const MIN_SPACING_PX = 8;
const MAJOR_EVERY = 5;

export interface GridLines {
  spacing: number;
  offsetX: number;
  offsetY: number;
}

export interface GridLayout {
  minor: GridLines;
  major: GridLines;
}

const wrap = (pan: number, step: number) => ((pan % step) + step) % step;

const linesOf = (spacing: number, { panX, panY }: Viewport): GridLines => ({
  spacing,
  offsetX: wrap(panX, spacing),
  offsetY: wrap(panY, spacing),
});

export const layoutGrid = (viewport: Viewport): GridLayout => {
  let spacing = BASE_SPACING * viewport.zoom;
  while (spacing < MIN_SPACING_PX) {
    spacing *= MAJOR_EVERY;
  }

  return {
    minor: linesOf(spacing, viewport),
    major: linesOf(spacing * MAJOR_EVERY, viewport),
  };
};
