<script lang="ts">
  import type { Space, SpaceId, Switch } from "@/api";
  import { LINK_ATTRIBUTE_FIELDS } from "@/lib/attributes";
  import { t } from "@/lib/i18n";
  import {
    type ChainDetail,
    type ChainNode,
    chainLinkKey,
    chainNodeKey,
    detailSpaceId,
  } from "@/lib/power-chain";
  import {
    type ChainLoads,
    formatLinkLoad,
    formatSwitchLoad,
    formatSwitchPower,
    formatSwitchVoltage,
    formatVoltageDrop,
    isLinkOverloaded,
    isOverloaded,
    isVoltageDropExcessive,
    type SwitchLoad,
  } from "@/lib/power-chain/load";
  import styles from "./chain-node-info.module.css";

  interface Props {
    detail?: ChainDetail;
    spaces: Space[];
    loads?: ChainLoads;
  }

  const { detail, spaces, loads }: Props = $props();

  const node = $derived.by((): ChainNode | undefined => {
    switch (detail?.kind) {
      case "switch":
        return { kind: "switch", sw: detail.sw, panel: detail.panel };
      case "device":
        return { kind: "device", device: detail.device };
      case "link":
        return detail.link.downstream;
      default:
        return undefined;
    }
  });
  const linkLoad = $derived(
    detail?.kind === "link" ? loads?.links.get(detail.link.key) : undefined,
  );
  const load = $derived(
    detail?.kind === "switch" ? loads?.switches.get(detail.sw.id) : undefined,
  );
  const drop = $derived(
    node ? loads?.drops.get(chainNodeKey(node)) : undefined,
  );

  const spaceName = (id: SpaceId | undefined) =>
    spaces.find((space) => space.id === id)?.name ?? t("common.none");

  const nodeName = (target: ChainNode) =>
    target.kind === "switch" ? target.sw.name : target.device.name;

  const detailName = (target: ChainDetail) => {
    switch (target.kind) {
      case "panel":
        return target.panel.name;
      case "switch":
        return target.sw.name;
      case "device":
        return target.device.name;
      case "link":
        return `${target.link.upstream.sw.name} → ${nodeName(target.link.downstream)}`;
    }
  };

  const attributeText = (value: unknown, unit: string | null) =>
    [typeof value === "string" ? value : JSON.stringify(value), unit]
      .filter(Boolean)
      .join(" ");

  const supplyText = (supply: SwitchLoad) =>
    [
      formatSwitchVoltage(supply),
      supply.phase && t(`attributes.options.phase.${supply.phase}`),
    ]
      .filter(Boolean)
      .join(" · ") || t("common.none");
</script>

