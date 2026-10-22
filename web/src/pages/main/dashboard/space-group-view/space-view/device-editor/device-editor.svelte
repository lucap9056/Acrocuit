<script lang="ts">
  import PlusIcon from "phosphor-svelte/lib/PlusIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import { untrack } from "svelte";
  import type {
    AuthSession,
    Device,
    Space,
    SpaceGroup,
    SpaceId,
    Switch,
    SwitchId,
    SwitchPanelId,
  } from "@/api";
  import load from "@/components/loading-bar";
  import Spinner from "@/components/spinner/spinner.svelte";
  import SwitchSelector from "@/components/switch-selector/switch-selector.svelte";
  import Textbox from "@/components/textbox/textbox.svelte";
  import { tooltip } from "@/components/tooltip";
  import {
    addDeviceSwitch,
    fetchDeviceSwitches,
    removeDeviceSwitch,
  } from "@/lib/cache/device-switches";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import styles from "./device-editor.module.css";

  interface Props {
    session: AuthSession;
    spaceGroup: SpaceGroup;
    space: Space;
    device: Device;
    onrename: (name: string) => Promise<unknown>;
  }

  const { session, spaceGroup, space, device, onrename }: Props = $props();

  let upstreams = $state<Switch[]>();
  let isPending = $state(false);
  let spaceId = $state<SpaceId>(untrack(() => space.id));
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

  const update = (message: string, request: () => Promise<Switch[]>) => {
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
  <section class={styles.field}>
    <h3 class={styles.heading}>{t("deviceEditor.upstream")}</h3>
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
  </section>
</div>
