<script lang="ts">
  import SidebarSimpleIcon from "phosphor-svelte/lib/SidebarSimpleIcon";
  import { tooltip } from "@/components/tooltip";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import type { Workspace, WorkspaceParts } from "@/lib/workspace";

  import styles from "./workspace-view.module.css";

  type Side = "left" | "right";

  const collapseLabelKeys = {
    left: "workspace.collapseLeftPanel",
    right: "workspace.collapseRightPanel",
  } as const;

  const expandLabelKeys = {
    left: "workspace.expandLeftPanel",
    right: "workspace.expandRightPanel",
  } as const;

  interface Props {
    workspace: Workspace;
  }

  const NARROW_WIDTH_PX = 960;

  const { workspace }: Props = $props();

  let currentParts = $state<WorkspaceParts>();
  let width = $state(0);
  let wideVisible = $state<Record<Side, boolean>>({ left: true, right: true });
  let overlaySide = $state<Side>();

  const isNarrow = $derived(width > 0 && width < NARROW_WIDTH_PX);

  const isVisible = (side: Side): boolean =>
    isNarrow ? overlaySide === side : wideVisible[side];

  const toggle = (side: Side) => {
    if (isNarrow) {
      overlaySide = overlaySide === side ? undefined : side;
    } else {
      wideVisible[side] = !wideVisible[side];
    }
  };

  const closeOverlay = () => {
    overlaySide = undefined;
  };

  const onkeydown = (event: KeyboardEvent) => {
    if (event.key === "Escape" && overlaySide !== undefined) {
      closeOverlay();
    }
  };

  $effect(() => {
    if (!isNarrow) {
      closeOverlay();
    }
  });

  onMount((signal) => {
    const onWorkspaceUpdated = (parts?: WorkspaceParts) => {
      if (signal.aborted) return;
      currentParts = parts;
    };

    const parts = workspace.on(onWorkspaceUpdated);
    onWorkspaceUpdated(parts);

    return () => {
      workspace.off(onWorkspaceUpdated);
    };
  });
</script>

<svelte:window {onkeydown} />

{#snippet toggleButton(
  side: Side,
)}
  {@const label = t(
    isVisible(side) ? collapseLabelKeys[side] : expandLabelKeys[side],
  )}
  <button
    class="icon_button"
    aria-expanded={isVisible(side)}
    aria-label={label}
    onclick={() => toggle(side)}
    {@attach tooltip(label)}
  >
    <SidebarSimpleIcon class={side === "right" ? styles.mirrored : undefined} />
  </button>
{/snippet}

{#if currentParts}
  <div
    class={[styles.workspace, isNarrow && styles.narrow]}
    bind:clientWidth={width}
  >
    {#if currentParts.leftpart}
      <aside
        class={[styles.side, styles.left]}
        data-visible={isVisible("left")}
        inert={!isVisible("left")}
      >
        <div class={styles.panel}>
          <div class={styles.header}>
            <div class={styles.header_end}>{@render toggleButton("left")}</div>
          </div>
          <div class={styles.content}>{@render currentParts.leftpart()}</div>
        </div>
      </aside>
    {/if}

    <div class={styles.main}>
      <div class={[styles.header, styles.toolbar]}>
        {#if currentParts.leftpart && !isVisible("left")}
          {@render toggleButton("left")}
        {/if}
        <div class={styles.toolbar_content}>
          {#if currentParts.toolbarpart}
            {@render currentParts.toolbarpart()}
          {/if}
        </div>
        {#if currentParts.rightpart && !isVisible("right")}
          {@render toggleButton("right")}
        {/if}
      </div>
      <main class={styles.center}>
        {@render currentParts.centerpart()}
      </main>
    </div>

    {#if currentParts.rightpart}
      <aside
        class={[styles.side, styles.right]}
        data-visible={isVisible("right")}
        inert={!isVisible("right")}
      >
        <div class={styles.panel}>
          <div class={styles.header}>{@render toggleButton("right")}</div>
          <div class={styles.content}>{@render currentParts.rightpart()}</div>
        </div>
      </aside>
    {/if}

    {#if isNarrow && overlaySide !== undefined}
      <button
        class={styles.backdrop}
        aria-label={t("workspace.closePanel")}
        onclick={closeOverlay}
      ></button>
    {/if}
  </div>
{/if}
