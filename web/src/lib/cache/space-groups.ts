import type { AuthSession, SpaceGroup } from "@/api";
import api from "@/api";
import { clearSpaceGroupCache } from "@/lib/cache/spaces";

let cache: Promise<SpaceGroup[]> | undefined;

export const clearCache = () => {
  cache = undefined;
};

const updateCache = (sgs: SpaceGroup[]) => {
  cache = Promise.resolve(sgs);
  return sgs;
};

export const fetchSpaceGroups = (s: AuthSession) => {
  if (cache === undefined) {
    cache = api.spaceGroups.getSpaceGroups(s);
  }
  return cache;
};

export const addSpaceGroup = (s: AuthSession, name: string) =>
  Promise.all([
    fetchSpaceGroups(s),
    api.spaceGroups.createSpaceGroup(s, name),
  ]).then(([spaceGroups, sg]) => {
    if (spaceGroups.find((s) => s.id === sg.id)) {
      return spaceGroups;
    } else {
      const updatedSpaceGroups = [...spaceGroups, sg];
      return updateCache(updatedSpaceGroups);
    }
  });

export const updateSpaceGroup = (s: AuthSession, sg: SpaceGroup) =>
  Promise.all([
    fetchSpaceGroups(s),
    api.spaceGroups.updateSpaceGroupName(s, sg.id, sg.name),
  ]).then(([spaceGroups, updated]) => {
    const updatedSpaceGroups = spaceGroups.map((g) =>
      g.id === sg.id ? updated : g,
    );

    return updateCache(updatedSpaceGroups);
  });

const dropSpaceGroup = (
  s: AuthSession,
  sg: SpaceGroup,
  request: Promise<unknown>,
) =>
  Promise.all([fetchSpaceGroups(s), request]).then(([spaceGroups]) => {
    const updatedSpaceGroups = spaceGroups.filter((g) => g.id !== sg.id);

    clearSpaceGroupCache(sg.id);

    return updateCache(updatedSpaceGroups);
  });

export const removeSpaceGroup = (s: AuthSession, sg: SpaceGroup) =>
  dropSpaceGroup(s, sg, api.spaceGroups.deleteSpaceGroup(s, sg.id));

export const leaveSpaceGroup = (s: AuthSession, sg: SpaceGroup) =>
  dropSpaceGroup(s, sg, api.spaceGroups.leaveSpaceGroup(s, sg.id));

export const transferSpaceGroupOwner = (
  s: AuthSession,
  sg: SpaceGroup,
  email: string,
) =>
  Promise.all([
    fetchSpaceGroups(s),
    api.spaceGroups.transferOwner(s, sg.id, email),
  ]).then(([spaceGroups]) => {
    const updatedSpaceGroups = spaceGroups.map((g): SpaceGroup =>
      g.id === sg.id ? { ...g, role: "editor" } : g,
    );

    return updateCache(updatedSpaceGroups);
  });
