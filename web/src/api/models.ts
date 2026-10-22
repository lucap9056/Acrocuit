export type SpaceGroupId = Brand<number, "SpaceGroupId">;
export type SpaceId = Brand<number, "SpaceId">;
export type SwitchPanelId = Brand<number, "SwitchPanelId">;
export type SwitchId = Brand<number, "SwitchId">;
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

export type SwitchPanel = {
  id: SwitchPanelId;
  name: string;
  space: Ref<SpaceId> | null;
  position: Position | null;
};

export type Switch = {
  id: SwitchId;
  name: string;
  display_order: number;
  switch_panel: Ref<SwitchPanelId> | null;
  upstream_switch: Ref<SwitchId> | null;
};

export type Device = {
  id: DeviceId;
  name: string;
  space: Ref<SpaceId> | null;
  switch: Ref<SwitchId> | null;
  position: Position | null;
};

export type SwitchDownstream = {
  switches: Switch[];
  devices: Device[];
};

export type Meta = {
  auth_required: boolean;
};
