import api, {
  ApiError,
  type AuthSession,
  type Space,
  type SpaceId,
} from "@/api";
import {
  type FloorPlanImage,
  type FloorPlanTone,
  toFloorPlanImage,
} from "@/lib/iso-scene/floor-plan-image";

interface CachedImage {
  version: number;
  image: Promise<FloorPlanImage | undefined>;
}

const caches: Record<FloorPlanTone, Map<SpaceId, CachedImage>> = {
  original: new Map(),
  blueprint: new Map(),
};

const release = (cached?: CachedImage) => {
  cached?.image.then((image) => image?.bitmap.close()).catch(() => {});
};

export const clearCache = (spaceId?: SpaceId) => {
  for (const cache of Object.values(caches)) {
    if (spaceId === undefined) {
      cache.forEach(release);
      cache.clear();
    } else {
      release(cache.get(spaceId));
      cache.delete(spaceId);
    }
  }
};

export const fetchFloorPlanImage = (
  s: AuthSession,
  space: Space,
  tone: FloorPlanTone,
): Promise<FloorPlanImage | undefined> => {
  const version = space.background_image_updated_at;
  if (!version) {
    return Promise.resolve(undefined);
  }

  const cache = caches[tone];
  const cached = cache.get(space.id);
  if (cached?.version === version) {
    return cached.image;
  }
  release(cached);

  const image = api.spaces.getSpaceBackgroundImage(s, space.id, version).then(
    (blob) => toFloorPlanImage(blob, tone),
    (err: unknown) => {
      if (err instanceof ApiError && err.status === 404) {
        return undefined;
      }
      throw err;
    },
  );
  cache.set(space.id, { version, image });
  image.catch(() => {
    if (cache.get(space.id)?.image === image) {
      cache.delete(space.id);
    }
  });

  return image;
};
