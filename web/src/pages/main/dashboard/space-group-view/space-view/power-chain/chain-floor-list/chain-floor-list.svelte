<script lang="ts">
  import type { Snippet } from "svelte";
  import type { Space, SpaceId } from "@/api";
  import { t } from "@/lib/i18n";
  import styles from "./chain-floor-list.module.css";

  interface Props {
    floors?: Space[];
    focusSpaceId?: SpaceId;
    focusIcon: Snippet;
    placeCounts: Map<SpaceId, number>;
    floorColors?: Map<SpaceId, string>;
    inspectedSpaceId?: SpaceId;
    onfloorhover: (id: SpaceId) => void;
  }

  const {
    floors,
    focusSpaceId,
    focusIcon,
    placeCounts,
    floorColors,
    inspectedSpaceId,
    onfloorhover,
  }: Props = $props();

  const topFirst = $derived([...(floors ?? [])].reverse());
</script>

<div class={styles.floor_list}>
  <h2 class={styles.title}>{t("powerChain.floors")}</h2>
  {#if floors === undefined}
    <p class={styles.message} role="status">{t("spaces.loading")}</p>
  {:else}
    <ul class={styles.list}>
      {#each topFirst as floor (floor.id)}
        {@const count = placeCounts.get(floor.id) ?? 0}
        {@const color = floorColors?.get(floor.id)}
        {@const isFocus = floor.id === focusSpaceId}
        <li>
          <button
            class={[
              styles.item,
              floor.id === inspectedSpaceId && styles.inspected,
              count === 0 && !isFocus && styles.passing,
            ]}
            onpointerenter={() => onfloorhover(floor.id)}
            onfocus={() => onfloorhover(floor.id)}
            onclick={() => onfloorhover(floor.id)}
          >
            {#if color}
              <span
                class={styles.swatch}
                style:background={color}
                aria-hidden="true"
              ></span>
            {/if}
            <span class={styles.name}>{floor.name}</span>
            {#if isFocus}
              <span class={styles.focus} aria-hidden="true">
                {@render focusIcon()}
              </span>
            {/if}
            {#if count > 0}
              <span class={styles.count}>{count}</span>
            {/if}
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</div>
