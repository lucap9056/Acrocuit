<script lang="ts">
  import ArrowLeftIcon from "phosphor-svelte/lib/ArrowLeftIcon";
  import { tooltip } from "@/components/tooltip";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import back, { type BackHandler, type BackListener } from "./index";

  let onback = $state<BackHandler>();

  onMount(() => {
    const listener: BackListener = (next) => {
      onback = next;
    };
    back.subscribe(listener);

    onback = back.current;

    return () => {
      back.unsubscribe(listener);
    };
  });
</script>

{#if onback}
  <button
    class="icon_button"
    aria-label={t("common.back")}
    onclick={onback}
    {@attach tooltip(t("common.back"))}
  >
    <ArrowLeftIcon />
  </button>
{/if}
