import type { AuthSession, Space, SpaceGroup } from "@/api";
import { fetchFloorPlanImage } from "@/lib/cache/space-images";
import type { FloorPlanImage, FloorPlanTone } from "./floor-plan-image";
import type { IsoSnapshot } from "./snapshot";

const DEFAULT_FOOTPRINT: Footprint = { width: 1000, height: 750 };
const FLOOR_GAP_RATIO = 0.26;

interface Footprint {
  width: number;
  height: number;
}

export interface FloorSnapshotInput {
  session: AuthSession;
  spaceGroup: SpaceGroup;
  spaces: Space[];
  tone: FloorPlanTone;
}

const loadImage = (
  session: AuthSession,
  space: Space,
  tone: FloorPlanTone,
): Promise<FloorPlanImage | undefined> =>
  fetchFloorPlanImage(session, space, tone).catch((err) => {
    console.error(err);
    return undefined;
  });

const largestFootprint = (
  images: (FloorPlanImage | undefined)[],
): Footprint => {
  let largest: Footprint | undefined;
  for (const image of images) {
    if (
      image &&
      (!largest || image.width * image.height > largest.width * largest.height)
    ) {
      largest = image;
    }
  }
  return largest ?? DEFAULT_FOOTPRINT;
};

export const createFloorSnapshot = async ({
  session,
  spaceGroup,
  spaces,
  tone,
}: FloorSnapshotInput): Promise<IsoSnapshot> => {
  const images = await Promise.all(
    spaces.map((space) => loadImage(session, space, tone)),
  );
  const fallback = largestFootprint(images);
  const floorGap = fallback.width * FLOOR_GAP_RATIO;

  return {
    title: spaceGroup.name,
    planes: spaces.map((space, index) => {
      const image = images[index];
      const { width, height } = image ?? fallback;
      return {
        id: space.id,
        rect: [0, 0, index * floorGap, width, height],
        image: image?.bitmap,
        text: space.name,
      };
    }),
  };
};
