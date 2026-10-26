import type { AuthSession } from "@/api/auth/auth-session";
import type {
  Attributes,
  Device,
  DeviceId,
  DeviceLink,
  LinkedSwitch,
  Position,
  PowerChainGraph,
  SwitchId,
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

const updateDeviceAttributes = (
  auth: AuthSession,
  deviceId: DeviceId,
  attributes: Attributes,
) => {
  return auth
    .fetch(`${base}/${deviceId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ attributes }),
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
    .then(unwrapApiResponse<LinkedSwitch[]>);
};

const addDeviceSwitch = (
  auth: AuthSession,
  deviceId: DeviceId,
  switchId: SwitchId,
  attributes?: Attributes,
) => {
  return auth
    .fetch(`${base}/${deviceId}/switches`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ switch_id: switchId, attributes }),
    })
    .then(unwrapApiResponse<DeviceLink>);
};

const updateDeviceSwitch = (
  auth: AuthSession,
  deviceId: DeviceId,
  switchId: SwitchId,
  attributes: Attributes,
) => {
  return auth
    .fetch(`${base}/${deviceId}/switches/${switchId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ attributes }),
    })
    .then(unwrapApiResponse<DeviceLink>);
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
    .then(unwrapApiResponse<PowerChainGraph>);
};

export default {
  getDevice,
  updateDeviceName,
  updateDeviceAttributes,
  updateDevicePosition,
  deleteDevice,
  getDeviceSwitches,
  addDeviceSwitch,
  updateDeviceSwitch,
  deleteDeviceSwitch,
  getDeviceUpstream,
};
