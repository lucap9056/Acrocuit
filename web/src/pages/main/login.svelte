<script lang="ts">
  import XIcon from "phosphor-svelte/lib/XIcon";
  import api, { type TokenPair } from "@/api";
  import load from "@/components/loading-bar";
  import { tooltip } from "@/components/tooltip";
  import { t } from "@/lib/i18n";
  import {
    assertNoProviderError,
    exchangeCallbackForToken,
    openPopup,
    type Popup,
    type PopupCallback,
  } from "@/lib/login-popup";

  import styles from "./login.module.css";

  interface Props {
    embedded?: boolean;
    onsuccess: (token: TokenPair) => void;
    onclose?: () => void;
    onlocal?: () => void;
  }
  const { embedded = false, onsuccess, onclose, onlocal }: Props = $props();

  let pending = $state<AbortController>();

  const awaitPopupCallback = (popup: Popup): Promise<PopupCallback> => {
    const controller = new AbortController();
    pending = controller;

    return popup.waitCallback(controller.signal);
  };

  const signIn = () => {
    const loading = load.begin(t("loading.signIn"));

    api.auth
      .login()
      .then(openPopup)
      .then((p) => {
        loading.release();
        return p;
      })
      .then(awaitPopupCallback)
      .then(assertNoProviderError)
      .then(exchangeCallbackForToken)
      .then(onsuccess)
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
        pending = undefined;
      });
  };

  const cancel = () => {
    if (pending) {
      pending.abort();
    }
  };

  const close = () => {
    cancel();
    onclose?.();
  };
</script>

<div class={styles.login} data-embedded={embedded}>
  {#if onclose}
    <button
      class="icon_button {styles.close}"
      aria-label={t("common.close")}
      onclick={close}
      {@attach tooltip(t("common.close"))}
    >
      <XIcon />
    </button>
  {/if}
  <h1>Acrocuit</h1>
  {#if pending}
    <button onclick={cancel}>{t("common.cancel")}</button>
  {:else}
    <button onclick={signIn}>{t("auth.signIn")}</button>
    {#if onlocal}
      <button class={styles.local} onclick={onlocal}>
        {t("auth.localMode")}
      </button>
    {/if}
  {/if}
</div>
