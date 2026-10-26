<script lang="ts">
  import BracketsCurlyIcon from "phosphor-svelte/lib/BracketsCurlyIcon";
  import PlusIcon from "phosphor-svelte/lib/PlusIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import { untrack } from "svelte";
  import type {
    Attributes,
    AuthSession,
    Device,
    LinkedSwitch,
    SpaceGroup,
    SpaceId,
    Switch,
    SwitchId,
    SwitchPanelId,
  } from "@/api";
  import AttributesSection from "@/components/attributes-section/attributes-section.svelte";
  import CollapsibleSection from "@/components/collapsible-section/collapsible-section.svelte";
  import { editLinkAttributes } from "@/components/link-attributes-dialog/link-attributes-dialog.svelte";
  import load from "@/components/loading-bar";
  import Spinner from "@/components/spinner/spinner.svelte";
  import SwitchSelector from "@/components/switch-selector/switch-selector.svelte";
  import Textbox from "@/components/textbox/textbox.svelte";
  import { tooltip } from "@/components/tooltip";
  import { DEVICE_ATTRIBUTE_FIELDS } from "@/lib/attributes";
  import {
    addDeviceSwitch,
    fetchDeviceSwitches,
    removeDeviceSwitch,
    setDeviceSwitchAttributes,
  } from "@/lib/cache/device-switches";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import styles from "./device-editor.module.css";

  interface Props {
    session: AuthSession;
    spaceGroup: SpaceGroup;
    spaceId: SpaceId;
    device: Device;
    onrename: (name: string) => Promise<unknown>;
    onsaveattributes: (attributes: Attributes) => Promise<unknown>;
  }

  const {
    session,
    spaceGroup,
    spaceId: initialSpaceId,
    device,
    onrename,
    onsaveattributes,
  }: Props = $props();

  let upstreams = $state<LinkedSwitch[]>();
  let isPending = $state(false);
  let spaceId = $state<SpaceId>(untrack(() => initialSpaceId));
  let switchPanelId = $state<SwitchPanelId | null>(null);
  let switchId = $state<SwitchId | null>(null);

  const upstreamIds = $derived(upstreams?.map((sw) => sw.id) ?? []);

  onMount((signal) => {
    fetchDeviceSwitches(session, device.id)
      .then((sws) => {
        if (!signal.aborted) {
          upstreams = sws;
        }
      })
      .catch((err) => {
        console.error(err);
      });
  });

  const update = (message: string, request: () => Promise<LinkedSwitch[]>) => {
    isPending = true;
    const loading = load.begin(message);
    return request()
      .then((sws) => {
        upstreams = sws;
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        isPending = false;
        loading.release();
      });
  };

  const add = () => {
    if (switchId === null) {
      return;
    }

    const target = switchId;
    update(t("loading.addDeviceUpstream"), () =>
      addDeviceSwitch(session, device.id, target),
    ).then(() => {
      switchId = null;
    });
  };

  const editUpstreamAttributes = (sw: LinkedSwitch) => {
    editLinkAttributes(sw.link.attributes, (attributes) => {
      isPending = true;
      const loading = load.begin(t("loading.updateDeviceUpstream"));
      return setDeviceSwitchAttributes(session, device.id, sw.id, attributes)
        .then((sws) => {
          upstreams = sws;
        })
        .finally(() => {
          isPending = false;
          loading.release();
        });
    });
  };

  const remove = (sw: Switch) => {
    update(t("loading.removeDeviceUpstream"), () =>
      removeDeviceSwitch(session, device.id, sw.id),
    );
  };
</script>

<div class={styles.editor}>
  <Textbox
    class={styles.title}
    label={t("deviceEditor.name")}
    value={device.name}
    onsave={onrename}
  />
  <CollapsibleSection title={t("deviceEditor.upstream")} open>
    <ul class={styles.rows}>
      <li class={styles.row}>
        <div class={styles.selector}>
          <SwitchSelector
            {session}
            {spaceGroup}
            bind:spaceId
            bind:switchPanelId
            bind:switchId
            excludedSwitchIds={upstreamIds}
          />
        </div>
        <button
          class="icon_button {styles.add}"
          aria-label={t("deviceEditor.addUpstream")}
          disabled={switchId === null || isPending}
          onclick={add}
          {@attach tooltip(t("deviceEditor.addUpstream"))}
        >
          <PlusIcon />
        </button>
      </li>
      {#if upstreams === undefined}
        <li
          class={styles.loading}
          role="status"
          aria-label={t("common.loading")}
        >
          <Spinner />
        </li>
      {:else if upstreams.length === 0}
        <li class={styles.message}>{t("common.noneYet")}</li>
      {:else}
        {#each upstreams as sw (sw.id)}
          <li class={styles.row}>
            <span class={styles.name}>{sw.name}</span>
            <button
              class="icon_button {styles.edit}"
              aria-label={t("attributes.edit")}
              disabled={isPending}
              onclick={() => editUpstreamAttributes(sw)}
              {@attach tooltip(t("attributes.edit"))}
            >
              <BracketsCurlyIcon />
            </button>
            <button
              class="icon_button {styles.delete}"
              aria-label={t("common.remove")}
              disabled={isPending}
              onclick={() => remove(sw)}
              {@attach tooltip(t("common.remove"))}
            >
              <TrashIcon />
            </button>
          </li>
        {/each}
      {/if}
    </ul>
  </CollapsibleSection>
  <AttributesSection
    fields={DEVICE_ATTRIBUTE_FIELDS}
    attributes={device.attributes}
    onsave={onsaveattributes}
  />
</div>
