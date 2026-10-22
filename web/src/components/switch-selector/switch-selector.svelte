<script lang="ts">
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
  import Select, { type SelectOption } from "@/components/select/select.svelte";
  import { fetchSpaces } from "@/lib/cache/spaces";
  import { fetchSwitchPanels } from "@/lib/cache/switch-panels";
  import { fetchSwitches } from "@/lib/cache/switches";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import styles from "./switch-selector.module.css";

  interface Props {
    session: AuthSession;
    spaceGroup: SpaceGroup;
    spaceId: SpaceId;
    switchPanelId: SwitchPanelId | null;
    switchId: SwitchId | null;
    excludedSwitchIds?: SwitchId[];
    disabled?: boolean;
  }

  let {
    session,
    spaceGroup,
    spaceId = $bindable(),
    switchPanelId = $bindable(),
    switchId = $bindable(),
    excludedSwitchIds = [],
    disabled = false,
  }: Props = $props();

  let spaces = $state<Space[]>();
  let switchPanels = $state<SwitchPanel[]>();
  let switches = $state<Switch[]>();

  const noneOption: SelectOption<null> = {
    value: null,
    label: t("common.none"),
  };

  const spaceOptions = $derived(
    spaces?.map((space) => ({ value: space.id, label: space.name })),
  );

  const switchPanelOptions = $derived(
    switchPanels && [
      noneOption,
      ...switchPanels.map((switchPanel) => ({
        value: switchPanel.id,
        label: switchPanel.name,
      })),
    ],
  );

  const switchOptions = $derived(
    switches && [
      noneOption,
      ...switches
        .filter((sw) => !excludedSwitchIds.includes(sw.id))
        .map((sw) => ({ value: sw.id, label: sw.name })),
    ],
  );

  onMount((signal) => {
    fetchSpaces(session, spaceGroup.id)
      .then((ss) => {
        if (!signal.aborted) {
          spaces = ss;
        }
      })
      .catch((err) => {
        console.error(err);
      });
  });

  $effect(() => {
    const target = spaceId;
    let isStale = false;
    switchPanels = undefined;

    fetchSwitchPanels(session, target)
      .then((sps) => {
        if (!isStale) {
          switchPanels = sps;
        }
      })
      .catch((err) => {
        console.error(err);
      });

    return () => {
      isStale = true;
    };
  });

  $effect(() => {
    const target = switchPanelId;
    if (target === null) {
      switches = [];
      return;
    }

    let isStale = false;
    switches = undefined;

    fetchSwitches(session, target)
      .then((sws) => {
        if (!isStale) {
          switches = sws;
        }
      })
      .catch((err) => {
        console.error(err);
      });

    return () => {
      isStale = true;
    };
  });

  const onspacechange = () => {
    switchPanelId = null;
    switchId = null;
  };

  const onswitchpanelchange = () => {
    switchId = null;
  };
</script>

<div class={styles.selector}>
  <Select
    label={t("switchSelector.space")}
    options={spaceOptions}
    bind:value={spaceId}
    {disabled}
    onchange={onspacechange}
  />
  <Select
    label={t("switchSelector.switchPanel")}
    options={switchPanelOptions}
    bind:value={switchPanelId}
    {disabled}
    onchange={onswitchpanelchange}
  />
  <Select
    label={t("switchSelector.switch")}
    options={switchOptions}
    bind:value={switchId}
    disabled={disabled || switchPanelId === null}
  />
</div>
