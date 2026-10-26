import type { Point } from "@/components/iso-scene/camera/projection";
import type { ResolvedIcon } from "@/lib/icon-packs/index.svelte";

const paths = new Map<string, Path2D>();

const toPath = (d: string) => {
  let path = paths.get(d);
  if (!path) {
    path = new Path2D(d);
    paths.set(d, path);
  }
  return path;
};

export const drawPackIcon = (
  ctx: CanvasRenderingContext2D,
  { pack, icon }: ResolvedIcon,
  { x, y }: Point,
  size: number,
  color: string,
  strokeScale = 1,
) => {
  const [minX, minY, width, height] = pack.style.viewBox.split(" ").map(Number);
  const scale = size / Math.max(width, height);

  ctx.save();
  ctx.translate(x - (width * scale) / 2, y - (height * scale) / 2);
  ctx.scale(scale, scale);
  ctx.translate(-minX, -minY);
  ctx.fillStyle = color;
  for (const d of icon.fills) {
    ctx.fill(toPath(d));
  }
  ctx.strokeStyle = color;
  ctx.lineWidth = pack.style.strokeWidth * strokeScale;
  ctx.lineCap = pack.style.lineCap;
  ctx.lineJoin = pack.style.lineJoin;
  for (const d of icon.strokes) {
    ctx.stroke(toPath(d));
  }
  ctx.restore();
};
