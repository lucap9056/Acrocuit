import type { Attributes, DeviceId, Position, SwitchId } from "@/api/models";
import { empty, ok } from "@/lib/local-backend/response";
import { include, json, type Router } from "@/lib/local-backend/router";
import type { LocalBackend } from "@/lib/local-backend/storage/types";

type SetDeviceBody = {
  name?: string;
  position?: Position;
  attributes?: Attributes;
};
type PositionBody = { position: Position };
type AddSwitchBody = { switch_id: SwitchId; attributes?: Attributes };
type AttributesBody = { attributes: Attributes };

const device = "/api/devices/:deviceId";
const deviceSwitch = `${device}/switches/:switchId`;

export const devicesHandlers = (router: Router, { Devices }: LocalBackend) => {
  router.get(device, ({ request, params }) =>
    Devices.GetDevice(params.deviceId as DeviceId, include(request)).then(ok),
  );

  router.put(device, ({ request, params }) => {
    const { name, position, attributes } = json<SetDeviceBody>(request);
    return Devices.SetDevice(
      params.deviceId as DeviceId,
      name ?? null,
      position ?? null,
      attributes as Attributes,
    ).then(ok);
  });

  router.put(`${device}/position`, ({ request, params }) =>
    Devices.SetDevicePosition(
      params.deviceId as DeviceId,
      json<PositionBody>(request).position,
    ).then(ok),
  );

  router.delete(device, ({ params }) =>
    Devices.DelDevice(params.deviceId as DeviceId).then(empty),
  );

  router.get(`${device}/switches`, ({ request, params }) =>
    Devices.GetDeviceSwitches(
      params.deviceId as DeviceId,
      include(request),
    ).then(ok),
  );

  router.post(`${device}/switches`, ({ request, params }) => {
    const { switch_id, attributes } = json<AddSwitchBody>(request);
    return Devices.AddDeviceSwitch(
      params.deviceId as DeviceId,
      switch_id,
      attributes ?? {},
    ).then(ok);
  });

  router.put(deviceSwitch, ({ request, params }) =>
    Devices.SetDeviceSwitch(
      params.deviceId as DeviceId,
      params.switchId as SwitchId,
      json<AttributesBody>(request).attributes,
    ).then(ok),
  );

  router.delete(deviceSwitch, ({ params }) =>
    Devices.DelDeviceSwitch(
      params.deviceId as DeviceId,
      params.switchId as SwitchId,
    ).then(empty),
  );

  router.get(`${device}/upstream`, ({ params }) =>
    Devices.GetDeviceUpstream(params.deviceId as DeviceId).then(ok),
  );
};
