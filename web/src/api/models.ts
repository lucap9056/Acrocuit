export type SpaceGroupId = Brand<number, "SpaceGroupId">;
export type SpaceId = Brand<number, "SpaceId">;
export type BreakerGroupId = Brand<number, "BreakerGroupId">;
export type BreakerId = Brand<number, "BreakerId">;
export type DeviceId = Brand<number, "DeviceId">;

export type Ref<Id> = {
  id: Id;
};

export type Position = {
  x: number;
  y: number;
  z: number;
};

export type SpaceGroup = {
  id: SpaceGroupId;
  name: string;
};

export type Space = {
  id: SpaceId;
  name: string;
  display_order: number;
  background_image_updated_at: number;
  space_group: Ref<SpaceGroupId> | null;
};

export type BreakerGroup = {
  id: BreakerGroupId;
  name: string;
  space: Ref<SpaceId> | null;
  position: Position | null;
};

export type Breaker = {
  id: BreakerId;
  name: string;
  display_order: number;
  breaker_group: Ref<BreakerGroupId> | null;
  upstream_breaker: Ref<BreakerId> | null;
};

export type Device = {
  id: DeviceId;
  name: string;
  space: Ref<SpaceId> | null;
  breaker: Ref<BreakerId> | null;
  position: Position | null;
};

export type BreakerDownstream = {
  breakers: Breaker[];
  devices: Device[];
};

export type Meta = {
  auth_required: boolean;
};
