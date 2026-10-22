import type { AuthSession } from "@/api/auth/auth-session";
import type {
  SwitchPanel,
  Device,
  Position,
  Space,
  SpaceId,
} from "@/api/models";
import { unwrapApiResponse } from "@/api/response";

const base = "/api/spaces";

const getSpace = (auth: AuthSession, spaceId: SpaceId) => {
  return auth
    .fetch(`${base}/${spaceId}`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Space>);
};

const updateSpaceName = (auth: AuthSession, spaceId: SpaceId, name: string) => {
  return auth
    .fetch(`${base}/${spaceId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name }),
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
    .then(unwrapApiResponse<Space[]>);
};

const getSpaceBackgroundImage = async (
  auth: AuthSession,
  spaceId: SpaceId,
  version?: number,
) => {
  const query = version === undefined ? "" : `?v=${version}`;
  const r = await auth.fetch(`${base}/${spaceId}/background-image${query}`, {
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

const getSwitchPanels = (auth: AuthSession, spaceId: SpaceId) => {
  return auth
    .fetch(`${base}/${spaceId}/switch-panels`, {
      method: "GET",
    })
    .then(unwrapApiResponse<SwitchPanel[]>);
};

const addSwitchPanel = (
  auth: AuthSession,
  spaceId: SpaceId,
  name: string,
  position?: Position,
) => {
  return auth
    .fetch(`${base}/${spaceId}/switch-panels`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name, position }),
    })
    .then(unwrapApiResponse<SwitchPanel>);
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
  updateSpaceName,
  deleteSpace,
  updateSpaceOrder,
  getSpaceBackgroundImage,
  setSpaceBackgroundImage,
  deleteSpaceBackgroundImage,
  getSwitchPanels,
  addSwitchPanel,
  getDevices,
  addDevice,
};
