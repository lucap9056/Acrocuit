import type { Attributes, AuthSession, SwitchId } from "@/api";
import api from "@/api";
import { createListCache } from "@/lib/cache/list-cache";

const cache = createListCache(api.switches.getSwitchUpstreams);

export const clearCache = cache.clear;

export const fetchSwitchUpstreams = cache.fetch;

export const addSwitchUpstream = (
  s: AuthSession,
  sw: SwitchId,
  upstream: SwitchId,
) =>
  api.switches.addSwitchUpstream(s, sw, upstream).then(() => {
    cache.clear(sw);
    return cache.fetch(s, sw);
  });

export const setSwitchUpstreamAttributes = (
  s: AuthSession,
  sw: SwitchId,
  upstream: SwitchId,
  attributes: Attributes,
) =>
  Promise.all([
    cache.fetch(s, sw),
    api.switches.updateSwitchUpstream(s, sw, upstream, attributes),
  ]).then(([items, link]) =>
    cache.set(
      sw,
      items.map((i) =>
        i.id === upstream ? { ...i, link: { attributes: link.attributes } } : i,
      ),
    ),
  );

export const removeSwitchUpstream = (
  s: AuthSession,
  sw: SwitchId,
  upstream: SwitchId,
) =>
  cache.remove(
    s,
    sw,
    upstream,
    api.switches.deleteSwitchUpstream(s, sw, upstream),
  );
