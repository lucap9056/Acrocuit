import type { AuthSession } from "@/api/auth/auth-session";
import type {
  Attributes,
  LinkedSwitch,
  PowerChainGraph,
  Switch,
  SwitchId,
  SwitchLink,
} from "@/api/models";
import { unwrapApiResponse } from "@/api/response";

const base = "/api/switches";

const getSwitch = (auth: AuthSession, switchId: SwitchId) => {
  return auth
    .fetch(`${base}/${switchId}`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Switch>);
};

const updateSwitchName = (
  auth: AuthSession,
  switchId: SwitchId,
  name: string,
) => {
  return auth
    .fetch(`${base}/${switchId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name }),
    })
    .then(unwrapApiResponse<Switch>);
};

const updateSwitchAttributes = (
  auth: AuthSession,
  switchId: SwitchId,
  attributes: Attributes,
) => {
  return auth
    .fetch(`${base}/${switchId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ attributes }),
    })
    .then(unwrapApiResponse<Switch>);
};

const deleteSwitch = (auth: AuthSession, switchId: SwitchId) => {
  return auth
    .fetch(`${base}/${switchId}`, {
      method: "DELETE",
    })
    .then(unwrapApiResponse<undefined>);
};

const updateSwitchOrder = (
  auth: AuthSession,
  switchId: SwitchId,
  displayOrder: number,
) => {
  return auth
    .fetch(`${base}/${switchId}/order`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ display_order: displayOrder }),
    })
    .then(unwrapApiResponse<Switch[]>);
};

const getSwitchUpstreams = (auth: AuthSession, switchId: SwitchId) => {
  return auth
    .fetch(`${base}/${switchId}/upstreams`, {
      method: "GET",
    })
    .then(unwrapApiResponse<LinkedSwitch[]>);
};

const addSwitchUpstream = (
  auth: AuthSession,
  switchId: SwitchId,
  upstreamSwitchId: SwitchId,
  attributes?: Attributes,
) => {
  return auth
    .fetch(`${base}/${switchId}/upstreams`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        upstream_switch_id: upstreamSwitchId,
        attributes,
      }),
    })
    .then(unwrapApiResponse<SwitchLink>);
};

const updateSwitchUpstream = (
  auth: AuthSession,
  switchId: SwitchId,
  upstreamSwitchId: SwitchId,
  attributes: Attributes,
) => {
  return auth
    .fetch(`${base}/${switchId}/upstreams/${upstreamSwitchId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ attributes }),
    })
    .then(unwrapApiResponse<SwitchLink>);
};

const deleteSwitchUpstream = (
  auth: AuthSession,
  switchId: SwitchId,
  upstreamSwitchId: SwitchId,
) => {
  return auth
    .fetch(`${base}/${switchId}/upstreams/${upstreamSwitchId}`, {
      method: "DELETE",
    })
    .then(unwrapApiResponse<undefined>);
};

const getSwitchDownstream = (auth: AuthSession, switchId: SwitchId) => {
  return auth
    .fetch(`${base}/${switchId}/downstream`, {
      method: "GET",
    })
    .then(unwrapApiResponse<PowerChainGraph>);
};

export default {
  getSwitch,
  updateSwitchName,
  updateSwitchAttributes,
  deleteSwitch,
  updateSwitchOrder,
  getSwitchUpstreams,
  addSwitchUpstream,
  updateSwitchUpstream,
  deleteSwitchUpstream,
  getSwitchDownstream,
};
