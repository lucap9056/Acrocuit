export type SpaceGroupId = Brand<string, "SpaceGroupId">;
export type SpaceId = Brand<string, "SpaceId">;
export type SwitchPanelId = Brand<string, "SwitchPanelId">;
export type SwitchId = Brand<string, "SwitchId">;
export type DeviceId = Brand<string, "DeviceId">;

export type Ref<Id> = {
  id: Id;
};

export type Position = {
  x: number;
  y: number;
  z: number;
};

export type Attributes = Record<string, unknown>;

export type SpaceGroupRole = "owner" | "editor" | "viewer";

export type SpaceGroupMemberRole = Exclude<SpaceGroupRole, "owner">;

export type SpaceGroup = {
  id: SpaceGroupId;
  name: string;
  role: SpaceGroupRole;
};

export type SpaceGroupMember = {
  email: string;
  role: SpaceGroupRole;
};

export type SpaceGroupMembers = {
  public: boolean;
  members: SpaceGroupMember[];
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
  attributes: Attributes | null;
  space: Ref<SpaceId> | null;
  position: Position | null;
};

export type Switch = {
  id: SwitchId;
  name: string;
  display_order: number;
  attributes: Attributes | null;
  switch_panel: Ref<SwitchPanelId> | null;
};

export type Device = {
  id: DeviceId;
  name: string;
  attributes: Attributes | null;
  space: Ref<SpaceId> | null;
  position: Position | null;
};

export type LinkInfo = {
  attributes: Attributes | null;
};

export type LinkedSwitch = Switch & {
  link: LinkInfo;
};

export type SwitchLink = {
  upstream_switch_id: SwitchId;
  downstream_switch_id: SwitchId;
  attributes: Attributes | null;
};

export type DeviceLink = {
  device_id: DeviceId;
  switch_id: SwitchId;
  attributes: Attributes | null;
};

export type ChainSwitch = Omit<Switch, "switch_panel"> & {
  switch_panel: SwitchPanel;
};

export type PowerChainGraph = {
  switches: ChainSwitch[];
  switch_links: SwitchLink[];
  devices: Device[];
  device_links: DeviceLink[];
};

export type Meta = {
  auth_required: boolean;
};
