import { AuthSession, type RefreshCallback } from "@/api";

let pending: Promise<AuthSession> | undefined;

export const getSession = (callback: RefreshCallback) => {
  if (pending === undefined) {
    pending = AuthSession.New(callback);
  }
  return pending;
};

export const setSession = (s: AuthSession) => {
  pending = Promise.resolve(s);
};

export const clearSession = () => {
  pending = undefined;
};
