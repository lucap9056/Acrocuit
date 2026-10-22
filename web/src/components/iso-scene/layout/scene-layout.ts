import type { ViewportState } from "@/components/iso-scene/camera/viewport";
import type { IsoSnapshot } from "@/lib/iso-scene/snapshot";
import { type GridLayout, layoutGrid } from "./grid";
import {
  type LinkLayout,
  layoutLink,
  layoutMarker,
  type MarkerLayout,
} from "./marker";
import { layoutPlane, type PlaneLayout } from "./plane";
import { layoutTitleBlock, type TitleBlockLayout } from "./title-block";

export interface SceneLayout {
  width: number;
  height: number;
  viewport: ViewportState;
  planes: PlaneLayout[];
  markers: MarkerLayout[];
  links: LinkLayout[];
  grid: GridLayout;
  titleBlock?: TitleBlockLayout;
}

export const layoutScene = (
  snapshot: IsoSnapshot,
  viewport: ViewportState,
  width: number,
  height: number,
): SceneLayout => ({
  width,
  height,
  viewport,
  planes: snapshot.planes.map((plane) => layoutPlane(plane, viewport)),
  markers: (snapshot.markers ?? []).map((marker) =>
    layoutMarker(marker, viewport),
  ),
  links: (snapshot.links ?? []).map((link) => layoutLink(link, viewport)),
  grid: layoutGrid(viewport),
  titleBlock: layoutTitleBlock(snapshot, viewport, width, height),
});
