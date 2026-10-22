import type { AuthSession, Ref, Space, SpaceGroupId } from "@/api";
import api from "@/api";
import { clearCache as clearSwitchPanelsCache } from "@/lib/cache/switch-panels";
import { clearCache as clearSwitchesCache } from "@/lib/cache/switches";
import { clearCache as clearDevicesCache } from "@/lib/cache/devices";
import { clearCache as clearDeviceSwitchesCache } from "@/lib/cache/device-switches";
import { clearCache as clearSpaceImagesCache } from "@/lib/cache/space-images";

type CompleteSpace = Space & { space_group: Ref<SpaceGroupId> };

const cache: Map<SpaceGroupId, Promise<Space[]>> = new Map();

export const clearCache = (sg?: SpaceGroupId) => {
  if (sg) {
    cache.delete(sg);
  } else {
    cache.clear();
  }
};

export const clearSpaceGroupCache = (sg: SpaceGroupId) => {
  const cached = cache.get(sg);
  cache.delete(sg);

  clearSwitchPanelsCache();
  clearSwitchesCache();
  clearDevicesCache();
  clearDeviceSwitchesCache();
  cached
    ?.then((spaces) => {
      for (const space of spaces) {
        clearSpaceImagesCache(space.id);
      }
    })
    .catch(() => {});
};

const updateCache = (sg: SpaceGroupId, s: Space[]) => {
  cache.set(sg, Promise.resolve(s));
  return s;
};

export const fetchSpaces = (s: AuthSession, sg: SpaceGroupId) => {
  return (
    cache.get(sg) ||
    (() => {
      const r = api.spaceGroups.getSpaces(s, sg);
      cache.set(sg, r);
      return r;
    })()
  );
};

export const addSpace = (s: AuthSession, sg: SpaceGroupId, name: string) =>
  Promise.all([fetchSpaces(s, sg), api.spaceGroups.addSpace(s, sg, name)]).then(
    ([spaces, space]) => {
      if (spaces.find((sp) => sp.id === space.id)) {
        return spaces;
      }

      const updatedSpaces = [...spaces, space];
      return updateCache(sg, updatedSpaces);
    },
  );

export const moveSpace = (s: AuthSession, space: CompleteSpace) => {
  const sg = space.space_group.id;

  return api.spaces
    .updateSpaceOrder(s, space.id, space.display_order)
    .then((spaces) => updateCache(sg, spaces));
};

export const removeSpace = (s: AuthSession, space: CompleteSpace) =>
  Promise.all([
    fetchSpaces(s, space.space_group.id),
    api.spaces.deleteSpace(s, space.id),
  ]).then(([spaces]) => {
    const updatedSpaces = spaces.filter((sp) => sp.id !== space.id);

    clearSwitchPanelsCache(space.id);
    clearDevicesCache();
    clearDeviceSwitchesCache();
    clearSpaceImagesCache(space.id);
    clearSwitchesCache();

    return updateCache(space.space_group.id, updatedSpaces);
  });

export const renameSpace = (
  s: AuthSession,
  space: CompleteSpace,
  name: string,
) =>
  Promise.all([
    fetchSpaces(s, space.space_group.id),
    api.spaces.updateSpaceName(s, space.id, name),
  ]).then(([spaces, updatedSpace]) => {
    const updatedSpaces = spaces.map((sp) =>
      sp.id === space.id ? updatedSpace : sp,
    );

    return updateCache(space.space_group.id, updatedSpaces);
  });

const replaceFloorPlan = (
  s: AuthSession,
  space: CompleteSpace,
  request: Promise<Space>,
) =>
  Promise.all([fetchSpaces(s, space.space_group.id), request]).then(
    ([spaces, updatedSpace]) => {
      const updatedSpaces = spaces.map((sp) =>
        sp.id === space.id ? updatedSpace : sp,
      );

      clearSpaceImagesCache(space.id);

      return updateCache(space.space_group.id, updatedSpaces);
    },
  );

export const uploadFloorPlan = (
  s: AuthSession,
  space: CompleteSpace,
  image: Blob,
) =>
  replaceFloorPlan(
    s,
    space,
    api.spaces.setSpaceBackgroundImage(s, space.id, image),
  );

export const removeFloorPlan = (s: AuthSession, space: CompleteSpace) =>
  replaceFloorPlan(
    s,
    space,
    api.spaces.deleteSpaceBackgroundImage(s, space.id),
  );
