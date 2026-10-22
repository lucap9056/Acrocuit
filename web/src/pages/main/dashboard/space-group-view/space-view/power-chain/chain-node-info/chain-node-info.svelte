<script lang="ts">
  import type { Position, Space, SpaceId, Switch } from "@/api";
  import { t } from "@/lib/i18n";
  import { type ChainDetail, detailSpaceId } from "@/lib/power-chain";
  import styles from "./chain-node-info.module.css";

  interface Props {
    detail?: ChainDetail;
    spaces: Space[];
  }

  const { detail, spaces }: Props = $props();

  const spaceName = (id: SpaceId | undefined) =>
    spaces.find((space) => space.id === id)?.name ?? t("common.none");

  const placement = (position: Position | null) =>
    position ? t("powerChain.nodeInfo.placed") : t("powerChain.unplaced");

  const names = (switches: Switch[]) =>
    switches.length === 0
      ? t("common.none")
      : switches.map(({ name }) => name).join("、");
</script>

<div class={styles.node_info}>
  {#if detail}
    <p class={styles.kind}>{t(`powerChain.nodeInfo.${detail.kind}`)}</p>
    <h2 class={styles.name}>
      {detail.kind === "panel"
        ? detail.panel.name
        : detail.kind === "switch"
          ? detail.sw.name
          : detail.device.name}
    </h2>
    <dl class={styles.details}>
      <dt>{t("floorInfo.floor")}</dt>
      <dd>{spaceName(detailSpaceId(detail))}</dd>
      {#if detail.kind === "panel"}
        <dt>{t("powerChain.nodeInfo.switches")}</dt>
        <dd>{names(detail.switches)}</dd>
        <dt>{t("powerChain.nodeInfo.placement")}</dt>
        <dd>{placement(detail.panel.position)}</dd>
      {:else if detail.kind === "switch"}
        <dt>{t("switchPanels.title")}</dt>
        <dd>{detail.panel.name}</dd>
        <dt>{t("powerChain.nodeInfo.upstream")}</dt>
        <dd>{detail.upstream?.name ?? t("common.none")}</dd>
        <dt>{t("powerChain.nodeInfo.placement")}</dt>
        <dd>{placement(detail.panel.position)}</dd>
      {:else}
        <dt>{t("powerChain.nodeInfo.feeders")}</dt>
        <dd>{names(detail.feeders)}</dd>
        <dt>{t("powerChain.nodeInfo.placement")}</dt>
        <dd>{placement(detail.device.position)}</dd>
      {/if}
    </dl>
  {:else}
    <p class={styles.message}>{t("powerChain.nodeInfo.hint")}</p>
  {/if}
</div>
