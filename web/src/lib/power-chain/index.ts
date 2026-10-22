import type {
  ChainSwitch,
  Device,
  DeviceId,
  Position,
  Ref,
  Space,
  SpaceId,
  Switch,
  SwitchId,
  SwitchPanel,
  SwitchPanelId,
} from "@/api";

export type ChainPlaceKey = string;

export type ChainPlaceTarget =
  | { kind: "panel"; panel: SwitchPanel }
  | { kind: "device"; device: Device };

export interface ChainPlace {
  key: ChainPlaceKey;
  text: string;
  space: Ref<SpaceId> | null;
  position: Position | null;
  shape: "square" | "circle";
  isFocus: boolean;
  target: ChainPlaceTarget;
}

export type ChainDetail =
  | { kind: "panel"; panel: SwitchPanel; switches: Switch[] }
  | { kind: "switch"; sw: Switch; panel: SwitchPanel; upstream?: Switch }
  | { kind: "device"; device: Device; feeders: Switch[] };

export interface ChainIndex {
  switches: Map<SwitchId, { sw: Switch; panel: SwitchPanel }>;
  panels: Map<SwitchPanelId, { panel: SwitchPanel; switches: Switch[] }>;
  devices: Map<DeviceId, { device: Device; feeders: Switch[] }>;
}

export type ChainNode =
  | { kind: "switch"; sw: Switch; panel: SwitchPanel }
  | { kind: "device"; device: Device };

export interface ChainTreeNode {
  key: string;
  node: ChainNode;
  isFocus: boolean;
  feedsSink: boolean;
  children: ChainTreeNode[];
}

export interface PowerChain {
  id: string;
  title: string;
  focusSpaceId: SpaceId;
  roots: ChainTreeNode[];
  sink?: ChainNode;
  places: ChainPlace[];
  edges: [downstream: ChainPlaceKey, upstream: ChainPlaceKey][];
}

export interface SwitchForestOptions {
  focusSwitchId?: SwitchId;
  feedsSink?: (sw: Switch) => boolean;
  leaves?: (sw: Switch) => ChainTreeNode[];
}

export const panelKey = (id: SwitchPanelId): ChainPlaceKey => `panel:${id}`;

export const deviceKey = (id: DeviceId): ChainPlaceKey => `device:${id}`;

export const panelPlace = (
  panel: SwitchPanel,
  isFocus: boolean,
): ChainPlace => ({
  key: panelKey(panel.id),
  text: panel.name,
  space: panel.space,
  position: panel.position,
  shape: "square",
  isFocus,
  target: { kind: "panel", panel },
});

export const devicePlace = (device: Device, isFocus: boolean): ChainPlace => ({
  key: deviceKey(device.id),
  text: device.name,
  space: device.space,
  position: device.position,
  shape: "circle",
  isFocus,
  target: { kind: "device", device },
});

export const deviceLeaf = (
  device: Device,
  feeder: SwitchId,
): ChainTreeNode => ({
  key: `device:${device.id}@${feeder}`,
  node: { kind: "device", device },
  isFocus: false,
  feedsSink: false,
  children: [],
});

export const panelsOf = (switches: ChainSwitch[]): SwitchPanel[] => [
  ...new Map(
    switches.map(({ switch_panel }) => [switch_panel.id, switch_panel]),
  ).values(),
];

export const buildSwitchForest = (
  switches: ChainSwitch[],
  { focusSwitchId, feedsSink, leaves }: SwitchForestOptions = {},
): ChainTreeNode[] => {
  const nodes = new Map<SwitchId, ChainTreeNode>(
    switches.map((sw) => [
      sw.id,
      {
        key: `switch:${sw.id}`,
        node: { kind: "switch", sw, panel: sw.switch_panel },
        isFocus: sw.id === focusSwitchId,
        feedsSink: feedsSink?.(sw) ?? false,
        children: [],
      },
    ]),
  );

  const roots: ChainTreeNode[] = [];
  for (const sw of switches) {
    const node = nodes.get(sw.id);
    if (!node) {
      continue;
    }
    const upstreamId = sw.upstream_switch?.id;
    const parent = upstreamId === undefined ? undefined : nodes.get(upstreamId);
    if (parent) {
      parent.children.push(node);
    } else {
      roots.push(node);
    }
  }
  for (const sw of switches) {
    nodes.get(sw.id)?.children.push(...(leaves?.(sw) ?? []));
  }
  return roots;
};

