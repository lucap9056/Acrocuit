import type { AuthSession } from "./auth/auth-session";
import type { BreakerGroup, Device, Position, Space, SpaceId } from "./models";
import { unwrapApiResponse } from "./response";

const base = "/api/spaces";

const getSpace = (auth: AuthSession, spaceId: SpaceId) => {
  return auth
    .fetch(`${base}/${spaceId}`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Space>);
};

export type SpaceUpdate = {
  name?: string;
  display_order?: number;
};

const updateSpace = (
  auth: AuthSession,
  spaceId: SpaceId,
  update: SpaceUpdate,
) => {
  return auth
    .fetch(`${base}/${spaceId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(update),
    })
    .then(unwrapApiResponse<Space>);
};

const deleteSpace = (auth: AuthSession, spaceId: SpaceId) => {
  return auth
    .fetch(`${base}/${spaceId}`, {
      method: "DELETE",
    })
    .then(unwrapApiResponse<undefined>);
};

const updateSpaceOrder = (
  auth: AuthSession,
  spaceId: SpaceId,
  displayOrder: number,
) => {
  return auth
    .fetch(`${base}/${spaceId}/order`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ display_order: displayOrder }),
    })
    .then(unwrapApiResponse<Space>);
};

const getSpaceBackgroundImage = async (auth: AuthSession, spaceId: SpaceId) => {
  const r = await auth.fetch(`${base}/${spaceId}/background-image`, {
    method: "GET",
  });

  if (!r.ok) {
    return unwrapApiResponse<never>(r);
  }

  return r.blob();
};

const setSpaceBackgroundImage = (
  auth: AuthSession,
  spaceId: SpaceId,
  image: Blob,
) => {
  const form = new FormData();
  form.append("image", image);

  return auth
    .fetch(`${base}/${spaceId}/background-image`, {
      method: "PUT",
      body: form,
    })
    .then(unwrapApiResponse<Space>);
};

const deleteSpaceBackgroundImage = (auth: AuthSession, spaceId: SpaceId) => {
  return auth
    .fetch(`${base}/${spaceId}/background-image`, {
      method: "DELETE",
    })
    .then(unwrapApiResponse<Space>);
};

const getBreakerGroups = (auth: AuthSession, spaceId: SpaceId) => {
  return auth
    .fetch(`${base}/${spaceId}/breaker-groups`, {
      method: "GET",
    })
    .then(unwrapApiResponse<BreakerGroup[]>);
};

const addBreakerGroup = (
  auth: AuthSession,
  spaceId: SpaceId,
  name: string,
  position?: Position,
) => {
  return auth
    .fetch(`${base}/${spaceId}/breaker-groups`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name, position }),
    })
    .then(unwrapApiResponse<BreakerGroup>);
};

const getDevices = (auth: AuthSession, spaceId: SpaceId) => {
  return auth
    .fetch(`${base}/${spaceId}/devices`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Device[]>);
};

const addDevice = (
  auth: AuthSession,
  spaceId: SpaceId,
  name: string,
  position?: Position,
) => {
  return auth
    .fetch(`${base}/${spaceId}/devices`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name, position }),
    })
    .then(unwrapApiResponse<Device>);
};

export default {
  getSpace,
  updateSpace,
  deleteSpace,
  updateSpaceOrder,
  getSpaceBackgroundImage,
  setSpaceBackgroundImage,
  deleteSpaceBackgroundImage,
  getBreakerGroups,
  addBreakerGroup,
  getDevices,
  addDevice,
};
