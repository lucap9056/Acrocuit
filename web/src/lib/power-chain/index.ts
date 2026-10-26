import type {
  Attributes,
  ChainSwitch,
  Device,
  DeviceId,
  DeviceLink,
  Position,
  Ref,
  Space,
  SpaceId,
  Switch,
  SwitchId,
  SwitchLink,
  SwitchPanel,
  SwitchPanelId,
} from "@/api";
import type { Supply } from "./load";

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
  | { kind: "switch"; sw: Switch; panel: SwitchPanel; upstreams: Switch[] }
  | { kind: "device"; device: Device; feeders: Switch[] }
  | { kind: "link"; link: ChainLink };

export interface ChainIndex {
  switches: Map<
    SwitchId,
    { sw: Switch; panel: SwitchPanel; upstreams: Switch[] }
  >;
  panels: Map<SwitchPanelId, { panel: SwitchPanel; switches: Switch[] }>;
  devices: Map<DeviceId, { device: Device; feeders: Switch[] }>;
}

export type ChainNode =
  | { kind: "switch"; sw: Switch; panel: SwitchPanel }
  | { kind: "device"; device: Device };

export interface ChainLink {
  key: string;
  upstream: { sw: Switch; panel: SwitchPanel };
  downstream: ChainNode;
  attributes: Attributes | null;
}

export interface ChainTreeNode {
  key: string;
  node: ChainNode;
  isFocus: boolean;
  feedsSink: boolean;
  link: Attributes | null;
  children: ChainTreeNode[];
}

export interface PowerChain {
  id: string;
  direction: "upstream" | "downstream";
  supply: Supply;
  title: string;
  focusSpaceId: SpaceId;
  roots: ChainTreeNode[];
  sink?: ChainNode;
  sinkLinks?: Map<SwitchId, Attributes | null>;
  places: ChainPlace[];
  edges: [downstream: ChainPlaceKey, upstream: ChainPlaceKey][];
}

export interface SwitchForestOptions {
  focusSwitchId?: SwitchId;
  feedsSink?: (sw: Switch) => boolean;
  leaves?: (sw: Switch, parentKey: string) => ChainTreeNode[];
}

export const panelKey = (id: SwitchPanelId): ChainPlaceKey => `panel:${id}`;

export const deviceKey = (id: DeviceId): ChainPlaceKey => `device:${id}`;

export const chainNodeKey = (node: ChainNode) =>
  node.kind === "switch" ? `switch:${node.sw.id}` : `device:${node.device.id}`;

export const chainLinkKey = (upstream: SwitchId, node: ChainNode) =>
  `switch:${upstream}>${chainNodeKey(node)}`;

export const chainLink = (
  upstream: ChainNode,
  downstream: ChainNode,
  attributes: Attributes | null,
): ChainLink | undefined =>
  upstream.kind === "switch"
    ? {
        key: chainLinkKey(upstream.sw.id, downstream),
        upstream: { sw: upstream.sw, panel: upstream.panel },
        downstream,
        attributes,
      }
    : undefined;

export const treeLinkKey = ({ key }: ChainTreeNode) =>
  key.split(">").slice(-2).join(">");

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
  parentKey: string,
  link: Attributes | null,
): ChainTreeNode => ({
  key: `${parentKey}>device:${device.id}`,
  node: { kind: "device", device },
  isFocus: false,
  feedsSink: false,
  link,
  children: [],
});

export const panelsOf = (switches: ChainSwitch[]): SwitchPanel[] => [
  ...new Map(
    switches.map(({ switch_panel }) => [switch_panel.id, switch_panel]),
  ).values(),
];

export const buildSwitchForest = (
  switches: ChainSwitch[],
  links: SwitchLink[],
  { focusSwitchId, feedsSink, leaves }: SwitchForestOptions = {},
): ChainTreeNode[] => {
  const byId = new Map(switches.map((sw) => [sw.id, sw]));
  const downstreams = new Map<
    SwitchId,
    { sw: ChainSwitch; link: Attributes | null }[]
  >();
  const fed = new Set<SwitchId>();
  for (const {
    upstream_switch_id,
    downstream_switch_id,
    attributes,
  } of links) {
    const downstream = byId.get(downstream_switch_id);
    if (downstream && byId.has(upstream_switch_id)) {
      downstreams.set(upstream_switch_id, [
        ...(downstreams.get(upstream_switch_id) ?? []),
        { sw: downstream, link: attributes },
      ]);
      fed.add(downstream_switch_id);
    }
  }

  const grow = (
    sw: ChainSwitch,
    parentKey?: string,
    link: Attributes | null = null,
  ): ChainTreeNode => {
    const key =
      parentKey === undefined
        ? `switch:${sw.id}`
        : `${parentKey}>switch:${sw.id}`;
    return {
      key,
      node: { kind: "switch", sw, panel: sw.switch_panel },
      isFocus: sw.id === focusSwitchId,
      feedsSink: feedsSink?.(sw) ?? false,
      link,
      children: [
        ...(downstreams.get(sw.id) ?? []).map((child) =>
          grow(child.sw, key, child.link),
        ),
        ...(leaves?.(sw, key) ?? []),
      ],
    };
  };

  return switches.filter(({ id }) => !fed.has(id)).map((sw) => grow(sw));
};

