import type { AuthSession, Device, Position, Ref, SpaceId } from "@/api";
import api from "@/api";
import { createListCache } from "@/lib/cache/list-cache";

type CompleteDevice = Device & { space: Ref<SpaceId> };

const cache = createListCache<SpaceId, Device>(api.spaces.getDevices);

export const clearCache = cache.clear;

export const fetchDevices = cache.fetch;

export const addDevice = (
  s: AuthSession,
  space: SpaceId,
  name: string,
  position?: Position,
) => cache.add(s, space, api.spaces.addDevice(s, space, name, position));

export const renameDevice = (
  s: AuthSession,
  device: CompleteDevice,
  name: string,
) =>
  cache.replace(
    s,
    device.space.id,
    api.devices.updateDevice(s, device.id, { name }),
  );

export const setDevicePosition = (
  s: AuthSession,
  device: CompleteDevice,
  position: Position,
) =>
  cache.replace(
    s,
    device.space.id,
    api.devices.updateDevicePosition(s, device.id, position),
  );

export const removeDevice = (s: AuthSession, device: CompleteDevice) =>
  cache.remove(
    s,
    device.space.id,
    device.id,
    api.devices.deleteDevice(s, device.id),
  );
