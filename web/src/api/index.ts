export type * from "./auth";
export { AuthError } from "./auth";
export { AuthSession, type RefreshCallback } from "./auth/auth-session";

export type { BreakerUpdate } from "./breakers";
export type { DeviceUpdate } from "./devices";
export type * from "./models";
export { ApiError } from "./response";
export type { SpaceUpdate } from "./spaces";

import auth from "./auth";
import breakerGroups from "./breaker-groups";
import breakers from "./breakers";
import devices from "./devices";
import spaceGroups from "./space-groups";
import spaces from "./spaces";

export default {
  auth,
  spaceGroups,
  spaces,
  breakerGroups,
  breakers,
  devices,
};
