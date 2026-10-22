<script lang="ts">
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import { notifyLoginCallback } from "@/lib/login-popup";

  onMount(() => {
    notifyLoginCallback(location.search);
    window.close();
  });
</script>

<main>
  <h2>{t("auth.signInReceived")}</h2>
  <p class="muted">{t("auth.closeWindowHint")}</p>
</main>

<style>
  main {
    display: flex;
    flex: 1;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 16px;
    text-align: center;
  }
</style>
