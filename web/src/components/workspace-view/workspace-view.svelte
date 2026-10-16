<script lang="ts">
  import { onMount } from "@/lib/lifecycle";
  import type { Workspace, WorkspaceParts } from "@/lib/workspace";

  import styles from "./workspace-view.module.css";

  interface Props {
    workspace: Workspace;
  }

  const { workspace }: Props = $props();

  let currentParts = $state<WorkspaceParts>();

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

{#if currentParts}
  <div class={styles.workspace}>
    {#if currentParts.leftpart}
      {@render currentParts.leftpart()}
    {/if}
    {#if currentParts.centerpart}
      {@render currentParts.centerpart()}
    {/if}
    {#if currentParts.rightpart}
      {@render currentParts.rightpart()}
    {/if}
  </div>
{/if}
