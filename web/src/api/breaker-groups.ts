import type { AuthSession } from "@/api/auth/auth-session";
import type {
  Breaker,
  BreakerGroup,
  BreakerGroupId,
  BreakerId,
  Position,
} from "@/api/models";
import { unwrapApiResponse } from "@/api/response";

const base = "/api/breaker-groups";

const getBreakerGroup = (auth: AuthSession, breakerGroupId: BreakerGroupId) => {
  return auth
    .fetch(`${base}/${breakerGroupId}`, {
      method: "GET",
    })
    .then(unwrapApiResponse<BreakerGroup>);
};

const updateBreakerGroupName = (
  auth: AuthSession,
  breakerGroupId: BreakerGroupId,
  name: string,
) => {
  return auth
    .fetch(`${base}/${breakerGroupId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name }),
    })
    .then(unwrapApiResponse<BreakerGroup>);
};

const updateBreakerGroupPosition = (
  auth: AuthSession,
  breakerGroupId: BreakerGroupId,
  position: Position,
) => {
  return auth
    .fetch(`${base}/${breakerGroupId}/position`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ position }),
    })
    .then(unwrapApiResponse<BreakerGroup>);
};

const deleteBreakerGroup = (
  auth: AuthSession,
  breakerGroupId: BreakerGroupId,
) => {
  return auth
    .fetch(`${base}/${breakerGroupId}`, {
      method: "DELETE",
    })
    .then(unwrapApiResponse<undefined>);
};

const getBreakers = (auth: AuthSession, breakerGroupId: BreakerGroupId) => {
  return auth
    .fetch(`${base}/${breakerGroupId}/breakers`, {
      method: "GET",
    })
    .then(unwrapApiResponse<Breaker[]>);
};

const addBreaker = (
  auth: AuthSession,
  breakerGroupId: BreakerGroupId,
  name: string,
  upstreamBreakerId?: BreakerId,
) => {
  return auth
    .fetch(`${base}/${breakerGroupId}/breakers`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name, upstream_breaker_id: upstreamBreakerId }),
    })
    .then(unwrapApiResponse<Breaker>);
};

export default {
  getBreakerGroup,
  updateBreakerGroupName,
  updateBreakerGroupPosition,
  deleteBreakerGroup,
  getBreakers,
  addBreaker,
};
