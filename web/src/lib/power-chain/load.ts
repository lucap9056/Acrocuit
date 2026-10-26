import type { Attributes, Device, DeviceId, Switch, SwitchId } from "@/api";
import { attributePhase, type Phase } from "@/lib/attributes";
import { parsePowerFactor, parseQuantity } from "@/lib/power";
import {
  type ChainNode,
  type ChainTreeNode,
  chainLinkKey,
  chainNodeKey,
} from "./index";

export interface Supply {
  voltage: number;
  phase: Phase | undefined;
}

export interface SwitchLoad extends Supply {
  power: number;
  apparent: number;
  current: number;
  limit: number;
  unknown: number;
  isVoltageInherited: boolean;
}

export interface LinkLoad {
  current: number;
  ampacity: number;
  voltageDrop: number;
}

export interface NodeVoltageDrop {
  drop: number;
  voltage: number;
  isPartial: boolean;
}

export interface ChainLoads {
  switches: Map<SwitchId, SwitchLoad>;
  links: Map<string, LinkLoad>;
  drops: Map<string, NodeVoltageDrop>;
}

interface Demand {
  power: number;
  apparent: number;
  current: number;
  voltage: number;
}

const COPPER_RESISTIVITY_OHM_MM2_PER_M = 0.0178;

export const VOLTAGE_DROP_LIMIT = 0.05;

const attributeText = (attributes: Attributes | null, key: string) => {
  const value = attributes?.[key];
  return typeof value === "number" || typeof value === "string"
    ? String(value)
    : "";
};

const attributeNumber = (attributes: Attributes | null, key: string) =>
  parseQuantity(attributeText(attributes, key));

export const switchVoltage = (sw: Switch) =>
  attributeNumber(sw.attributes, "rated_voltage");

export const switchPhase = (sw: Switch) => attributePhase(sw.attributes);

const lineFactor = (phase: Phase | undefined) =>
  phase === "three" ? Math.sqrt(3) : 1;

const conductorFactor = (phase: Phase | undefined) =>
  phase === "three" ? Math.sqrt(3) : 2;

const deviceDemand = ({ attributes }: Device, supply: Supply): Demand => {
  const ownVoltage = attributeNumber(attributes, "rated_voltage");
  const voltage = Number.isFinite(ownVoltage) ? ownVoltage : supply.voltage;
  const powerFactor = parsePowerFactor(
    attributeText(attributes, "power_factor"),
  );
  const ratedPower = attributeNumber(attributes, "rated_power");
  const ratedCurrent = attributeNumber(attributes, "rated_current");
  const factor = lineFactor(supply.phase);
  const power = Number.isFinite(ratedPower)
    ? ratedPower
    : factor * voltage * ratedCurrent * powerFactor;
  const apparent = power / powerFactor;
  return {
    power,
    apparent,
    current: Number.isFinite(ratedCurrent)
      ? ratedCurrent
      : apparent / (factor * voltage),
    voltage,
  };
};

const effectiveSupply = (sw: Switch, inherited: Supply) => {
  const ownVoltage = switchVoltage(sw);
  const isVoltageInherited = !Number.isFinite(ownVoltage);
  return {
    voltage: isVoltageInherited ? inherited.voltage : ownVoltage,
    phase: switchPhase(sw) ?? inherited.phase,
    isVoltageInherited,
  };
};

const summarize = (
  sw: Switch,
  supply: Supply & { isVoltageInherited: boolean },
  demands: Demand[],
): SwitchLoad => {
  let power = 0;
  let apparent = 0;
  let unknown = 0;
  for (const demand of demands) {
    if (Number.isFinite(demand.power)) {
      power += demand.power;
      apparent += demand.apparent;
    } else {
      unknown++;
    }
  }
  return {
    ...supply,
    power,
    apparent,
    current:
      supply.voltage > 0
        ? apparent / (lineFactor(supply.phase) * supply.voltage)
        : Number.NaN,
    limit: attributeNumber(sw.attributes, "max_current"),
    unknown,
  };
};

const linkVoltageDrop = (
  link: Attributes | null,
  current: number,
  phase: Phase | undefined,
) => {
  const length = attributeNumber(link, "wire_length");
  const size = attributeNumber(link, "wire_size");
  if (!(length >= 0 && size > 0)) {
    return Number.NaN;
  }
  return (
    (conductorFactor(phase) *
      length *
      current *
      COPPER_RESISTIVITY_OHM_MM2_PER_M) /
    size
  );
};

