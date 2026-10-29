import type { Attributes, Position, SpaceId } from "@/api/models";
import { empty, ok } from "@/lib/local-backend/response";
import { include, json, type Router } from "@/lib/local-backend/router";
import type { LocalBackend } from "@/lib/local-backend/storage/types";

type SetSpaceBody = { name?: string };
type OrderBody = { display_order: number };
type CreateBody = {
  name: string;
  position?: Position;
  attributes?: Attributes;
};

const space = "/api/spaces/:spaceId";

export const spacesHandlers = (
  router: Router,
  { Spaces, SwitchPanels, Devices }: LocalBackend,
) => {
  router.get(space, ({ request, params }) =>
    Spaces.GetSpace(params.spaceId as SpaceId, include(request)).then(ok),
  );

  router.put(space, ({ request, params }) =>
    Spaces.SetSpace(
      params.spaceId as SpaceId,
      json<SetSpaceBody>(request).name ?? null,
    ).then(ok),
  );

  router.delete(space, ({ params }) =>
    Spaces.DelSpace(params.spaceId as SpaceId).then(empty),
  );

  router.patch(`${space}/order`, ({ request, params }) =>
    Spaces.MoveSpace(
      params.spaceId as SpaceId,
      json<OrderBody>(request).display_order,
    ).then(ok),
  );

  router.get(`${space}/switch-panels`, ({ request, params }) =>
    SwitchPanels.GetSwitchPanels(
      params.spaceId as SpaceId,
      include(request),
    ).then(ok),
  );

  router.post(`${space}/switch-panels`, ({ request, params }) => {
    const { name, position, attributes } = json<CreateBody>(request);
    return SwitchPanels.AddSwitchPanel(
      params.spaceId as SpaceId,
      name,
      position ?? null,
      attributes ?? {},
    ).then(ok);
  });

  router.get(`${space}/devices`, ({ request, params }) =>
    Devices.GetDevices(params.spaceId as SpaceId, include(request)).then(ok),
  );

  router.post(`${space}/devices`, ({ request, params }) => {
    const { name, position, attributes } = json<CreateBody>(request);
    return Devices.AddDevice(
      params.spaceId as SpaceId,
      name,
      position ?? null,
      attributes ?? {},
    ).then(ok);
  });
};
