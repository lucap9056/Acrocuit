<script lang="ts">
  import CaretRightIcon from "phosphor-svelte/lib/CaretRightIcon";
  import LightbulbIcon from "phosphor-svelte/lib/LightbulbIcon";
  import ToggleRightIcon from "phosphor-svelte/lib/ToggleRightIcon";
  import { untrack } from "svelte";
  import type {
    AuthSession,
    Device,
    DeviceId,
    Space,
    SpaceGroup,
    Switch,
    SwitchId,
    SwitchPanel,
  } from "@/api";
  import api from "@/api";
  import back from "@/components/back-button";
  import dialog from "@/components/dialog";
  import headerTitle from "@/components/header-title";
  import load from "@/components/loading-bar";
  import {
    fetchDevices,
    setDevicePosition,
    watchDevices,
  } from "@/lib/cache/devices";
  import { removeFloorPlan, uploadFloorPlan } from "@/lib/cache/spaces";
  import {
    fetchSwitchPanels,
    setSwitchPanelPosition,
    watchSwitchPanels,
  } from "@/lib/cache/switch-panels";
  import { field, type HashRouter, type RouteParams } from "@/lib/hash-routers";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import { canEdit } from "@/lib/permissions";
  import type { ChainNode, ChainPlace, PowerChain } from "@/lib/power-chain";
  import { fetchDownstreamChain } from "@/lib/power-chain/downstream";
  import { fetchUpstreamChain } from "@/lib/power-chain/upstream";
  import type { Workspace } from "@/lib/workspace";
  import DeviceList from "./device-list/device-list.svelte";
  import EntityEditors from "./entity-editors/entity-editors.svelte";
  import FloorPlanCanvas from "./floor-plan-canvas/floor-plan-canvas.svelte";
  import { type PositionDraft, PositionDrafts } from "./position-drafts.svelte";
  import PowerChainView from "./power-chain/power-chain-view.svelte";
  import { SpaceFocus } from "./space-focus.svelte";
  import SpaceInfo from "./space-info/space-info.svelte";
  import styles from "./space-view.module.css";
  import SwitchPanelList from "./switch-panel-list/switch-panel-list.svelte";

  interface Props {
    session: AuthSession;
    workspace: Workspace;
    router: HashRouter;
    spaceGroup: SpaceGroup;
    space: Space;
    onspacesupdate: (spaces: Space[]) => void;
  }

  type ChainKind = "upstream" | "downstream";

  interface ChainRoute {
    kind: ChainKind;
    key: string;
    name: string;
    router: HashRouter;
    load: () => Promise<PowerChain>;
    back: () => void;
  }

  const {
    session,
    workspace,
    router,
    spaceGroup,
    space,
    onspacesupdate,
  }: Props = $props();

  let switchPanels = $state<SwitchPanel[]>();
  let devices = $state<Device[]>();
  let isUpdatingFloorPlan = $state(false);
  let chainRoute = $state.raw<ChainRoute>();
  let chainView = $state<ReturnType<typeof PowerChainView>>();

  const editable = $derived(canEdit(spaceGroup));

  const spaceRouter = untrack(() => router.child(`spaces/${space.id}`));

  const focus = new SpaceFocus();
  const drafts = new PositionDrafts();

  const withSpace = <T>(target: T) => ({
    ...target,
    space: { id: space.id },
  });

  const savePosition = (draft: PositionDraft): Promise<void> =>
    draft.kind === "switch_panel"
      ? setSwitchPanelPosition(
          session,
          withSpace(draft.target),
          draft.position,
        ).then((bgs) => {
          switchPanels = bgs;
        })
      : setDevicePosition(
          session,
          withSpace(draft.target),
          draft.position,
        ).then((ds) => {
          devices = ds;
        });

  const onsavepositions = () => {
    drafts.isSaving = true;
    const loading = load.begin(t("loading.savePositions"));
    [...drafts.entries.values()]
      .reduce(
        (saving, draft) => saving.then(() => savePosition(draft)),
        Promise.resolve(),
      )
      .then(() => {
        drafts.clear();
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        drafts.isSaving = false;
        loading.release();
      });
  };

  const withSpaceGroup = (target: Space) => ({
    ...target,
    space_group: { id: spaceGroup.id },
  });

  const updateFloorPlan = (
    message: string,
    request: () => Promise<Space[]>,
  ) => {
    isUpdatingFloorPlan = true;
    const loading = load.begin(message);
    request()
      .then(onspacesupdate)
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        isUpdatingFloorPlan = false;
        loading.release();
      });
  };

  const onuploadfloorplan = (image: File) => {
    updateFloorPlan(t("loading.uploadFloorPlan"), () =>
      uploadFloorPlan(session, withSpaceGroup(space), image),
    );
  };

  const onremovefloorplan = () => {
    dialog.push({
      text: t("floorPlan.confirmRemove", { name: space.name }),
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => true,
        },
        {
          text: t("common.remove"),
          type: "",
          onclick: () => {
            updateFloorPlan(t("loading.removeFloorPlan"), () =>
              removeFloorPlan(session, withSpaceGroup(space)),
            );
            return true;
          },
        },
      ],
    });
  };

  const oneditswitchpanel = (switchPanel: SwitchPanel) => {
    spaceRouter.to(`switch-panels/${switchPanel.id}`);
  };

  const oneditdevice = (device: Device) => {
    spaceRouter.to(`devices/${device.id}`);
  };

  const onshowupstream = (device: Device) => {
    spaceRouter.to(`devices/${device.id}/upstream`);
  };

  const onshowdownstream = (sw: Switch) => {
    spaceRouter.to(`switches/${sw.id}/downstream`);
  };

  const CHAIN_ORIGIN = "chain";

  const onopenchainnode = (chainRouter: HashRouter, node: ChainNode) => {
    chainRouter.to(
      node.kind === "switch"
        ? `switch-panels/${node.panel.id}/switches/${node.sw.id}`
        : `devices/${node.device.id}`,
      { origin: CHAIN_ORIGIN },
    );
  };

  const onopenchainplace = (
    chainRouter: HashRouter,
    { target }: ChainPlace,
  ) => {
    if (target.kind === "device") {
      onopenchainnode(chainRouter, target);
    } else {
      chainRouter.to(`chain-panels/${target.panel.id}`, {
        origin: CHAIN_ORIGIN,
      });
    }
  };

  const idParam = <Id extends string>(params: RouteParams, name: string) =>
    field(params, name, "") as Id;

  const routeChain = (
    signal: AbortSignal,
    kind: ChainKind,
    path: string,
    resolve: (params: RouteParams) => Promise<ChainRoute | undefined>,
  ) => {
    let latestRequest = 0;
    const handle = (params?: RouteParams) => {
      const request = ++latestRequest;
      if (params === undefined) {
        if (chainRoute?.kind === kind) {
          chainRoute = undefined;
        }
        return;
      }

      const loading = load.begin(t("loading.fetchPowerChain"));
      resolve(params)
        .then((route) => {
          if (signal.aborted || request !== latestRequest) {
            return;
          }
          chainRoute = route;
          if (!route) {
            spaceRouter.back("");
          }
        })
        .catch((err) => {
          console.error(err);
        })
        .finally(() => {
          loading.release();
        });
    };
    handle(spaceRouter.on(path, handle, signal));
  };

  const routeChains = (signal: AbortSignal) => {
    routeChain(
      signal,
      "upstream",
      "devices/:deviceId/upstream/*",
      async (params) => {
        const deviceId = idParam<DeviceId>(params, "deviceId");
        const device = (await fetchDevices(session, space.id)).find(
          ({ id }) => id === deviceId,
        );
        if (!device) {
          return undefined;
        }
        return {
          kind: "upstream",
          key: `upstream:${device.id}`,
          name: device.name,
          router: spaceRouter.child(`devices/${device.id}/upstream`),
          load: () => fetchUpstreamChain(session, device, space.id),
          back: () => spaceRouter.backTo(`devices/${device.id}`),
        };
      },
    );

    routeChain(
      signal,
      "downstream",
      "switches/:switchId/downstream/*",
      async (params) => {
        const sw = await api.switches.getSwitch(
          session,
          idParam<SwitchId>(params, "switchId"),
        );
        const panelId = sw.switch_panel?.id;
        return {
          kind: "downstream",
          key: `downstream:${sw.id}`,
          name: sw.name,
          router: spaceRouter.child(`switches/${sw.id}/downstream`),
          load: () => fetchDownstreamChain(session, sw, space.id),
          back: () =>
            spaceRouter.backTo(
              panelId === undefined
                ? ""
                : `switch-panels/${panelId}/switches/${sw.id}`,
            ),
        };
      },
    );
  };

  onMount((signal) => {
    workspace.push(
      {
        leftpart: spaceSidebar,
        centerpart: spaceContent,
        rightpart: spaceInfo,
      },
      signal,
    );
    back.push(() => router.backTo(""), signal);
    headerTitle.push(title, signal);
    routeChains(signal);
    watchSwitchPanels(
      session,
      space.id,
      (bgs) => {
        switchPanels = bgs;
      },
      signal,
    );
    watchDevices(
      session,
      space.id,
      (ds) => {
        devices = ds;
      },
      signal,
    );

    const loading = load.begin(t("loading.fetchSwitchesAndDevices"));
    Promise.all([
      fetchSwitchPanels(session, space.id).then((bgs) => {
        if (!signal.aborted) {
          switchPanels = bgs;
        }
      }),
      fetchDevices(session, space.id).then((ds) => {
        if (!signal.aborted) {
          devices = ds;
        }
      }),
    ])
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
      });
  });
