export type * from "./auth";
export { AuthError } from "./auth";
export {
  type AuthSession,
  connectSession,
  GuestAuthSession,
  MemberAuthSession,
  type RefreshCallback,
} from "./auth/auth-session";

export type * from "./models";
export { ApiError } from "./response";

import auth from "./auth";
import devices from "./devices";
import spaceGroups from "./space-groups";
import spaces from "./spaces";
import switchPanels from "./switch-panels";
import switches from "./switches";

export default {
  auth,
  spaceGroups,
  spaces,
  switchPanels,
  switches,
  devices,
};
