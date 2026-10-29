import type { Attributes, SwitchId } from "@/api/models";
import { empty, ok } from "@/lib/local-backend/response";
import { include, json, type Router } from "@/lib/local-backend/router";
import type { LocalBackend } from "@/lib/local-backend/storage/types";

type SetSwitchBody = { name?: string; attributes?: Attributes };
type OrderBody = { display_order: number };
type AddUpstreamBody = {
  upstream_switch_id: SwitchId;
  attributes?: Attributes;
};
type AttributesBody = { attributes: Attributes };

const sw = "/api/switches/:switchId";
const upstream = `${sw}/upstreams/:upstreamSwitchId`;

export const switchesHandlers = (
  router: Router,
  { Switches }: LocalBackend,
) => {
  router.get(`${sw}/downstream`, ({ params }) =>
    Switches.GetSwitchDownstream(params.switchId as SwitchId).then(ok),
  );

  router.get(sw, ({ request, params }) =>
    Switches.GetSwitch(params.switchId as SwitchId, include(request)).then(ok),
  );

  router.put(sw, ({ request, params }) => {
    const { name, attributes } = json<SetSwitchBody>(request);
    return Switches.SetSwitch(
      params.switchId as SwitchId,
      name ?? null,
      attributes as Attributes,
    ).then(ok);
  });

  router.delete(sw, ({ params }) =>
    Switches.DelSwitch(params.switchId as SwitchId).then(empty),
  );

  router.patch(`${sw}/order`, ({ request, params }) =>
    Switches.MoveSwitch(
      params.switchId as SwitchId,
      json<OrderBody>(request).display_order,
    ).then(ok),
  );

  router.get(`${sw}/upstreams`, ({ request, params }) =>
    Switches.GetSwitchUpstreams(
      params.switchId as SwitchId,
      include(request),
    ).then(ok),
  );

  router.post(`${sw}/upstreams`, ({ request, params }) => {
    const { upstream_switch_id, attributes } = json<AddUpstreamBody>(request);
    return Switches.AddSwitchUpstream(
      params.switchId as SwitchId,
      upstream_switch_id,
      attributes ?? {},
    ).then(ok);
  });

  router.put(upstream, ({ request, params }) =>
    Switches.SetSwitchUpstream(
      params.switchId as SwitchId,
      params.upstreamSwitchId as SwitchId,
      json<AttributesBody>(request).attributes,
    ).then(ok),
  );

  router.delete(upstream, ({ params }) =>
    Switches.DelSwitchUpstream(
      params.switchId as SwitchId,
      params.upstreamSwitchId as SwitchId,
    ).then(empty),
  );
};
