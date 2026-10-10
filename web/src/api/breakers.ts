import type { AuthSession } from "./auth/auth-session";
import type { Breaker, BreakerDownstream, BreakerId } from "./models";
import { unwrapApiResponse } from "./response";

const base = "/api/breakers";

const getBreaker = (auth: AuthSession, breakerId: BreakerId) => {
  return auth
    .fetch(`${base}/${breakerId}`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Breaker>);
};

export type BreakerUpdate = {
  name?: string;
  display_order?: number;
};

const updateBreaker = (
  auth: AuthSession,
  breakerId: BreakerId,
  update: BreakerUpdate,
) => {
  return auth
    .fetch(`${base}/${breakerId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(update),
    })
    .then(unwrapApiResponse<Breaker>);
};

const deleteBreaker = (auth: AuthSession, breakerId: BreakerId) => {
  return auth
    .fetch(`${base}/${breakerId}`, {
      method: "DELETE",
    })
    .then(unwrapApiResponse<undefined>);
};

const updateBreakerOrder = (
  auth: AuthSession,
  breakerId: BreakerId,
  displayOrder: number,
) => {
  return auth
    .fetch(`${base}/${breakerId}/order`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ display_order: displayOrder }),
    })
    .then(unwrapApiResponse<Breaker>);
};

const updateBreakerUpstream = (
  auth: AuthSession,
  breakerId: BreakerId,
  upstreamBreakerId: BreakerId | null,
) => {
  return auth
    .fetch(`${base}/${breakerId}/upstream`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ upstream_breaker_id: upstreamBreakerId }),
    })
    .then(unwrapApiResponse<Breaker>);
};

const getBreakerDownstream = (auth: AuthSession, breakerId: BreakerId) => {
  return auth
    .fetch(`${base}/${breakerId}/downstream`, {
      method: "GET",
    })
    .then(unwrapApiResponse<BreakerDownstream>);
};

export default {
  getBreaker,
  updateBreaker,
  deleteBreaker,
  updateBreakerOrder,
  updateBreakerUpstream,
  getBreakerDownstream,
};
