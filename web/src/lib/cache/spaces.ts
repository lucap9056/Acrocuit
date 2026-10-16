import type { AuthSession, Ref, Space, SpaceGroupId } from "@/api";
import api from "@/api";

type CompleteSpace = Space & { space_group: Ref<SpaceGroupId> };

const cache: Map<SpaceGroupId, Promise<Space[]>> = new Map();

export const clearCache = (sg?: SpaceGroupId) => {
  if (sg) {
    cache.delete(sg);
  } else {
    cache.clear();
  }
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
    .then(() => {
      clearCache(sg);
      return fetchSpaces(s, sg);
    });
};

export const removeSpace = (s: AuthSession, space: CompleteSpace) =>
  Promise.all([
    fetchSpaces(s, space.space_group.id),
    api.spaces.deleteSpace(s, space.id),
  ]).then(([spaces]) => {
    const updatedSpaces = spaces.filter((sp) => sp.id !== space.id);

    return updateCache(space.space_group.id, updatedSpaces);
  });

export const renameSpace = (
  s: AuthSession,
  space: CompleteSpace,
  name: string,
) =>
  Promise.all([
    fetchSpaces(s, space.space_group.id),
    api.spaces.updateSpace(s, space.id, { name }),
  ]).then(([spaces, updatedSpace]) => {
    const updatedSpaces = spaces.map((sp) =>
      sp.id === space.id ? updatedSpace : sp,
    );

    return updateCache(space.space_group.id, updatedSpaces);
  });
