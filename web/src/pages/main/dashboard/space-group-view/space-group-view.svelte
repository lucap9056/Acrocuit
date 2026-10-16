<script lang="ts">
  import type { AuthSession, Space, SpaceGroup } from "@/api";
  import load from "@/components/loading-bar";
  import { fetchSpaces } from "@/lib/cache/spaces";
  import router, {
    field,
    type RouteHandler,
    type RouteParams,
  } from "@/lib/hash-routers";
  import { onMount } from "@/lib/lifecycle";

  import styles from "./space-group-view.module.css";
    import type { Workspace } from "@/lib/workspace";

  interface Props {
    session: AuthSession;
    workspace:Workspace;
    spaceGroup: SpaceGroup;
  }

  const { session, workspace, spaceGroup }: Props = $props();

  let spaces = $state<Space[]>([]);

  let currentIndex = $state(0);
  let isFocus = $state(false);

  const onSpaceSelected = (signal: AbortSignal, spaceIdStr: string) => {
    const spaceId = parseInt(spaceIdStr, 10);
    if (Number.isNaN(spaceId)) {
      router.back(`/space-groups/${spaceGroup.id.toString()}`);
      return;
    }

    const loading = load.begin();
    fetchSpaces(session, spaceGroup.id)
      .then((s) => {
        if (!signal.aborted) {
          currentIndex = s.findIndex(({ id }) => id === spaceId);
        }
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
      });
  };

  onMount((signal) => {
    currentIndex = 0;
    workspace.set({
        leftpart: spaceList,
        centerpart: spacesView,
        rightpart:spaceInfo
    });

    const loading = load.begin();

    fetchSpaces(session, spaceGroup.id)
      .then((s) => {
        if (!signal.aborted) {
          spaces = s;
        }
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
      });

    const paramName = "spaceId";
    const routePath = `space-groups/*/space/:${paramName}`;

    const onSpaceSelectedRouteHandle: RouteHandler = (params?: RouteParams) => {
      if (params !== undefined) {
        const spaceId = field(params, paramName, "");

        onSpaceSelected(signal, spaceId);
      } else {
        if (!signal.aborted) {
          currentIndex = 0;
        }
      }
    };

    const params = router.on(routePath, onSpaceSelectedRouteHandle);
    onSpaceSelectedRouteHandle(params);

    return () => {
      router.off(routePath, onSpaceSelectedRouteHandle);
    };
  });
</script>

{#snippet spaceList()}
<div></div>
{/snippet}

{#snippet spacesView()}
    <div class={styles.items}>
      {#each spaces as space}
        <div>{space.name}</div>
      {/each}
    </div>
{/snippet}

{#snippet spaceInfo()}
<div></div>
  
{/snippet}

{#if spaces.length > 0 && isFocus}
<div></div>
{/if}
