import type { AuthSession, Device, SpaceId } from "@/api";
import api from "@/api";
import { fetchDeviceSwitches } from "@/lib/cache/device-switches";
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
  const [switches, directSwitches] = await Promise.all([
    api.devices.getDeviceUpstream(session, device.id),
    fetchDeviceSwitches(session, device.id),
  ]);
  const space = device.space ?? { id: fallbackSpaceId };
  const target = { ...device, space };
  const directIds = new Set(directSwitches.map(({ id }) => id));

  return {
    id: `upstream:${device.id}`,
    title: device.name,
    focusSpaceId: space.id,
    roots: buildSwitchForest(switches, {
      feedsSink: (sw) => directIds.has(sw.id),
    }),
    sink: { kind: "device", device: target },
    places: [
      devicePlace(target, true),
      ...panelsOf(switches).map((panel) => panelPlace(panel, false)),
    ],
    edges: collectEdges(
      switches,
      [...directIds].map((sw) => [device.id, sw]),
    ),
  };
};
