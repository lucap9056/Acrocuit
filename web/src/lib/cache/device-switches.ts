import type { AuthSession, DeviceId, SwitchId } from "@/api";
import api from "@/api";
import { createListCache } from "@/lib/cache/list-cache";

const cache = createListCache(api.devices.getDeviceSwitches);

export const clearCache = cache.clear;

export const fetchDeviceSwitches = cache.fetch;

export const addDeviceSwitch = (
  s: AuthSession,
  device: DeviceId,
  sw: SwitchId,
) =>
  api.devices.addDeviceSwitch(s, device, sw).then(() => {
    cache.clear(device);
    return cache.fetch(s, device);
  });

export const removeDeviceSwitch = (
  s: AuthSession,
  device: DeviceId,
  sw: SwitchId,
) => cache.remove(s, device, sw, api.devices.deleteDeviceSwitch(s, device, sw));
