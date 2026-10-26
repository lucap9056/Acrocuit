import type { AuthSession, Device, SpaceId } from "@/api";
import api from "@/api";
import {
  buildSwitchForest,
  collectEdges,
  devicePlace,
  type PowerChain,
  panelPlace,
  panelsOf,
} from "./index";

export const fetchUpstreamChain = async (
  session: AuthSession,
  device: Device,
  fallbackSpaceId: SpaceId,
): Promise<PowerChain> => {
  const { switches, switch_links, device_links } =
    await api.devices.getDeviceUpstream(session, device.id);
  const space = device.space ?? { id: fallbackSpaceId };
  const target = { ...device, space };
  const directIds = new Set(device_links.map(({ switch_id }) => switch_id));

  return {
    id: `upstream:${device.id}`,
    direction: "upstream",
    supply: { voltage: Number.NaN, phase: undefined },
    title: device.name,
    focusSpaceId: space.id,
    roots: buildSwitchForest(switches, switch_links, {
      feedsSink: (sw) => directIds.has(sw.id),
    }),
    sink: { kind: "device", device: target },
    sinkLinks: new Map(
      device_links.map(({ switch_id, attributes }) => [switch_id, attributes]),
    ),
    places: [
      devicePlace(target, true),
      ...panelsOf(switches).map((panel) => panelPlace(panel, false)),
    ],
    edges: collectEdges(switches, switch_links, device_links),
  };
};
