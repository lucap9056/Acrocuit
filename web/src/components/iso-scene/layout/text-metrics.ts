const FALLBACK_CHAR_WIDTH_PX = 7.4;

const measureContext = new OffscreenCanvas(1, 1).getContext("2d");

export const measureTextWidth = (text: string, font: string): number => {
  if (!measureContext) {
    return text.length * FALLBACK_CHAR_WIDTH_PX;
  }
  measureContext.font = font;
  return measureContext.measureText(text).width;
};
