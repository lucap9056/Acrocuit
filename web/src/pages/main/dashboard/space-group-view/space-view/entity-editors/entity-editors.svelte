<script lang="ts">
  import type {
    Attributes,
    AuthSession,
    Device,
    DeviceId,
    Ref,
    SpaceGroup,
    SpaceId,
    Switch,
    SwitchPanel,
    SwitchPanelId,
  } from "@/api";
  import api from "@/api";
  import { routeDialog } from "@/components/dialog/route-dialog";
  import load from "@/components/loading-bar";
  import {
    fetchDevices,
    renameDevice,
    setDeviceAttributes,
  } from "@/lib/cache/devices";
  import {
    fetchSwitchPanels,
    renameSwitchPanel,
    setSwitchPanelAttributes,
  } from "@/lib/cache/switch-panels";
  import { fetchSwitches } from "@/lib/cache/switches";
  import { field, type HashRouter, type RouteParams } from "@/lib/hash-routers";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import { canEdit } from "@/lib/permissions";
  import DeviceEditor from "../device-editor/device-editor.svelte";
  import SwitchEditor from "../switch-editor/switch-editor.svelte";
  import SwitchPanelEditor, {
    type ReadonlyView,
  } from "../switch-panel-editor/switch-panel-editor.svelte";

  type Placed<T> = T & { space: Ref<SpaceId> };

  interface SpaceLocator {
    device: (id: DeviceId) => SpaceId | undefined;
    switchPanel: (id: SwitchPanelId) => SpaceId | undefined;
  }

  interface Props {
    session: AuthSession;
    spaceGroup: SpaceGroup;
    router: HashRouter;
    locate: SpaceLocator;
    isChainSwitch?: (sw: Switch) => boolean;
    onshowupstream?: (device: Device) => void;
    onshowdownstream?: (sw: Switch) => void;
    onclose?: () => void;
  }

  const {
    session,
    spaceGroup,
    router,
    locate,
    isChainSwitch,
    onshowupstream,
    onshowdownstream,
    onclose,
  }: Props = $props();

  const editable = $derived(canEdit(spaceGroup));

  let editingDevice = $state<Placed<Device>>();
  let editingSwitchPanel = $state<Placed<SwitchPanel>>();
  let editingSwitch = $state<Switch>();

  const placed = <T extends { space: Ref<SpaceId> | null }>(
    target: T | undefined,
  ): Placed<T> | undefined =>
    target?.space ? { ...target, space: target.space } : undefined;

  const idParam = (params: RouteParams, name: string) =>
    parseInt(field(params, name, ""), 10);

  const fromSpace = <T extends { id: unknown; space: Ref<SpaceId> | null }>(
    items: T[],
    id: T["id"],
    spaceId: SpaceId,
  ) => {
    const found = items.find((item) => item.id === id);
    return found && placed({ ...found, space: found.space ?? { id: spaceId } });
  };

  const resolveDevice = async (id: DeviceId) => {
    const spaceId = locate.device(id);
    const cached =
      spaceId === undefined
        ? undefined
        : fromSpace(await fetchDevices(session, spaceId), id, spaceId);
    return cached ?? placed(await api.devices.getDevice(session, id));
  };

  const resolveSwitchPanel = async (id: SwitchPanelId) => {
    const spaceId = locate.switchPanel(id);
    const cached =
      spaceId === undefined
        ? undefined
        : fromSpace(await fetchSwitchPanels(session, spaceId), id, spaceId);
    return cached ?? placed(await api.switchPanels.getSwitchPanel(session, id));
  };

  const withLoading = <T>(message: string, request: Promise<T>) => {
    const loading = load.begin(message);
    return request.finally(() => {
      loading.release();
    });
  };

  const onrenamedevice = (device: Placed<Device>, name: string) =>
    withLoading(
      t("loading.renameDevice"),
      renameDevice(session, device, name),
    ).then((devices) => {
      editingDevice = placed(devices.find(({ id }) => id === device.id));
    });

  const onsavedeviceattributes = (
    device: Placed<Device>,
    attributes: Attributes,
  ) =>
    withLoading(
      t("loading.updateDeviceAttributes"),
      setDeviceAttributes(session, device, attributes),
    ).then((devices) => {
      editingDevice = placed(devices.find(({ id }) => id === device.id));
    });

  const onrenameswitchpanel = (panel: Placed<SwitchPanel>, name: string) =>
    withLoading(
      t("loading.renameSwitchPanel"),
      renameSwitchPanel(session, panel, name),
    ).then((panels) => {
      editingSwitchPanel = placed(panels.find(({ id }) => id === panel.id));
    });

  const onsaveswitchpanelattributes = (
    panel: Placed<SwitchPanel>,
    attributes: Attributes,
  ) =>
    withLoading(
      t("loading.updateSwitchPanelAttributes"),
      setSwitchPanelAttributes(session, panel, attributes),
    ).then((panels) => {
      editingSwitchPanel = placed(panels.find(({ id }) => id === panel.id));
    });

  const oneditswitch = (sw: Switch) => {
    const switchPanelId = sw.switch_panel?.id ?? editingSwitchPanel?.id;
    if (switchPanelId !== undefined) {
      router.to(`switch-panels/${switchPanelId}/switches/${sw.id}`);
    }
  };

  const onopenchainswitch = (sw: Switch) => {
    if (editingSwitchPanel) {
      router.to(`chain-panels/${editingSwitchPanel.id}/switches/${sw.id}`);
    }
  };

  const chainView = $derived<ReadonlyView | undefined>(
    isChainSwitch && {
      hint: t("switchPanels.chainReadonly"),
      includes: isChainSwitch,
    },
  );

  onMount((signal) => {
    let isClosePending = false;
    const notifyClose = () => {
      if (isClosePending || signal.aborted) {
        return;
      }
      isClosePending = true;
      queueMicrotask(() => {
        isClosePending = false;
        if (!signal.aborted) {
          onclose?.();
        }
      });
    };

    const routeSwitchEditor = (
      panelRouter: HashRouter,
      opened: AbortSignal,
    ) => {
      if (!editingSwitchPanel) {
        return;
      }
      const switchPanelId = editingSwitchPanel.id;
      routeDialog({
        router: panelRouter,
        path: "switches/:switchId",
        closeTo: "",
        loading: t("loading.fetchSwitch"),
        open: async (switchParams, switchOpened) => {
          const switchId = idParam(switchParams, "switchId");
          const sw = (await fetchSwitches(session, switchPanelId)).find(
            ({ id }) => id === switchId,
          );
          if (!sw) {
            return undefined;
          }
          editingSwitch = sw;
          switchOpened.addEventListener("abort", notifyClose, { once: true });
          return {
            content: switchEditor,
            actions: onshowdownstream
              ? [
                  {
                    text: t("switchEditor.showDownstreamChain"),
                    type: "",
                    align: "start",
                    onclick: () => {
                      onshowdownstream(sw);
                      return false;
                    },
                  },
                ]
              : [],
          };
        },
        signal: opened,
      });
    };

    routeDialog({
      router,
      path: "devices/:deviceId",
      closeTo: "",
      loading: t("loading.fetchDevice"),
      open: async (params, opened) => {
        const device = await resolveDevice(
          idParam(params, "deviceId") as DeviceId,
        );
        if (!device) {
          return undefined;
        }
        editingDevice = device;
        opened.addEventListener("abort", notifyClose, { once: true });
        return {
          content: deviceEditor,
          actions: onshowupstream
            ? [
                {
                  text: t("deviceEditor.showUpstreamChain"),
                  type: "",
                  align: "start",
                  onclick: () => {
                    onshowupstream(device);
                    return false;
                  },
                },
              ]
            : [],
        };
      },
      signal,
    });

    routeDialog({
      router,
      path: "switch-panels/:switchPanelId/*",
      closeTo: "",
      loading: t("loading.fetchSwitchPanel"),
      open: async (params, opened) => {
        const switchPanel = await resolveSwitchPanel(
          idParam(params, "switchPanelId") as SwitchPanelId,
        );
        if (!switchPanel) {
          return undefined;
        }
        editingSwitchPanel = switchPanel;
        opened.addEventListener("abort", notifyClose, { once: true });
        return { content: switchPanelEditor };
      },
      onopened: (opened) => {
        if (editingSwitchPanel) {
          routeSwitchEditor(
            router.child(`switch-panels/${editingSwitchPanel.id}`),
            opened,
          );
        }
      },
      signal,
    });

    if (isChainSwitch) {
      routeDialog({
        router,
        path: "chain-panels/:switchPanelId/*",
        closeTo: "",
        loading: t("loading.fetchSwitchPanel"),
        open: async (params, opened) => {
          const switchPanel = await resolveSwitchPanel(
            idParam(params, "switchPanelId") as SwitchPanelId,
          );
          if (!switchPanel) {
            return undefined;
          }
          editingSwitchPanel = switchPanel;
          opened.addEventListener("abort", notifyClose, { once: true });
          return { content: chainPanelDialog };
        },
        onopened: (opened) => {
          if (editingSwitchPanel) {
            routeSwitchEditor(
              router.child(`chain-panels/${editingSwitchPanel.id}`),
              opened,
            );
          }
        },
        signal,
      });
    }
  });
</script>

{#snippet switchPanelEditor()}
  {#if editingSwitchPanel}
    {@const switchPanel = editingSwitchPanel}
    <SwitchPanelEditor
      {session}
      {switchPanel}
      {editable}
      onrename={(name) => onrenameswitchpanel(switchPanel, name)}
      onsaveattributes={(attributes) =>
        onsaveswitchpanelattributes(switchPanel, attributes)}
      {oneditswitch}
    />
  {/if}
{/snippet}

{#snippet chainPanelDialog()}
  {#if editingSwitchPanel && chainView}
    <SwitchPanelEditor
      {session}
      switchPanel={editingSwitchPanel}
      readonly={chainView}
      oneditswitch={onopenchainswitch}
    />
  {/if}
{/snippet}

{#snippet switchEditor()}
  {#if editingSwitchPanel && editingSwitch}
    <SwitchEditor
      {session}
      {spaceGroup}
      spaceId={editingSwitchPanel.space.id}
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
      spaceId={device.space.id}
      {device}
      onrename={(name) => onrenamedevice(device, name)}
      onsaveattributes={(attributes) =>
        onsavedeviceattributes(device, attributes)}
    />
  {/if}
{/snippet}
