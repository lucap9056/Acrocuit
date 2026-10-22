import type { AuthSession } from "@/api/auth/auth-session";
import type { Switch, SwitchDownstream, SwitchId } from "@/api/models";
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

const updateSwitchUpstream = (
  auth: AuthSession,
  switchId: SwitchId,
  upstreamSwitchId: SwitchId | null,
) => {
  return auth
    .fetch(`${base}/${switchId}/upstream`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ upstream_switch_id: upstreamSwitchId }),
    })
    .then(unwrapApiResponse<Switch>);
};

const getSwitchDownstream = (auth: AuthSession, switchId: SwitchId) => {
  return auth
    .fetch(`${base}/${switchId}/downstream`, {
      method: "GET",
    })
    .then(unwrapApiResponse<SwitchDownstream>);
};

export default {
  getSwitch,
  updateSwitchName,
  deleteSwitch,
  updateSwitchOrder,
  updateSwitchUpstream,
  getSwitchDownstream,
};
