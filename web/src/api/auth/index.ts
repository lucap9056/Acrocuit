type Body<T> =
  | { success: true; message: T }
  | { success: false; message: string };

export type AuthErrorReason = "device_not_found" | "invalid_signature";

export class AuthError extends Error {
  readonly status: number;
  readonly reason: AuthErrorReason | null;

  constructor(message: string, status: number, reason: AuthErrorReason | null) {
    super(message);
    this.name = "AuthError";
    this.status = status;
    this.reason = reason;
  }
}

const unwrapResponse = async <T>(r: Response): Promise<T> => {
  const body = await r.json().then(
    (b) => b as Body<T>,
    () => null,
  );

  if (r.ok && body?.success) {
    return body.message;
  }

  const message =
    body && !body.success && typeof body.message === "string"
      ? body.message
      : r.statusText || `HTTP ${r.status}`;
  const reason = r.headers.get("X-Auth-Error") as AuthErrorReason | null;

  throw new AuthError(message, r.status, reason);
};

//
//
//

export type Verifier = Brand<string, "Verifier">;

export type LoginChallenge = {
  url: string;
  verifier?: Verifier;
};

const login = () => fetch(`/auth/login`).then(unwrapResponse<LoginChallenge>);

export type RefreshToken = Brand<string, "RefreshToken">;
export type AccessToken = Brand<string, "AccessToken">;

export type TokenPair = {
  refresh_token: RefreshToken;
  access_token: AccessToken;
};

export type CallbackOptions = {
  deviceName?: string;
  verifier?: Verifier;
};

const loginCallback = (
  search: string | URLSearchParams,
  { deviceName, verifier }: CallbackOptions = {},
) => {
  const headers = new Headers();
  if (deviceName) {
    headers.set("X-Device-Name", deviceName);
  }
  if (verifier) {
    headers.set("X-PKCE-Verifier", verifier);
  }
  return fetch(`/auth/callback?${new URLSearchParams(search)}`, {
    headers,
  }).then(unwrapResponse<TokenPair>);
};

const refresh = () =>
  fetch(`/auth/refresh`, { method: "POST" }).then(unwrapResponse<TokenPair>);

const refreshAccess = () =>
  fetch(`/auth/refresh-access`, { method: "POST" }).then(
    unwrapResponse<AccessToken>,
  );

export type SessionStatus = {
  device_id: string;
  issued_at: number;
  expired_at: number;
};

const refreshStatus = () =>
  fetch(`/auth/refresh-status`, { method: "POST" }).then(
    unwrapResponse<SessionStatus>,
  );

const logout = () =>
  fetch(`/auth/logout`, { method: "POST" }).then(unwrapResponse<string>);

const deleteMe = () =>
  fetch(`/auth/users/me`, { method: "DELETE" }).then(unwrapResponse<string>);

export default {
  login,
  loginCallback,
  refresh,
  refreshAccess,
  refreshStatus,
  logout,
  deleteMe,
};
