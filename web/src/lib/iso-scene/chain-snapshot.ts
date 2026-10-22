import type {
  AuthSession,
  Position,
  Ref,
  Space,
  SpaceGroup,
  SpaceId,
} from "@/api";
import {
  type ChainPlace,
  floorGroups,
  type PowerChain,
} from "@/lib/power-chain";
import type { FloorPlanTone } from "./floor-plan-image";
import { createFloorSnapshot } from "./floor-snapshot";
import type { WorldPoint } from "./rect";
import type { IsoLink, IsoMarker, IsoSnapshot, PlaneRect } from "./snapshot";

export interface ChainSnapshotInput {
  session: AuthSession;
  spaceGroup: SpaceGroup;
  spaces: Space[];
  chain: PowerChain;
  tone: FloorPlanTone;
}

const locate = (
  rects: Map<number, PlaneRect>,
  space: Ref<SpaceId> | null,
  position: Position | null,
): WorldPoint | undefined => {
  const rect = space && rects.get(space.id);
  if (!rect || !position) {
    return undefined;
  }
  const [x, y, z, width, depth] = rect;
  return { x: x + width / 2 + position.x, y: y + depth / 2 + position.y, z };
};

const riser = (
  from: WorldPoint,
  to: WorldPoint,
  group: number | undefined,
): IsoLink => ({
  points:
    from.z === to.z ? [from, to] : [from, { x: to.x, y: to.y, z: from.z }, to],
  group,
});

export const createChainSnapshot = async ({
  session,
  spaceGroup,
  spaces,
  chain,
  tone,
}: ChainSnapshotInput): Promise<IsoSnapshot> => {
  const floors = await createFloorSnapshot({
    session,
    spaceGroup,
    spaces,
    tone,
  });
  const rects = new Map(floors.planes.map(({ id, rect }) => [id, rect]));
  const groups = floorGroups(chain, spaces);
  const placeGroup = (place: ChainPlace | undefined) =>
    !place || place.isFocus || !place.space
      ? undefined
      : groups.get(place.space.id);

  const places = new Map(chain.places.map((place) => [place.key, place]));
  const points = new Map<string, WorldPoint>();
  const markers: IsoMarker[] = [];
  for (const place of chain.places) {
    const point = locate(rects, place.space, place.position);
    if (point) {
      points.set(place.key, point);
      markers.push({
        id: place.key,
        point,
        shape: place.shape,
        text: place.text,
        group: placeGroup(place),
      });
    }
  }

  const links = chain.edges.flatMap(([downstream, upstream]) => {
    const start = points.get(downstream);
    const end = points.get(upstream);
    return start && end
      ? [riser(start, end, placeGroup(places.get(upstream)))]
      : [];
  });

  return {
    ...floors,
    title: chain.title,
    planes: floors.planes.map((plane) => ({
      ...plane,
      group: groups.get(plane.id as SpaceId),
    })),
    markers,
    links,
  };
};
