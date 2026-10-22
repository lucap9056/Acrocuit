<script lang="ts">
  import FloppyDiskIcon from "phosphor-svelte/lib/FloppyDiskIcon";
  import { untrack } from "svelte";
  import type {
    AuthSession,
    Space,
    SpaceGroup,
    SpaceId,
    Switch,
    SwitchId,
    SwitchPanel,
    SwitchPanelId,
  } from "@/api";
  import api from "@/api";
  import load from "@/components/loading-bar";
  import Spinner from "@/components/spinner/spinner.svelte";
  import SwitchSelector from "@/components/switch-selector/switch-selector.svelte";
  import Textbox from "@/components/textbox/textbox.svelte";
  import { tooltip } from "@/components/tooltip";
  import { renameSwitch, setSwitchUpstream } from "@/lib/cache/switches";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import styles from "./switch-editor.module.css";

  interface Props {
    session: AuthSession;
    spaceGroup: SpaceGroup;
    space: Space;
    switchPanel: SwitchPanel;
    sw: Switch;
  }

  interface UpstreamSelection {
    spaceId: SpaceId;
    switchPanelId: SwitchPanelId | null;
    switchId: SwitchId | null;
  }

  const { session, spaceGroup, space, switchPanel, sw }: Props = $props();

  let current = $state(untrack(() => sw));
  let downstreamIds = $state<SwitchId[]>(untrack(() => [sw.id]));
  let isReady = $state(false);
  let isPending = $state(false);
  let spaceId = $state<SpaceId>(untrack(() => space.id));
  let switchPanelId = $state<SwitchPanelId | null>(null);
  let switchId = $state<SwitchId | null>(null);

  const isDirty = $derived(
    switchId !== (current.upstream_switch?.id ?? null),
  );

  const withSwitchPanel = (target: Switch) => ({
    ...target,
    switch_panel: { id: switchPanel.id },
  });

  const fetchUpstreamSelection = async (
    target: Switch,
  ): Promise<UpstreamSelection> => {
    const none = { spaceId: space.id, switchPanelId: null, switchId: null };
    if (!target.upstream_switch) {
      return none;
    }

    const upstream = await api.switches.getSwitch(
      session,
      target.upstream_switch.id,
    );
    if (!upstream.switch_panel) {
      return none;
    }

    const panel = await api.switchPanels.getSwitchPanel(
      session,
      upstream.switch_panel.id,
    );
    return {
      spaceId: panel.space?.id ?? space.id,
      switchPanelId: panel.id,
      switchId: upstream.id,
    };
  };

  const loadSelection = (signal?: AbortSignal) => {
    isReady = false;
    return fetchUpstreamSelection(current)
      .then((selection) => {
        if (signal?.aborted) {
          return;
        }
        spaceId = selection.spaceId;
        switchPanelId = selection.switchPanelId;
        switchId = selection.switchId;
        isReady = true;
      })
      .catch((err) => {
        console.error(err);
      });
  };

  onMount((signal) => {
    loadSelection(signal);

    api.switches
      .getSwitchDownstream(session, current.id)
      .then(({ switches }) => {
        if (!signal.aborted) {
          downstreamIds = switches.map((s) => s.id);
        }
      })
      .catch((err) => {
        console.error(err);
      });
  });

  const replaceCurrent = (switches: Switch[]) => {
    current = switches.find((s) => s.id === current.id) ?? current;
  };

  const rename = (name: string) => {
    const loading = load.begin(t("loading.renameSwitch"));
    return renameSwitch(session, withSwitchPanel(current), name)
      .then(replaceCurrent)
      .finally(() => {
        loading.release();
      });
  };

  const saveUpstream = () => {
    if (!isDirty || isPending) {
      return;
    }
    const target = switchId;
    isPending = true;
    const loading = load.begin(
      target === null
        ? t("loading.removeSwitchUpstream")
        : t("loading.setSwitchUpstream"),
    );
    setSwitchUpstream(session, withSwitchPanel(current), target)
      .then(replaceCurrent)
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        isPending = false;
        loading.release();
      });
  };
</script>

<div class={styles.editor}>
  <Textbox
    class={styles.title}
    label={t("switchEditor.name")}
    value={current.name}
    onsave={rename}
  />
  <section class={styles.field}>
    <h3 class={styles.heading}>{t("switchEditor.upstream")}</h3>
    {#if isReady}
      <div class={styles.upstream}>
        <div class={styles.selector}>
          <SwitchSelector
            {session}
            {spaceGroup}
            bind:spaceId
            bind:switchPanelId
            bind:switchId
            excludedSwitchIds={downstreamIds}
            disabled={isPending}
          />
        </div>
        <button
          class={["icon_button", styles.save, !isDirty && styles.hidden]}
          aria-label={t("common.save")}
          disabled={isPending || !isDirty}
          onclick={saveUpstream}
          {@attach tooltip(t("common.save"))}
        >
          <FloppyDiskIcon />
        </button>
      </div>
    {:else}
      <div class={styles.loading} role="status" aria-label={t("common.loading")}>
        <Spinner />
      </div>
    {/if}
  </section>
</div>
