import type {
  Attributes,
  AuthSession,
  Device,
  DeviceId,
  SpaceId,
  Switch,
  SwitchId,
} from "@/api";
import api from "@/api";
import { fetchSwitchUpstreams } from "@/lib/cache/switch-upstreams";
import {
  buildSwitchForest,
  collectEdges,
  deviceLeaf,
  devicePlace,
  type PowerChain,
  panelPlace,
  panelsOf,
} from "./index";
import { type Supply, switchPhase, switchVoltage } from "./load";

const isSupplyComplete = ({ voltage, phase }: Supply) =>
  Number.isFinite(voltage) && phase !== undefined;

const fetchSupply = async (
  session: AuthSession,
  root: Switch,
): Promise<Supply> => {
  const supply: Supply = {
    voltage: switchVoltage(root),
    phase: switchPhase(root),
  };
  const visited = new Set([root.id]);
  let frontier = [root.id];
  while (frontier.length > 0 && !isSupplyComplete(supply)) {
    const upstreams = (
      await Promise.all(frontier.map((id) => fetchSwitchUpstreams(session, id)))
    )
      .flat()
      .filter(({ id }) => !visited.has(id));
    if (!Number.isFinite(supply.voltage)) {
      supply.voltage =
        upstreams.map(switchVoltage).find(Number.isFinite) ?? Number.NaN;
    }
    supply.phase ??= upstreams
      .map(switchPhase)
      .find((phase) => phase !== undefined);
    frontier = [...new Set(upstreams.map(({ id }) => id))];
    for (const id of frontier) {
      visited.add(id);
    }
  }
  return supply;
};

export const fetchDownstreamChain = async (
  session: AuthSession,
  sw: Switch,
  fallbackSpaceId: SpaceId,
): Promise<PowerChain> => {
  const { switches, switch_links, devices, device_links } =
    await api.switches.getSwitchDownstream(session, sw.id);

  const devicesById = new Map<DeviceId, Device>(
    devices.map((device) => [device.id, device]),
  );
  const fedDevices = new Map<
    SwitchId,
    { device: Device; link: Attributes | null }[]
  >();
  for (const { device_id, switch_id, attributes } of device_links) {
    const device = devicesById.get(device_id);
    if (device) {
      fedDevices.set(switch_id, [
        ...(fedDevices.get(switch_id) ?? []),
        { device, link: attributes },
      ]);
    }
  }

  const root = switches.find(({ id }) => id === sw.id);
  const rootPanel = root?.switch_panel;
  const supply = root
    ? await fetchSupply(session, root)
    : { voltage: Number.NaN, phase: undefined };

  return {
    id: `downstream:${sw.id}`,
    direction: "downstream",
    supply,
    title: sw.name,
    focusSpaceId: rootPanel?.space?.id ?? fallbackSpaceId,
    roots: buildSwitchForest(switches, switch_links, {
      focusSwitchId: sw.id,
      leaves: (target, parentKey) =>
        (fedDevices.get(target.id) ?? []).map(({ device, link }) =>
          deviceLeaf(device, parentKey, link),
        ),
    }),
    places: [
      ...panelsOf(switches).map((panel) =>
        panelPlace(panel, panel.id === rootPanel?.id),
      ),
      ...devices.map((device) => devicePlace(device, false)),
    ],
    edges: collectEdges(switches, switch_links, device_links),
  };
};
