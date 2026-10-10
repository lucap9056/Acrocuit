<script lang="ts">
  import type { ApiError, AuthSession, SpaceGroup } from "@/api";
  import api from "@/api";
  import cache from "@/lib/cache";
  import { onMount } from "@/lib/lifecycle";
  import load from "@/lib/load";

  interface Props {
    session: AuthSession;
    onselected: (spaceGroup: SpaceGroup) => void;
  }

  const { session, onselected }: Props = $props();
  let spaceGroups = $state<SpaceGroup[]>([]);
  let spaceGroupsErr = $state<ApiError | undefined>(undefined);

  const getSpaceGroups = () => {
    return (
      cache.spaceGroups ||
      (() => {
        const req = api.spaceGroups.getSpaceGroups(session);
        cache.spaceGroups = req;
        return req;
      })()
    );
  };

  onMount((signal) => {
    const loading = load.begin();

    getSpaceGroups()
      .then((sgs) => {
        if (!signal.aborted) {
          spaceGroups = sgs;
          spaceGroupsErr = undefined;
        }
      })
      .catch((err: ApiError) => {
        if (!signal.aborted) {
          spaceGroupsErr = err;
        }
      })
      .finally(() => {
        loading.release();
      });
  });
</script>

<div>
  {#each spaceGroups as sp}
    <button onclick={() => onselected(sp)}>{sp.name}</button>
  {/each}
</div>
