import type {
  Attributes,
  AuthSession,
  Ref,
  Switch,
  SwitchPanelId,
} from "@/api";
import api from "@/api";
import { clearCache as clearDeviceSwitchesCache } from "@/lib/cache/device-switches";
import { clearCache as clearDevicesCache } from "@/lib/cache/devices";
import { createListCache } from "@/lib/cache/list-cache";
import { clearCache as clearSwitchUpstreamsCache } from "@/lib/cache/switch-upstreams";

type CompleteSwitch = Switch & { switch_panel: Ref<SwitchPanelId> };

const cache = createListCache<SwitchPanelId, Switch>(
  api.switchPanels.getSwitches,
);

export const clearCache = cache.clear;

export const fetchSwitches = cache.fetch;

export const watchSwitches = cache.watch;

export const addSwitch = (s: AuthSession, panel: SwitchPanelId, name: string) =>
  cache.add(s, panel, api.switchPanels.addSwitch(s, panel, name));

export const renameSwitch = (
  s: AuthSession,
  sw: CompleteSwitch,
  name: string,
) =>
  cache
    .replace(
      s,
      sw.switch_panel.id,
      api.switches.updateSwitchName(s, sw.id, name),
    )
    .then((switches) => {
      clearSwitchUpstreamsCache();
      return switches;
    });

export const setSwitchAttributes = (
  s: AuthSession,
  sw: CompleteSwitch,
  attributes: Attributes,
) =>
  cache
    .replace(
      s,
      sw.switch_panel.id,
      api.switches.updateSwitchAttributes(s, sw.id, attributes),
    )
    .then((switches) => {
      clearSwitchUpstreamsCache();
      return switches;
    });

export const moveSwitch = (s: AuthSession, sw: CompleteSwitch) => {
  const panel = sw.switch_panel.id;

  return api.switches
    .updateSwitchOrder(s, sw.id, sw.display_order)
    .then((switches) => cache.set(panel, switches));
};

export const removeSwitch = (s: AuthSession, sw: CompleteSwitch) =>
  api.switches.deleteSwitch(s, sw.id).then(() => {
    clearCache();
    clearDevicesCache();
    clearDeviceSwitchesCache();
    clearSwitchUpstreamsCache();
    return fetchSwitches(s, sw.switch_panel.id);
  });
