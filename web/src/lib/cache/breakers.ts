import type {
  AuthSession,
  Breaker,
  BreakerGroupId,
  BreakerId,
  Ref,
} from "@/api";
import api from "@/api";
import { createListCache } from "@/lib/cache/list-cache";

type CompleteBreaker = Breaker & { breaker_group: Ref<BreakerGroupId> };

const cache = createListCache<BreakerGroupId, Breaker>(
  api.breakerGroups.getBreakers,
);

export const clearCache = cache.clear;

export const fetchBreakers = cache.fetch;

export const addBreaker = (
  s: AuthSession,
  bg: BreakerGroupId,
  name: string,
  upstream?: BreakerId,
) => cache.add(s, bg, api.breakerGroups.addBreaker(s, bg, name, upstream));

export const renameBreaker = (
  s: AuthSession,
  breaker: CompleteBreaker,
  name: string,
) =>
  cache.replace(
    s,
    breaker.breaker_group.id,
    api.breakers.updateBreaker(s, breaker.id, { name }),
  );

export const setBreakerUpstream = (
  s: AuthSession,
  breaker: CompleteBreaker,
  upstream: BreakerId | null,
) =>
  cache.replace(
    s,
    breaker.breaker_group.id,
    api.breakers.updateBreakerUpstream(s, breaker.id, upstream),
  );

export const moveBreaker = (s: AuthSession, breaker: CompleteBreaker) => {
  const bg = breaker.breaker_group.id;

  return api.breakers
    .updateBreakerOrder(s, breaker.id, breaker.display_order)
    .then(() => {
      clearCache(bg);
      return fetchBreakers(s, bg);
    });
};

export const removeBreaker = (s: AuthSession, breaker: CompleteBreaker) =>
  api.breakers.deleteBreaker(s, breaker.id).then(() => {
    clearCache();
    return fetchBreakers(s, breaker.breaker_group.id);
  });
