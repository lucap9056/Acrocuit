import {
  type Point,
  project,
  type Viewport,
} from "@/components/iso-scene/camera/projection";
import type { IsoLink, IsoMarker } from "@/lib/iso-scene/snapshot";

export interface MarkerLayout {
  marker: IsoMarker;
  point: Point;
}

export const layoutMarker = (
  marker: IsoMarker,
  viewport: Viewport,
): MarkerLayout => ({
  marker,
  point: project(marker.point, viewport),
});

export interface LinkLayout {
  link: IsoLink;
  points: Point[];
}

export const layoutLink = (link: IsoLink, viewport: Viewport): LinkLayout => ({
  link,
  points: link.points.map((point) => project(point, viewport)),
});
