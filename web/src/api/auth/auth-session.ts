import type { AccessToken, TokenPair } from "@/api/auth";
import auth, { AuthError } from "@/api/auth";
import meta from "@/api/meta";

const refreshAheadMs = 60_000;

const decodeExpiresAt = (token: AccessToken) => {
  try {
    const payload = token.split(".")[1].replace(/-/g, "+").replace(/_/g, "/");
    const { exp } = JSON.parse(atob(payload)) as { exp?: number };
    return typeof exp === "number" ? exp * 1000 : 0;
  } catch {
    return 0;
  }
};

export interface RefreshCallback {
  onsuccess: () => void;
  onerror: (error: AuthError) => void;
}

export type AuthSession = GuestAuthSession | MemberAuthSession;

export const connectSession = async (
  refreshCallback: RefreshCallback,
): Promise<AuthSession> => {
  const { auth_required } = await meta.getMeta();
  if (!auth_required) {
    return new MemberAuthSession(refreshCallback, false);
  }

  try {
    await auth.refreshStatus();
  } catch (err) {
    if (err instanceof AuthError && err.status === 401) {
      return new GuestAuthSession();
    }
    throw err;
  }
  return new MemberAuthSession(refreshCallback, true);
};

export class GuestAuthSession {
  readonly kind = "guest";

  public fetch(input: RequestInfo | URL, init: RequestInit = {}) {
    return fetch(input, init);
  }
}

export class MemberAuthSession {
  readonly kind = "member";

  private accessToken?: AccessToken;
  private expiresAt = 0;
  private refreshing: Promise<AccessToken> | null = null;
  constructor(
    private refreshCallback: RefreshCallback,
    readonly authRequired: boolean,
    token?: TokenPair,
  ) {
    if (token) {
      this.accessToken = token.access_token;
      this.expiresAt = decodeExpiresAt(token.access_token);
    }
  }

  private async getAccessToken() {
    const remaining = this.expiresAt - Date.now();

    if (!this.accessToken || remaining <= 0) {
      return this.refresh();
    }

    if (remaining <= refreshAheadMs) {
      this.refresh().catch(() => {});
    }

    return this.accessToken;
  }

  public async fetch(input: RequestInfo | URL, init: RequestInit = {}) {
    if (!this.authRequired) {
      return fetch(input, init);
    }
    const accessToken = await this.getAccessToken();
    const headers = new Headers(init.headers);
    headers.set("Authorization", `Bearer ${accessToken}`);
    return fetch(input, { ...init, headers });
  }

  private refresh() {
    this.refreshing ??= auth
      .refreshAccess()
      .then((token) => {
        this.accessToken = token;
        this.expiresAt = decodeExpiresAt(token);
        this.refreshCallback.onsuccess();
        return token;
      })
      .catch((error: AuthError) => {
        this.refreshCallback.onerror(error);
        throw error;
      })
      .finally(() => {
        this.refreshing = null;
      });

    return this.refreshing;
  }
}
