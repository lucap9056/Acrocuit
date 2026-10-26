<script lang="ts">
  import PlusIcon from "phosphor-svelte/lib/PlusIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import { untrack } from "svelte";
  import type {
    Attributes,
    AuthSession,
    LinkedSwitch,
    SpaceGroup,
    SpaceId,
    Switch,
    SwitchId,
    SwitchPanel,
    SwitchPanelId,
  } from "@/api";
  import { ApiError } from "@/api";
  import AttributesSection from "@/components/attributes-section/attributes-section.svelte";
  import CollapsibleSection from "@/components/collapsible-section/collapsible-section.svelte";
  import { editLinkAttributes } from "@/components/link-attributes-dialog/link-attributes-dialog.svelte";
  import load from "@/components/loading-bar";
  import Spinner from "@/components/spinner/spinner.svelte";
  import SwitchSelector from "@/components/switch-selector/switch-selector.svelte";
  import Textbox from "@/components/textbox/textbox.svelte";
  import { tooltip } from "@/components/tooltip";
  import WireIcon from "@/components/wire-icon/wire-icon.svelte";
  import { SWITCH_ATTRIBUTE_FIELDS } from "@/lib/attributes";
  import {
    addSwitchUpstream,
    fetchSwitchUpstreams,
    removeSwitchUpstream,
    setSwitchUpstreamAttributes,
  } from "@/lib/cache/switch-upstreams";
  import { renameSwitch, setSwitchAttributes } from "@/lib/cache/switches";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import styles from "./switch-editor.module.css";

  interface Props {
    session: AuthSession;
    spaceGroup: SpaceGroup;
    spaceId: SpaceId;
    switchPanel: SwitchPanel;
    sw: Switch;
  }

  const {
    session,
    spaceGroup,
    spaceId: initialSpaceId,
    switchPanel,
    sw,
  }: Props = $props();

  let current = $state(untrack(() => sw));
  let upstreams = $state<LinkedSwitch[]>();
  let isPending = $state(false);
  let addError = $state<{ switchId: SwitchId; message: string }>();
  let spaceId = $state<SpaceId>(untrack(() => initialSpaceId));
  let switchPanelId = $state<SwitchPanelId | null>(null);
  let switchId = $state<SwitchId | null>(null);

  const excludedIds = $derived([
    current.id,
    ...(upstreams?.map(({ id }) => id) ?? []),
  ]);

  onMount((signal) => {
    fetchSwitchUpstreams(session, current.id)
      .then((sws) => {
        if (!signal.aborted) {
          upstreams = sws;
        }
      })
      .catch((err) => {
        console.error(err);
      });
  });

  const rename = (name: string) => {
    const loading = load.begin(t("loading.renameSwitch"));
    return renameSwitch(
      session,
      { ...current, switch_panel: { id: switchPanel.id } },
      name,
    )
      .then((switches) => {
        current = switches.find((s) => s.id === current.id) ?? current;
      })
      .finally(() => {
        loading.release();
      });
  };

  const saveAttributes = (attributes: Attributes) => {
    const loading = load.begin(t("loading.updateSwitchAttributes"));
    return setSwitchAttributes(
      session,
      { ...current, switch_panel: { id: switchPanel.id } },
      attributes,
    )
      .then((switches) => {
        current = switches.find((s) => s.id === current.id) ?? current;
      })
      .finally(() => {
        loading.release();
      });
  };

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

  const describeAddError = (err: unknown) =>
    err instanceof ApiError && err.status === 400
      ? t("switchEditor.cycleRejected")
      : (err as Error).message;

  const add = () => {
    if (switchId === null) {
      return;
    }

    const target = switchId;
    addError = undefined;
    isPending = true;
    const loading = load.begin(t("loading.addSwitchUpstream"));
    addSwitchUpstream(session, current.id, target)
      .then((sws) => {
        upstreams = sws;
        switchId = null;
      })
      .catch((err) => {
        console.error(err);
        addError = { switchId: target, message: describeAddError(err) };
      })
      .finally(() => {
        isPending = false;
        loading.release();
      });
  };

  const editUpstreamAttributes = (upstream: LinkedSwitch) => {
    editLinkAttributes(upstream.link.attributes, (attributes) => {
      isPending = true;
      const loading = load.begin(t("loading.updateSwitchUpstream"));
      return setSwitchUpstreamAttributes(
        session,
        current.id,
        upstream.id,
        attributes,
      )
        .then((sws) => {
          upstreams = sws;
        })
        .finally(() => {
          isPending = false;
          loading.release();
        });
    });
  };

  const remove = (upstream: Switch) => {
    update(t("loading.removeSwitchUpstream"), () =>
      removeSwitchUpstream(session, current.id, upstream.id),
    );
  };
</script>

<div class={styles.editor}>
  <Textbox
    class={styles.title}
    label={t("switchEditor.name")}
    value={current.name}
    onsave={rename}
  />
  <CollapsibleSection title={t("switchEditor.upstream")} open>
    <ul class={styles.rows}>
      <li class={styles.row}>
        <div class={styles.selector}>
          <SwitchSelector
            {session}
            {spaceGroup}
            bind:spaceId
            bind:switchPanelId
            bind:switchId
            excludedSwitchIds={excludedIds}
          />
        </div>
        <button
          class="icon_button {styles.add}"
          aria-label={t("switchEditor.addUpstream")}
          disabled={switchId === null || isPending}
          onclick={add}
          {@attach tooltip(t("switchEditor.addUpstream"))}
        >
          <PlusIcon />
        </button>
      </li>
      {#if addError && addError.switchId === switchId}
        <li class={styles.error} role="alert">{addError.message}</li>
      {/if}
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
        {#each upstreams as upstream (upstream.id)}
          <li class={styles.row}>
            <span class={styles.name}>{upstream.name}</span>
            <button
              class="icon_button {styles.edit}"
              aria-label={t("attributes.edit")}
              disabled={isPending}
              onclick={() => editUpstreamAttributes(upstream)}
              {@attach tooltip(t("attributes.edit"))}
            >
              <WireIcon />
            </button>
            <button
              class="icon_button {styles.delete}"
              aria-label={t("common.remove")}
              disabled={isPending}
              onclick={() => remove(upstream)}
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
    fields={SWITCH_ATTRIBUTE_FIELDS}
    attributes={current.attributes}
    onsave={saveAttributes}
  />
</div>