export const collectEdges = (
  switches: Switch[],
  deviceFeeds: [DeviceId, SwitchId][],
): PowerChain["edges"] => {
  const panelOf = new Map(
    switches.map(({ id, switch_panel }) => [
      id,
      switch_panel ? panelKey(switch_panel.id) : undefined,
    ]),
  );
  const edges = new Map<string, [ChainPlaceKey, ChainPlaceKey]>();
  const addEdge = (
    from: ChainPlaceKey | undefined,
    to: ChainPlaceKey | undefined,
  ) => {
    if (from !== undefined && to !== undefined && from !== to) {
      edges.set(`${from}>${to}`, [from, to]);
    }
  };

  for (const sw of switches) {
    if (sw.upstream_switch) {
      addEdge(panelOf.get(sw.id), panelOf.get(sw.upstream_switch.id));
    }
  }
  for (const [device, sw] of deviceFeeds) {
    addEdge(deviceKey(device), panelOf.get(sw));
  }
  return [...edges.values()];
};

export const nodeSpaceId = (node: ChainNode): SpaceId | undefined =>
  node.kind === "switch" ? node.panel.space?.id : node.device.space?.id;

export const isNodePlaced = (node: ChainNode): boolean =>
  node.kind === "switch"
    ? node.panel.position != null
    : node.device.position != null;

export const countPlacesBySpace = ({
  places,
}: PowerChain): Map<SpaceId, number> => {
  const counts = new Map<SpaceId, number>();
  for (const { space, isFocus } of places) {
    if (space && !isFocus) {
      counts.set(space.id, (counts.get(space.id) ?? 0) + 1);
    }
  }
  return counts;
};

export const floorGroups = (
  chain: PowerChain,
  floors: Space[],
): Map<SpaceId, number> => {
  const counts = countPlacesBySpace(chain);
  return new Map(
    floors
      .filter(({ id }) => counts.has(id))
      .map(({ id }, group) => [id, group]),
  );
};

export const spannedSpaces = (chain: PowerChain, spaces: Space[]): Space[] => {
  const involved = new Set([
    chain.focusSpaceId,
    ...chain.places.flatMap(({ space }) => (space ? [space.id] : [])),
  ]);
  const indices = spaces.flatMap(({ id }, index) =>
    involved.has(id) ? [index] : [],
  );
  if (indices.length === 0) {
    return [];
  }
  return spaces.slice(Math.min(...indices), Math.max(...indices) + 1);
};

export const indexChain = ({ roots, sink }: PowerChain): ChainIndex => {
  const index: ChainIndex = {
    switches: new Map(),
    panels: new Map(),
    devices: new Map(),
  };
  const addFeeder = (device: Device, feeder: Switch) => {
    const entry = index.devices.get(device.id) ?? { device, feeders: [] };
    entry.feeders.push(feeder);
    index.devices.set(device.id, entry);
  };

  const visit = ({ node, feedsSink, children }: ChainTreeNode) => {
    if (node.kind !== "switch") {
      return;
    }
    const { sw, panel } = node;
    index.switches.set(sw.id, { sw, panel });
    const entry = index.panels.get(panel.id) ?? { panel, switches: [] };
    entry.switches.push(sw);
    index.panels.set(panel.id, entry);
    if (feedsSink && sink?.kind === "device") {
      addFeeder(sink.device, sw);
    }
    for (const child of children) {
      if (child.node.kind === "device") {
        addFeeder(child.node.device, sw);
      } else {
        visit(child);
      }
    }
  };
  roots.forEach(visit);
  return index;
};

export const describeNode = (
  index: ChainIndex,
  node: ChainNode,
): ChainDetail => {
  if (node.kind === "device") {
    return {
      kind: "device",
      device: node.device,
      feeders: index.devices.get(node.device.id)?.feeders ?? [],
    };
  }
  const upstreamId = node.sw.upstream_switch?.id;
  return {
    kind: "switch",
    sw: node.sw,
    panel: node.panel,
    upstream:
      upstreamId === undefined ? undefined : index.switches.get(upstreamId)?.sw,
  };
};

export const describePlace = (
  index: ChainIndex,
  { target }: ChainPlace,
): ChainDetail =>
  target.kind === "device"
    ? describeNode(index, target)
    : {
        kind: "panel",
        panel: target.panel,
        switches: index.panels.get(target.panel.id)?.switches ?? [],
      };

export const detailPlaceKey = (detail: ChainDetail): ChainPlaceKey =>
  detail.kind === "device"
    ? deviceKey(detail.device.id)
    : panelKey(detail.panel.id);

export const detailSpaceId = (detail: ChainDetail): SpaceId | undefined =>
  detail.kind === "device" ? detail.device.space?.id : detail.panel.space?.id;