</script>

{#snippet title()}
  <div class={styles.breadcrumb}>
    <button
      class={[styles.crumb, styles.parent]}
      onclick={() => router.backTo("")}
    >
      {spaceGroup.name}
    </button>
    <CaretRightIcon class={styles.separator} aria-hidden="true" />
    <span class={styles.crumb}>{space.name}</span>
  </div>
{/snippet}

{#snippet chainTitle()}
  <div class={styles.breadcrumb}>
    <button
      class={[styles.crumb, styles.parent]}
      onclick={() => spaceRouter.backTo("")}
    >
      {space.name}
    </button>
    <CaretRightIcon class={styles.separator} aria-hidden="true" />
    <span class={styles.crumb}
      >{chainRoute
        ? t(`powerChain.${chainRoute.kind}Title`, { name: chainRoute.name })
        : ""}</span
    >
  </div>
{/snippet}

{#snippet spaceSidebar()}
  <div class={styles.sidebar}>
    <SwitchPanelList
      {session}
      {space}
      {focus}
      bind:switchPanels
      {editable}
      onedit={oneditswitchpanel}
    />
    <DeviceList
      {session}
      {space}
      {focus}
      bind:devices
      {editable}
      onedit={oneditdevice}
    />
  </div>
{/snippet}

{#snippet spaceContent()}
  <FloorPlanCanvas
    {session}
    {space}
    {switchPanels}
    {devices}
    {focus}
    {drafts}
    {editable}
    {onsavepositions}
    {oneditswitchpanel}
    {oneditdevice}
  />
{/snippet}

{#snippet spaceInfo()}
  <SpaceInfo
    {space}
    {switchPanels}
    {devices}
    {isUpdatingFloorPlan}
    {editable}
    {onuploadfloorplan}
    {onremovefloorplan}
  />
{/snippet}

{#snippet deviceIcon()}
  <LightbulbIcon />
{/snippet}

{#snippet switchIcon()}
  <ToggleRightIcon />
{/snippet}

<EntityEditors
  {session}
  {spaceGroup}
  router={spaceRouter}
  locate={{ device: () => space.id, switchPanel: () => space.id }}
  {onshowupstream}
  {onshowdownstream}
/>

{#each chainRoute ? [chainRoute] : [] as route (route.key)}
  <PowerChainView
    bind:this={chainView}
    {session}
    {workspace}
    {spaceGroup}
    title={chainTitle}
    focusIcon={route.kind === "upstream" ? deviceIcon : switchIcon}
    emptyText={t(`powerChain.${route.kind}Empty`)}
    load={route.load}
    onback={route.back}
    onopennode={(node) => onopenchainnode(route.router, node)}
    onopenplace={(place) => onopenchainplace(route.router, place)}
  />
  <EntityEditors
    {session}
    {spaceGroup}
    router={route.router}
    locate={{
      device: (id) => chainView?.locate.device(id),
      switchPanel: (id) => chainView?.locate.switchPanel(id),
    }}
    isChainSwitch={(sw) => chainView?.isChainSwitch(sw) ?? false}
    onclose={() => chainView?.reload()}
  />
{/each}
