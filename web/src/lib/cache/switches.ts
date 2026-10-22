import type { AuthSession, Switch, SwitchPanelId, SwitchId, Ref } from "@/api";
import api from "@/api";
import { clearCache as clearDeviceSwitchesCache } from "@/lib/cache/device-switches";
import { clearCache as clearDevicesCache } from "@/lib/cache/devices";
import { createListCache } from "@/lib/cache/list-cache";

type CompleteSwitch = Switch & { switch_panel: Ref<SwitchPanelId> };

const cache = createListCache<SwitchPanelId, Switch>(
  api.switchPanels.getSwitches,
);

export const clearCache = cache.clear;

export const fetchSwitches = cache.fetch;

export const addSwitch = (
  s: AuthSession,
  panel: SwitchPanelId,
  name: string,
  upstream?: SwitchId,
) => cache.add(s, panel, api.switchPanels.addSwitch(s, panel, name, upstream));

export const renameSwitch = (
  s: AuthSession,
  sw: CompleteSwitch,
  name: string,
) =>
  cache.replace(
    s,
    sw.switch_panel.id,
    api.switches.updateSwitchName(s, sw.id, name),
  );

export const setSwitchUpstream = (
  s: AuthSession,
  sw: CompleteSwitch,
  upstream: SwitchId | null,
) =>
  cache.replace(
    s,
    sw.switch_panel.id,
    api.switches.updateSwitchUpstream(s, sw.id, upstream),
  );

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
    return fetchSwitches(s, sw.switch_panel.id);
  });
