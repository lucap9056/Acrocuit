<script lang="ts">
  import api, { type TokenPair } from "@/api";
  import load from "@/components/loading-bar";
  import {
    assertNoProviderError,
    exchangeCallbackForToken,
    openPopup,
    type Popup,
    type PopupCallback,
  } from "@/lib/login-popup";

  import styles from "./login.module.css";

  interface Props {
    onsuccess: (token: TokenPair) => void;
  }
  const { onsuccess }: Props = $props();

  let pending = $state<AbortController>();
  let error = $state<string>();

  const awaitPopupCallback = (popup: Popup): Promise<PopupCallback> => {
    const controller = new AbortController();
    pending = controller;

    return popup.waitCallback(controller.signal);
  };

  const signIn = () => {
    const loading = load.begin();

    error = undefined;

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
</script>

<div class={styles.login}>
  <h1>Acrocuit</h1>
  {#if pending}
    <button onclick={cancel}>Cancel</button>
  {:else}
    <button onclick={signIn}>SignIn</button>
  {/if}
</div>
