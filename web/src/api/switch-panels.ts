import type { AuthSession } from "@/api/auth/auth-session";
import type {
  Attributes,
  Position,
  Switch,
  SwitchPanel,
  SwitchPanelId,
} from "@/api/models";
import { unwrapApiResponse } from "@/api/response";

const base = "/api/switch-panels";

const getSwitchPanel = (auth: AuthSession, switchPanelId: SwitchPanelId) => {
  return auth
    .fetch(`${base}/${switchPanelId}`, {
      method: "GET",
    })
    .then(unwrapApiResponse<SwitchPanel>);
};

const updateSwitchPanelName = (
  auth: AuthSession,
  switchPanelId: SwitchPanelId,
  name: string,
) => {
  return auth
    .fetch(`${base}/${switchPanelId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name }),
    })
    .then(unwrapApiResponse<SwitchPanel>);
};

const updateSwitchPanelAttributes = (
  auth: AuthSession,
  switchPanelId: SwitchPanelId,
  attributes: Attributes,
) => {
  return auth
    .fetch(`${base}/${switchPanelId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ attributes }),
    })
    .then(unwrapApiResponse<SwitchPanel>);
};

const updateSwitchPanelPosition = (
  auth: AuthSession,
  switchPanelId: SwitchPanelId,
  position: Position,
) => {
  return auth
    .fetch(`${base}/${switchPanelId}/position`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ position }),
    })
    .then(unwrapApiResponse<SwitchPanel>);
};

const deleteSwitchPanel = (auth: AuthSession, switchPanelId: SwitchPanelId) => {
  return auth
    .fetch(`${base}/${switchPanelId}`, {
      method: "DELETE",
    })
    .then(unwrapApiResponse<undefined>);
};

const getSwitches = (auth: AuthSession, switchPanelId: SwitchPanelId) => {
  return auth
    .fetch(`${base}/${switchPanelId}/switches`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Switch[]>);
};

const addSwitch = (
  auth: AuthSession,
  switchPanelId: SwitchPanelId,
  name: string,
  attributes?: Attributes,
) => {
  return auth
    .fetch(`${base}/${switchPanelId}/switches`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name, attributes }),
    })
    .then(unwrapApiResponse<Switch>);
};

export default {
  getSwitchPanel,
  updateSwitchPanelName,
  updateSwitchPanelAttributes,
  updateSwitchPanelPosition,
  deleteSwitchPanel,
  getSwitches,
  addSwitch,
};
