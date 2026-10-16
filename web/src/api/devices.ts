import type { AuthSession } from "@/api/auth/auth-session";
import type {
  Breaker,
  BreakerId,
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

export type DeviceUpdate = {
  name?: string;
  position?: Position;
};

const updateDevice = (
  auth: AuthSession,
  deviceId: DeviceId,
  update: DeviceUpdate,
) => {
  return auth
    .fetch(`${base}/${deviceId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(update),
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

const getDeviceBreakers = (auth: AuthSession, deviceId: DeviceId) => {
  return auth
    .fetch(`${base}/${deviceId}/breakers`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Breaker[]>);
};

const addDeviceBreaker = (
  auth: AuthSession,
  deviceId: DeviceId,
  breakerId: BreakerId,
) => {
  return auth
    .fetch(`${base}/${deviceId}/breakers`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ breaker_id: breakerId }),
    })
    .then(unwrapApiResponse<undefined>);
};

const deleteDeviceBreaker = (
  auth: AuthSession,
  deviceId: DeviceId,
  breakerId: BreakerId,
) => {
  return auth
    .fetch(`${base}/${deviceId}/breakers/${breakerId}`, {
      method: "DELETE",
    })
    .then(unwrapApiResponse<undefined>);
};

const getDeviceUpstream = (auth: AuthSession, deviceId: DeviceId) => {
  return auth
    .fetch(`${base}/${deviceId}/upstream`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Breaker[]>);
};

export default {
  getDevice,
  updateDevice,
  updateDevicePosition,
  deleteDevice,
  getDeviceBreakers,
  addDeviceBreaker,
  deleteDeviceBreaker,
  getDeviceUpstream,
};
