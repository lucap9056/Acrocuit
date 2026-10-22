import type {
  AuthSession,
  Device,
  DeviceId,
  SpaceId,
  Switch,
  SwitchId,
} from "@/api";
import api from "@/api";
import {
  buildSwitchForest,
  type ChainTreeNode,
  collectEdges,
  deviceLeaf,
  devicePlace,
  type PowerChain,
  panelPlace,
  panelsOf,
} from "./index";

export const fetchDownstreamChain = async (
  session: AuthSession,
  sw: Switch,
  fallbackSpaceId: SpaceId,
): Promise<PowerChain> => {
  const { switches, devices } = await api.switches.getSwitchDownstream(
    session,
    sw.id,
  );

  const uniqueDevices = new Map<DeviceId, Device>(
    devices.map((device) => [device.id, device]),
  );
  const feeds = devices.flatMap(({ id, switch: feeder }) =>
    feeder ? [[id, feeder.id] as [DeviceId, SwitchId]] : [],
  );
  const leaves = new Map<SwitchId, ChainTreeNode[]>();
  for (const device of devices) {
    if (device.switch) {
      const feeder = device.switch.id;
      leaves.set(feeder, [
        ...(leaves.get(feeder) ?? []),
        deviceLeaf(device, feeder),
      ]);
    }
  }

  const rootPanel = switches.find(({ id }) => id === sw.id)?.switch_panel;

  return {
    id: `downstream:${sw.id}`,
    title: sw.name,
    focusSpaceId: rootPanel?.space?.id ?? fallbackSpaceId,
    roots: buildSwitchForest(switches, {
      focusSwitchId: sw.id,
      leaves: (target) => leaves.get(target.id) ?? [],
    }),
    places: [
      ...panelsOf(switches).map((panel) =>
        panelPlace(panel, panel.id === rootPanel?.id),
      ),
      ...[...uniqueDevices.values()].map((device) =>
        devicePlace(device, false),
      ),
    ],
    edges: collectEdges(switches, feeds),
  };
};
