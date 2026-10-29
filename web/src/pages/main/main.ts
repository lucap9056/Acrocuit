import {
  type AuthSession,
  connectSession,
  LocalAuthSession,
  type RefreshCallback,
} from "@/api";
import { readLocalMode } from "@/lib/local-backend/mode";

let pending: Promise<AuthSession> | undefined;

export const getSession = (callback: RefreshCallback) => {
  if (pending === undefined) {
    pending = readLocalMode()
      ? Promise.resolve(new LocalAuthSession())
      : connectSession(callback);
  }
  return pending;
};

export const setSession = (s: AuthSession) => {
  pending = Promise.resolve(s);
};

export const clearSession = () => {
  pending = undefined;
};
