import type { AuthSession, BreakerGroup, Position, Ref, SpaceId } from "@/api";
import api from "@/api";
import { clearCache as clearBreakersCache } from "@/lib/cache/breakers";
import { createListCache } from "@/lib/cache/list-cache";

type CompleteBreakerGroup = BreakerGroup & { space: Ref<SpaceId> };

const cache = createListCache<SpaceId, BreakerGroup>(
  api.spaces.getBreakerGroups,
);

export const clearCache = cache.clear;

export const fetchBreakerGroups = cache.fetch;

export const addBreakerGroup = (
  s: AuthSession,
  space: SpaceId,
  name: string,
  position?: Position,
) => cache.add(s, space, api.spaces.addBreakerGroup(s, space, name, position));

export const renameBreakerGroup = (
  s: AuthSession,
  bg: CompleteBreakerGroup,
  name: string,
) =>
  cache.replace(
    s,
    bg.space.id,
    api.breakerGroups.updateBreakerGroupName(s, bg.id, name),
  );

export const setBreakerGroupPosition = (
  s: AuthSession,
  bg: CompleteBreakerGroup,
  position: Position,
) =>
  cache.replace(
    s,
    bg.space.id,
    api.breakerGroups.updateBreakerGroupPosition(s, bg.id, position),
  );

export const removeBreakerGroup = (s: AuthSession, bg: CompleteBreakerGroup) =>
  cache
    .remove(
      s,
      bg.space.id,
      bg.id,
      api.breakerGroups.deleteBreakerGroup(s, bg.id),
    )
    .then((breakerGroups) => {
      clearBreakersCache();
      return breakerGroups;
    });
