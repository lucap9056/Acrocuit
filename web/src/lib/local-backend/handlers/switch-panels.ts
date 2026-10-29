import type { Attributes, Position, SwitchPanelId } from "@/api/models";
import { empty, ok } from "@/lib/local-backend/response";
import { include, json, type Router } from "@/lib/local-backend/router";
import type { LocalBackend } from "@/lib/local-backend/storage/types";

type SetSwitchPanelBody = { name?: string; attributes?: Attributes };
type PositionBody = { position: Position };
type CreateSwitchBody = { name: string; attributes?: Attributes };

const switchPanel = "/api/switch-panels/:switchPanelId";

export const switchPanelsHandlers = (
  router: Router,
  { SwitchPanels, Switches }: LocalBackend,
) => {
  router.get(switchPanel, ({ request, params }) =>
    SwitchPanels.GetSwitchPanel(
      params.switchPanelId as SwitchPanelId,
      include(request),
    ).then(ok),
  );

  router.put(switchPanel, ({ request, params }) => {
    const { name, attributes } = json<SetSwitchPanelBody>(request);
    return SwitchPanels.SetSwitchPanel(
      params.switchPanelId as SwitchPanelId,
      name ?? null,
      attributes as Attributes,
    ).then(ok);
  });

  router.put(`${switchPanel}/position`, ({ request, params }) =>
    SwitchPanels.SetSwitchPanelPosition(
      params.switchPanelId as SwitchPanelId,
      json<PositionBody>(request).position,
    ).then(ok),
  );

  router.delete(switchPanel, ({ params }) =>
    SwitchPanels.DelSwitchPanel(params.switchPanelId as SwitchPanelId).then(
      empty,
    ),
  );

  router.get(`${switchPanel}/switches`, ({ request, params }) =>
    Switches.GetSwitches(
      params.switchPanelId as SwitchPanelId,
      include(request),
    ).then(ok),
  );

  router.post(`${switchPanel}/switches`, ({ request, params }) => {
    const { name, attributes } = json<CreateSwitchBody>(request);
    return Switches.AddSwitch(
      params.switchPanelId as SwitchPanelId,
      name,
      attributes ?? {},
    ).then(ok);
  });
};