{#snippet wires(
  upstreams: Switch[],
  target: ChainNode,
)}
  {#each upstreams as upstream (upstream.id)}
    {@const link = loads?.links.get(chainLinkKey(upstream.id, target))}
    {@const text = link && formatLinkLoad(link)}
    {#if link && text}
      <div class={styles.row}>
        <dt>{t("powerChain.nodeInfo.wireFrom", { name: upstream.name })}</dt>
        <dd class={[isLinkOverloaded(link) && styles.overloaded]}>
          {text}
          {#if Number.isFinite(link.voltageDrop)}
            <span class={styles.note}>
              ΔV {Number(link.voltageDrop.toFixed(2))} V
            </span>
          {/if}
        </dd>
      </div>
    {/if}
  {/each}
{/snippet}

{#snippet switchBlocks(
  switches: Switch[],
)}
  {#if switches.length === 0}
    <dd>{t("common.none")}</dd>
  {:else}
    <dd class={styles.blocks}>
      <ul>
        {#each switches as sw (sw.id)}
          <li>{sw.name}</li>
        {/each}
      </ul>
    </dd>
  {/if}
{/snippet}

{#snippet voltageDrop()}
  {#if drop && drop.drop > 0}
    <div class={styles.row}>
      <dt>{t("powerChain.nodeInfo.voltageDrop")}</dt>
      <dd class={[isVoltageDropExcessive(drop) && styles.overloaded]}>
        {formatVoltageDrop(drop)}
      </dd>
    </div>
  {/if}
{/snippet}

<div class={styles.node_info}>
  {#if detail}
    <p class={styles.kind}>{t(`powerChain.nodeInfo.${detail.kind}`)}</p>
    <h2 class={styles.name}>{detailName(detail)}</h2>
    <dl class={styles.details}>
      {#if detail.kind === "link"}
        {@const { link } = detail}
        <div class={styles.row}>
          <dt>{t("powerChain.nodeInfo.linkFrom")}</dt>
          <dd>{link.upstream.sw.name}</dd>
        </div>
        <div class={styles.row}>
          <dt>{t("powerChain.nodeInfo.linkTo")}</dt>
          <dd>{nodeName(link.downstream)}</dd>
        </div>
        {#each LINK_ATTRIBUTE_FIELDS as field (field.key)}
          {@const value = link.attributes?.[field.key]}
          {#if value != null}
            <div class={styles.row}>
              <dt>{t(`attributes.fields.${field.key}`)}</dt>
              <dd>{attributeText(value, field.unit)}</dd>
            </div>
          {/if}
        {/each}
        {#if linkLoad}
          {@const text = formatLinkLoad(linkLoad)}
          {#if text}
            <div class={styles.row}>
              <dt>{t("powerChain.nodeInfo.linkCurrent")}</dt>
              <dd class={[isLinkOverloaded(linkLoad) && styles.overloaded]}>
                {text}
                {#if Number.isFinite(linkLoad.voltageDrop)}
                  <span class={styles.note}>
                    ΔV {Number(linkLoad.voltageDrop.toFixed(2))} V
                  </span>
                {/if}
              </dd>
            </div>
          {/if}
        {/if}
        {@render voltageDrop()}
      {:else if detail.kind === "panel"}
        <div class={styles.row}>
          <dt>{t("floorInfo.floor")}</dt>
          <dd>{spaceName(detailSpaceId(detail))}</dd>
        </div>
        <div class={styles.row}>
          <dt>{t("powerChain.nodeInfo.switches")}</dt>
          {@render switchBlocks(detail.switches)}
        </div>
      {:else if detail.kind === "switch"}
        <div class={styles.row}>
          <dt>{t("floorInfo.floor")}</dt>
          <dd>{spaceName(detailSpaceId(detail))}</dd>
        </div>
        <div class={styles.row}>
          <dt>{t("powerChain.nodeInfo.panel")}</dt>
          <dd>{detail.panel.name}</dd>
        </div>
        <div class={styles.row}>
          <dt>{t("powerChain.nodeInfo.upstream")}</dt>
          {@render switchBlocks(detail.upstreams)}
        </div>
        {#if load}
          <div class={styles.row}>
            <dt>{t("powerChain.nodeInfo.voltage")}</dt>
            <dd>
              {supplyText(load)}
              {#if load.isVoltageInherited && Number.isFinite(load.voltage)}
                <span class={styles.note}>
                  {t("powerChain.nodeInfo.inheritedVoltage")}
                </span>
              {/if}
            </dd>
          </div>
          <div class={styles.row}>
            <dt>{t("powerChain.nodeInfo.power")}</dt>
            <dd>{formatSwitchPower(load)}</dd>
          </div>
          <div class={styles.row}>
            <dt>{t("powerChain.nodeInfo.load")}</dt>
            <dd class={[isOverloaded(load) && styles.overloaded]}>
              {formatSwitchLoad(load)}
              {#if load.unknown > 0}
                <span class={styles.note}>
                  {t("powerChain.nodeInfo.unknownDevices", {
                    count: load.unknown,
                  })}
                </span>
              {/if}
            </dd>
          </div>
        {/if}
        {#if node}
          {@render wires(detail.upstreams, node)}
        {/if}
        {@render voltageDrop()}
      {:else}
        <div class={styles.row}>
          <dt>{t("floorInfo.floor")}</dt>
          <dd>{spaceName(detailSpaceId(detail))}</dd>
        </div>
        <div class={styles.row}>
          <dt>{t("powerChain.nodeInfo.feeders")}</dt>
          {@render switchBlocks(detail.feeders)}
        </div>
        {#if node}
          {@render wires(detail.feeders, node)}
        {/if}
        {@render voltageDrop()}
      {/if}
    </dl>
    {#if detail.kind === "link"}
      <p class={styles.action}>{t("powerChain.nodeInfo.linkHint")}</p>
    {/if}
  {:else}
    <p class={styles.message}>{t("powerChain.nodeInfo.hint")}</p>
  {/if}
</div>