export const computeChainLoads = (
  roots: ChainTreeNode[],
  supply: Supply,
  isEnabled: (node: ChainNode) => boolean,
): ChainLoads => {
  const switches = new Map<SwitchId, SwitchLoad>();
  const links = new Map<string, LinkLoad>();
  const drops = new Map<string, NodeVoltageDrop>();

  const isAttached = ({ node, isFocus }: ChainTreeNode) =>
    isFocus || isEnabled(node);

  const collect = (
    item: ChainTreeNode,
    inherited: Supply,
  ): Map<DeviceId, Demand> => {
    const { node, children } = item;
    if (!isAttached(item)) {
      return new Map();
    }
    if (node.kind === "device") {
      return new Map([[node.device.id, deviceDemand(node.device, inherited)]]);
    }

    const own = effectiveSupply(node.sw, inherited);
    const demands = new Map<DeviceId, Demand>();
    for (const child of children) {
      for (const [id, demand] of collect(child, own)) {
        if (!demands.has(id)) {
          demands.set(id, demand);
        }
      }
    }

    if (!switches.has(node.sw.id)) {
      switches.set(node.sw.id, summarize(node.sw, own, [...demands.values()]));
    }
    return demands;
  };

  const walk = (
    item: ChainTreeNode,
    parent: { sw: Switch; load: SwitchLoad } | undefined,
    upstreamDrop: number,
    isUpstreamPartial: boolean,
  ) => {
    const { node, link, children } = item;
    if (!isAttached(item)) {
      return;
    }
    const ownLoad =
      node.kind === "switch" ? switches.get(node.sw.id) : undefined;
    const demand =
      node.kind === "device" && parent
        ? deviceDemand(node.device, parent.load)
        : undefined;
    const current = ownLoad?.current ?? demand?.current ?? Number.NaN;
    const voltage = ownLoad?.voltage ?? demand?.voltage ?? Number.NaN;

    let drop = upstreamDrop;
    let isPartial = isUpstreamPartial;
    if (parent) {
      const voltageDrop = linkVoltageDrop(link, current, parent.load.phase);
      links.set(chainLinkKey(parent.sw.id, node), {
        current,
        ampacity: attributeNumber(link, "rated_current"),
        voltageDrop,
      });
      if (Number.isFinite(voltageDrop)) {
        drop += voltageDrop;
      } else {
        isPartial = true;
      }
    }

    const key = chainNodeKey(node);
    const previous = drops.get(key);
    if (!previous || drop > previous.drop) {
      drops.set(key, { drop, voltage, isPartial });
    }

    if (node.kind === "switch" && ownLoad) {
      for (const child of children) {
        walk(child, { sw: node.sw, load: ownLoad }, drop, isPartial);
      }
    }
  };

  for (const root of roots) {
    collect(root, supply);
  }
  for (const root of roots) {
    walk(root, undefined, 0, false);
  }
  return { switches, links, drops };
};

export const isOverloaded = ({ current, limit }: SwitchLoad) =>
  Number.isFinite(current) && Number.isFinite(limit) && current > limit;

export const isLinkOverloaded = ({ current, ampacity }: LinkLoad) =>
  Number.isFinite(current) && Number.isFinite(ampacity) && current > ampacity;

export const voltageDropRatio = ({ drop, voltage }: NodeVoltageDrop) =>
  voltage > 0 ? drop / voltage : Number.NaN;

export const isVoltageDropExcessive = (drop: NodeVoltageDrop) =>
  voltageDropRatio(drop) > VOLTAGE_DROP_LIMIT;

const formatQuantity = (value: number) => String(Number(value.toFixed(1)));

const lowerBound = ({ unknown }: SwitchLoad) => (unknown > 0 ? "≥" : "");

export const formatSwitchPower = (load: SwitchLoad) =>
  load.power >= 1000
    ? `${lowerBound(load)}${formatQuantity(load.power / 1000)} kW`
    : `${lowerBound(load)}${formatQuantity(load.power)} W`;

export const formatSwitchLoad = (load: SwitchLoad) => {
  if (!Number.isFinite(load.current)) {
    return formatSwitchPower(load);
  }
  const current = `${lowerBound(load)}${formatQuantity(load.current)}`;
  return Number.isFinite(load.limit)
    ? `${current} / ${formatQuantity(load.limit)} A`
    : `${current} A`;
};

export const formatSwitchVoltage = ({ voltage }: SwitchLoad) =>
  Number.isFinite(voltage) ? `${formatQuantity(voltage)} V` : undefined;

export const formatLinkLoad = ({ current, ampacity }: LinkLoad) => {
  if (!Number.isFinite(current)) {
    return undefined;
  }
  return Number.isFinite(ampacity)
    ? `${formatQuantity(current)} / ${formatQuantity(ampacity)} A`
    : `${formatQuantity(current)} A`;
};

export const formatVoltageDrop = (drop: NodeVoltageDrop) => {
  const ratio = voltageDropRatio(drop);
  const prefix = drop.isPartial ? "≥" : "";
  const volts = `${prefix}${Number(drop.drop.toFixed(2))} V`;
  return Number.isFinite(ratio)
    ? `${volts} (${prefix}${formatQuantity(ratio * 100)}%)`
    : volts;
};
