<script lang="ts">
  import CaretRightIcon from "phosphor-svelte/lib/CaretRightIcon";
  import { untrack } from "svelte";
  import type {
    AuthSession,
    Switch,
    SwitchPanel,
    Device,
    Space,
    SpaceGroup,
  } from "@/api";
  import back from "@/components/back-button";
  import dialog from "@/components/dialog";
  import { routeDialog } from "@/components/dialog/route-dialog";
  import headerTitle from "@/components/header-title";
  import load from "@/components/loading-bar";
  import {
    fetchSwitchPanels,
    renameSwitchPanel,
    setSwitchPanelPosition,
  } from "@/lib/cache/switch-panels";
  import {
    fetchDevices,
    renameDevice,
    setDevicePosition,
  } from "@/lib/cache/devices";
  import { removeFloorPlan, uploadFloorPlan } from "@/lib/cache/spaces";
  import { fetchSwitches } from "@/lib/cache/switches";
  import { field, type HashRouter, type RouteParams } from "@/lib/hash-routers";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import type { Workspace } from "@/lib/workspace";
  import SwitchEditor from "./switch-editor/switch-editor.svelte";
  import SwitchPanelEditor from "./switch-panel-editor/switch-panel-editor.svelte";
  import SwitchPanelList from "./switch-panel-list/switch-panel-list.svelte";
  import DeviceList from "./device-list/device-list.svelte";
  import DeviceEditor from "./device-editor/device-editor.svelte";
  import FloorPlanCanvas from "./floor-plan-canvas/floor-plan-canvas.svelte";
  import { type PositionDraft, PositionDrafts } from "./position-drafts.svelte";
  import { SpaceFocus } from "./space-focus.svelte";
  import SpaceInfo from "./space-info/space-info.svelte";
  import styles from "./space-view.module.css";

  interface Props {
    session: AuthSession;
    workspace: Workspace;
    router: HashRouter;
    spaceGroup: SpaceGroup;
    space: Space;
    onspacesupdate: (spaces: Space[]) => void;
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
  let editingSwitchPanel = $state<SwitchPanel>();
  let editingSwitch = $state<Switch>();
  let editingDevice = $state<Device>();

  const spaceRouter = untrack(() => router.child(`space/${space.id}`));

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

  const onrenameswitchpanel = (switchPanel: SwitchPanel, name: string) => {
    const loading = load.begin(t("loading.renameSwitchPanel"));
    return renameSwitchPanel(session, withSpace(switchPanel), name)
      .then((sps) => {
        switchPanels = sps;
        editingSwitchPanel = sps.find((sp) => sp.id === switchPanel.id);
      })
      .finally(() => {
        loading.release();
      });
  };

  const oneditswitchpanel = (switchPanel: SwitchPanel) => {
    spaceRouter.to(`switch-panels/${switchPanel.id}`);
  };

  const oneditswitch = (sw: Switch) => {
    const switchPanelId = sw.switch_panel?.id ?? editingSwitchPanel?.id;
    if (switchPanelId !== undefined) {
      spaceRouter.to(`switch-panels/${switchPanelId}/switches/${sw.id}`);
    }
  };

  const onrenamedevice = (device: Device, name: string) => {
    const loading = load.begin(t("loading.renameDevice"));
    return renameDevice(session, withSpace(device), name)
      .then((ds) => {
        devices = ds;
        editingDevice = ds.find((d) => d.id === device.id);
      })
      .finally(() => {
        loading.release();
      });
  };

  const oneditdevice = (device: Device) => {
    spaceRouter.to(`devices/${device.id}`);
  };

  const idParam = (params: RouteParams, name: string) =>
    parseInt(field(params, name, ""), 10);

  const routeEditors = (signal: AbortSignal) => {
    routeDialog({
      router: spaceRouter,
      path: "devices/:deviceId",
      closeTo: "",
      open: async (params) => {
        const deviceId = idParam(params, "deviceId");
        const device = (await fetchDevices(session, space.id)).find(
          ({ id }) => id === deviceId,
        );
        if (!device) {
          return undefined;
        }
        editingDevice = device;
        return { content: deviceEditor };
      },
      signal,
    });

    routeDialog({
      router: spaceRouter,
      path: "switch-panels/:switchPanelId/*",
      closeTo: "",
      open: async (params) => {
        const switchPanelId = idParam(params, "switchPanelId");
        const switchPanel = (await fetchSwitchPanels(session, space.id)).find(
          ({ id }) => id === switchPanelId,
        );
        if (!switchPanel) {
          return undefined;
        }
        editingSwitchPanel = switchPanel;
        return { content: switchPanelEditor };
      },
      onopened: (opened) => {
        if (!editingSwitchPanel) {
          return;
        }
        const switchPanelId = editingSwitchPanel.id;
        routeDialog({
          router: spaceRouter.child(`switch-panels/${switchPanelId}`),
          path: "switches/:switchId",
          closeTo: "",
          open: async (switchParams) => {
            const switchId = idParam(switchParams, "switchId");
            const sw = (await fetchSwitches(session, switchPanelId)).find(
              ({ id }) => id === switchId,
            );
            if (!sw) {
              return undefined;
            }
            editingSwitch = sw;
            return { content: switchEditor };
          },
          signal: opened,
        });
      },
      signal,
    });
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
    routeEditors(signal);

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

{#snippet switchPanelEditor()}
  {#if editingSwitchPanel}
    {@const switchPanel = editingSwitchPanel}
    <SwitchPanelEditor
      {session}
      {switchPanel}
      onrename={(name) => onrenameswitchpanel(switchPanel, name)}
      {oneditswitch}
    />
  {/if}
{/snippet}

{#snippet switchEditor()}
  {#if editingSwitchPanel && editingSwitch}
    <SwitchEditor
      {session}
      {spaceGroup}
      {space}
      switchPanel={editingSwitchPanel}
      sw={editingSwitch}
    />
  {/if}
{/snippet}

{#snippet deviceEditor()}
  {#if editingDevice}
    {@const device = editingDevice}
    <DeviceEditor
      {session}
      {spaceGroup}
      {space}
      {device}
      onrename={(name) => onrenamedevice(device, name)}
    />
  {/if}
{/snippet}

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

{#snippet spaceSidebar()}
  <div class={styles.sidebar}>
    <SwitchPanelList
      {session}
      {space}
      {focus}
      bind:switchPanels
      onedit={oneditswitchpanel}
    />
    <DeviceList
      {session}
      {space}
      {focus}
      bind:devices
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
    {onuploadfloorplan}
    {onremovefloorplan}
  />
{/snippet}
