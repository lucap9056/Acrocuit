const MAX_BITMAP_SIZE = 2048;
const INK_RGB = [233, 242, 255] as const;

export type FloorPlanTone = "original" | "blueprint";

export interface FloorPlanImage {
  bitmap: ImageBitmap;
  width: number;
  height: number;
}

const downscale = async (source: ImageBitmap): Promise<ImageBitmap> => {
  const scale = Math.min(
    1,
    MAX_BITMAP_SIZE / Math.max(source.width, source.height),
  );
  if (scale === 1) {
    return source;
  }

  const bitmap = await createImageBitmap(source, {
    resizeWidth: Math.max(1, Math.round(source.width * scale)),
    resizeHeight: Math.max(1, Math.round(source.height * scale)),
    resizeQuality: "high",
  });
  source.close();
  return bitmap;
};

const inkBitmap = (source: ImageBitmap): ImageBitmap => {
  const { width, height } = source;
  const canvas = new OffscreenCanvas(width, height);
  const ctx = canvas.getContext("2d", { willReadFrequently: true });
  if (!ctx) {
    return source;
  }

  ctx.drawImage(source, 0, 0);
  source.close();

  const image = ctx.getImageData(0, 0, width, height);
  const { data } = image;
  for (let i = 0; i < data.length; i += 4) {
    const luminance =
      0.299 * data[i] + 0.587 * data[i + 1] + 0.114 * data[i + 2];
    const ink = (255 - luminance) * (data[i + 3] / 255);
    data[i] = INK_RGB[0];
    data[i + 1] = INK_RGB[1];
    data[i + 2] = INK_RGB[2];
    data[i + 3] = ink;
  }
  ctx.putImageData(image, 0, 0);

  return canvas.transferToImageBitmap();
};

export const toFloorPlanImage = async (
  blob: Blob,
  tone: FloorPlanTone,
): Promise<FloorPlanImage> => {
  const source = await createImageBitmap(blob);
  const { width, height } = source;
  const resized = await downscale(source);
  return {
    bitmap: tone === "blueprint" ? inkBitmap(resized) : resized,
    width,
    height,
  };
};
