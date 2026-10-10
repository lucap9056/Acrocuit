import type {
  AuthSession,
  Breaker,
  BreakerGroup,
  BreakerGroupId,
  Device,
  Space,
  SpaceGroup,
  SpaceGroupId,
  SpaceId,
} from "@/api";

export interface Cache {
  auth: Promise<AuthSession> | undefined;
  spaceGroups: Promise<SpaceGroup[]> | undefined;
  spaces: Map<SpaceGroupId, Promise<Space[]>>;
  breakerGroups: Map<SpaceId, Promise<BreakerGroup[]>>;
  breakers: Map<BreakerGroupId, Promise<Breaker[]>>;
  devices: Map<SpaceId, Promise<Device[]>>;
}

const cache: Cache = {
  auth: undefined,
  spaceGroups: undefined,
  spaces: new Map(),
  breakerGroups: new Map(),
  breakers: new Map(),
  devices: new Map(),
};

export default cache;