export const collectEdges = (
  switches: Switch[],
  switchLinks: SwitchLink[],
  deviceLinks: DeviceLink[],
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

  for (const { upstream_switch_id, downstream_switch_id } of switchLinks) {
    addEdge(panelOf.get(downstream_switch_id), panelOf.get(upstream_switch_id));
  }
  for (const { device_id, switch_id } of deviceLinks) {
    addEdge(deviceKey(device_id), panelOf.get(switch_id));
  }
  return [...edges.values()];
};

export interface ChainAttachment {
  items: Set<string>;
  nodes: Set<string>;
  places: Set<ChainPlaceKey>;
  edges: Set<string>;
}

const nodePlaceKey = (node: ChainNode): ChainPlaceKey =>
  node.kind === "switch" ? panelKey(node.panel.id) : deviceKey(node.device.id);

export const attachChain = (
  roots: ChainTreeNode[],
  isEnabled: (node: ChainNode) => boolean,
): ChainAttachment => {
  const attachment: ChainAttachment = {
    items: new Set(),
    nodes: new Set(),
    places: new Set(),
    edges: new Set(),
  };

  const visit = (item: ChainTreeNode, parentPlace?: ChainPlaceKey) => {
    if (!item.isFocus && !isEnabled(item.node)) {
      return;
    }
    const place = nodePlaceKey(item.node);
    attachment.items.add(item.key);
    attachment.nodes.add(chainNodeKey(item.node));
    attachment.places.add(place);
    if (parentPlace !== undefined && parentPlace !== place) {
      attachment.edges.add(`${place}>${parentPlace}`);
    }
    for (const child of item.children) {
      visit(child, place);
    }
  };

  for (const root of roots) {
    visit(root);
  }
  return attachment;
};

export const withAttachment = (
  chain: PowerChain,
  attachment: ChainAttachment,
): PowerChain => ({
  ...chain,
  places: chain.places.filter(({ key }) => attachment.places.has(key)),
  edges: chain.edges.filter(([downstream, upstream]) =>
    attachment.edges.has(`${downstream}>${upstream}`),
  ),
});

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
  const addUnique = <T>(items: T[], item: T) => {
    if (!items.includes(item)) {
      items.push(item);
    }
  };
  const addFeeder = (device: Device, feeder: Switch) => {
    const entry = index.devices.get(device.id) ?? { device, feeders: [] };
    addUnique(entry.feeders, feeder);
    index.devices.set(device.id, entry);
  };

  const visit = (
    { node, feedsSink, children }: ChainTreeNode,
    parent?: Switch,
  ) => {
    if (node.kind === "device") {
      if (parent) {
        addFeeder(node.device, parent);
      }
      return;
    }
    const { sw, panel } = node;
    const entry = index.switches.get(sw.id) ?? { sw, panel, upstreams: [] };
    if (parent) {
      addUnique(entry.upstreams, parent);
    }
    index.switches.set(sw.id, entry);
    const panelEntry = index.panels.get(panel.id) ?? { panel, switches: [] };
    addUnique(panelEntry.switches, sw);
    index.panels.set(panel.id, panelEntry);
    if (feedsSink && sink?.kind === "device") {
      addFeeder(sink.device, sw);
    }
    for (const child of children) {
      visit(child, sw);
    }
  };
  for (const root of roots) {
    visit(root);
  }
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
  return {
    kind: "switch",
    sw: node.sw,
    panel: node.panel,
    upstreams: index.switches.get(node.sw.id)?.upstreams ?? [],
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

export const detailPlaceKey = (
  detail: ChainDetail,
): ChainPlaceKey | undefined => {
  switch (detail.kind) {
    case "device":
      return deviceKey(detail.device.id);
    case "link":
      return undefined;
    default:
      return panelKey(detail.panel.id);
  }
};

export const detailSpaceId = (detail: ChainDetail): SpaceId | undefined => {
  switch (detail.kind) {
    case "device":
      return detail.device.space?.id;
    case "link":
      return undefined;
    default:
      return detail.panel.space?.id;
  }
};
