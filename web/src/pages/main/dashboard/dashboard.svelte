<script lang="ts">
  import type { AuthSession, SpaceGroup, SpaceGroupId } from "@/api";
  import SpaceGroupSelector from "./selectors/space-group-selector.svelte";

  interface Props {
    session: AuthSession;
  }
  const { session }: Props = $props();

  let spaceGroup = $state<SpaceGroup>();

  const onSpaceGroupSelected = (sg: SpaceGroup) => {
    spaceGroup = sg;
  };
</script>

<div>
  {#if spaceGroup}
  {:else}
    <SpaceGroupSelector {session} onselected={onSpaceGroupSelected} />
  {/if}
</div>
