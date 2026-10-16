import type { AuthSession } from "@/api/auth/auth-session";
import type { Space, SpaceGroup, SpaceGroupId } from "@/api/models";
import { unwrapApiResponse } from "@/api/response";

const base = "/api/space-groups";

const getSpaceGroups = (auth: AuthSession) => {
  return auth
    .fetch(base, {
      method: "GET",
    })
    .then(unwrapApiResponse<SpaceGroup[]>);
};

const createSpaceGroup = (auth: AuthSession, name: string) => {
  return auth
    .fetch(base, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name }),
    })
    .then(unwrapApiResponse<SpaceGroup>);
};

const getSpaceGroup = async (auth: AuthSession, spaceGroupId: SpaceGroupId) => {
  return auth
    .fetch(`${base}/${spaceGroupId}`, {
      method: "GET",
    })
    .then(unwrapApiResponse<SpaceGroup>);
};

const updateSpaceGroupName = async (
  auth: AuthSession,
  spaceGroupId: SpaceGroupId,
  name: string,
) => {
  return auth
    .fetch(`${base}/${spaceGroupId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name }),
    })
    .then(unwrapApiResponse<SpaceGroup>);
};

const deleteSpaceGroup = async (
  auth: AuthSession,
  spaceGroupId: SpaceGroupId,
) => {
  return auth
    .fetch(`${base}/${spaceGroupId}`, {
      method: "DELETE",
    })
    .then(unwrapApiResponse<undefined>);
};

const getSpaces = async (auth: AuthSession, spaceGroupId: SpaceGroupId) => {
  return auth
    .fetch(`${base}/${spaceGroupId}/spaces`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Space[]>);
};

const addSpace = async (
  auth: AuthSession,
  spaceGroupId: SpaceGroupId,
  name: string,
) => {
  return auth
    .fetch(`${base}/${spaceGroupId}/spaces`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name }),
    })
    .then(unwrapApiResponse<Space>);
};

export default {
  getSpaceGroups,
  createSpaceGroup,
  getSpaceGroup,
  updateSpaceGroupName,
  deleteSpaceGroup,
  getSpaces,
  addSpace,
};
