import type { AuthSession } from "@/api/auth/auth-session";
import type {
  ChainSwitch,
  Switch,
  SwitchId,
  Device,
  DeviceId,
  Position,
} from "@/api/models";
import { unwrapApiResponse } from "@/api/response";

const base = "/api/devices";

const getDevice = (auth: AuthSession, deviceId: DeviceId) => {
  return auth
    .fetch(`${base}/${deviceId}`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Device>);
};

const updateDeviceName = (
  auth: AuthSession,
  deviceId: DeviceId,
  name: string,
) => {
  return auth
    .fetch(`${base}/${deviceId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name }),
    })
    .then(unwrapApiResponse<Device>);
};

const updateDevicePosition = (
  auth: AuthSession,
  deviceId: DeviceId,
  position: Position,
) => {
  return auth
    .fetch(`${base}/${deviceId}/position`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ position }),
    })
    .then(unwrapApiResponse<Device>);
};

const deleteDevice = (auth: AuthSession, deviceId: DeviceId) => {
  return auth
    .fetch(`${base}/${deviceId}`, {
      method: "DELETE",
    })
    .then(unwrapApiResponse<undefined>);
};

const getDeviceSwitches = (auth: AuthSession, deviceId: DeviceId) => {
  return auth
    .fetch(`${base}/${deviceId}/switches`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Switch[]>);
};

const addDeviceSwitch = (
  auth: AuthSession,
  deviceId: DeviceId,
  switchId: SwitchId,
) => {
  return auth
    .fetch(`${base}/${deviceId}/switches`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ switch_id: switchId }),
    })
    .then(unwrapApiResponse<undefined>);
};

const deleteDeviceSwitch = (
  auth: AuthSession,
  deviceId: DeviceId,
  switchId: SwitchId,
) => {
  return auth
    .fetch(`${base}/${deviceId}/switches/${switchId}`, {
      method: "DELETE",
    })
    .then(unwrapApiResponse<undefined>);
};

const getDeviceUpstream = (auth: AuthSession, deviceId: DeviceId) => {
  return auth
    .fetch(`${base}/${deviceId}/upstream`, {
      method: "GET",
    })
    .then(unwrapApiResponse<ChainSwitch[]>);
};

export default {
  getDevice,
  updateDeviceName,
  updateDevicePosition,
  deleteDevice,
  getDeviceSwitches,
  addDeviceSwitch,
  deleteDeviceSwitch,
  getDeviceUpstream,
};
