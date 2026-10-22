import type { AuthSession, SwitchPanel, Position, Ref, SpaceId } from "@/api";
import api from "@/api";
import { clearCache as clearSwitchesCache } from "@/lib/cache/switches";
import { clearCache as clearDeviceSwitchesCache } from "@/lib/cache/device-switches";
import { clearCache as clearDevicesCache } from "@/lib/cache/devices";
import { createListCache } from "@/lib/cache/list-cache";

type CompleteSwitchPanel = SwitchPanel & { space: Ref<SpaceId> };

const cache = createListCache<SpaceId, SwitchPanel>(api.spaces.getSwitchPanels);

export const clearCache = cache.clear;

export const fetchSwitchPanels = cache.fetch;

export const addSwitchPanel = (
  s: AuthSession,
  space: SpaceId,
  name: string,
  position?: Position,
) => cache.add(s, space, api.spaces.addSwitchPanel(s, space, name, position));

export const renameSwitchPanel = (
  s: AuthSession,
  panel: CompleteSwitchPanel,
  name: string,
) =>
  cache.replace(
    s,
    panel.space.id,
    api.switchPanels.updateSwitchPanelName(s, panel.id, name),
  );

export const setSwitchPanelPosition = (
  s: AuthSession,
  panel: CompleteSwitchPanel,
  position: Position,
) =>
  cache.replace(
    s,
    panel.space.id,
    api.switchPanels.updateSwitchPanelPosition(s, panel.id, position),
  );

export const removeSwitchPanel = (s: AuthSession, panel: CompleteSwitchPanel) =>
  cache
    .remove(
      s,
      panel.space.id,
      panel.id,
      api.switchPanels.deleteSwitchPanel(s, panel.id),
    )
    .then((switchPanels) => {
      clearSwitchesCache();
      clearDevicesCache();
      clearDeviceSwitchesCache();
      return switchPanels;
    });
