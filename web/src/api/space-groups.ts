import type { AuthSession } from "@/api/auth/auth-session";
import type {
  Space,
  SpaceGroup,
  SpaceGroupId,
  SpaceGroupMemberRole,
  SpaceGroupMembers,
} from "@/api/models";
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

const memberUrl = (spaceGroupId: SpaceGroupId, email: string) =>
  `${base}/${spaceGroupId}/members/${encodeURIComponent(email)}`;

const getMembers = async (auth: AuthSession, spaceGroupId: SpaceGroupId) => {
  return auth
    .fetch(`${base}/${spaceGroupId}/members`, {
      method: "GET",
    })
    .then(unwrapApiResponse<SpaceGroupMembers>);
};

const addMember = async (
  auth: AuthSession,
  spaceGroupId: SpaceGroupId,
  email: string,
  role: SpaceGroupMemberRole,
) => {
  return auth
    .fetch(`${base}/${spaceGroupId}/members`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email, role }),
    })
    .then(unwrapApiResponse<undefined>);
};

const updateMemberRole = async (
  auth: AuthSession,
  spaceGroupId: SpaceGroupId,
  email: string,
  role: SpaceGroupMemberRole,
) => {
  return auth
    .fetch(memberUrl(spaceGroupId, email), {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ role }),
    })
    .then(unwrapApiResponse<undefined>);
};

const deleteMember = async (
  auth: AuthSession,
  spaceGroupId: SpaceGroupId,
  email: string,
) => {
  return auth
    .fetch(memberUrl(spaceGroupId, email), {
      method: "DELETE",
    })
    .then(unwrapApiResponse<undefined>);
};

const leaveSpaceGroup = async (
  auth: AuthSession,
  spaceGroupId: SpaceGroupId,
) => {
  return deleteMember(auth, spaceGroupId, "me");
};

const transferOwner = async (
  auth: AuthSession,
  spaceGroupId: SpaceGroupId,
  email: string,
) => {
  return auth
    .fetch(`${base}/${spaceGroupId}/owner`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email }),
    })
    .then(unwrapApiResponse<undefined>);
};

const setPublic = async (
  auth: AuthSession,
  spaceGroupId: SpaceGroupId,
  isPublic: boolean,
) => {
  return auth
    .fetch(`${base}/${spaceGroupId}/public`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ public: isPublic }),
    })
    .then(unwrapApiResponse<undefined>);
};

export default {
  getSpaceGroups,
  createSpaceGroup,
  getSpaceGroup,
  updateSpaceGroupName,
  deleteSpaceGroup,
  getSpaces,
  addSpace,
  getMembers,
  addMember,
  updateMemberRole,
  deleteMember,
  leaveSpaceGroup,
  transferOwner,
  setPublic,
};
