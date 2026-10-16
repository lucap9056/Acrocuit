<script lang="ts">
  import {
    AuthError,
    AuthSession,
    type RefreshCallback,
    type TokenPair,
  } from "@/api";
  import Dialog from "@/components/dialog/dialog.svelte";
  import load from "@/components/loading-bar";
  import LoadingBar from "@/components/loading-bar/loading-bar.svelte";
  import { onMount } from "@/lib/lifecycle";
  import Dashboard from "./dashboard/dashboard.svelte";
  import Login from "./login.svelte";
  import { clearSession, getSession, setSession } from "./main";
  import styles from "./main.module.css";

  let session = $state<AuthSession>();
  let sessionErr = $state<Error>();
  let isLoaded = $state(false);
  const isUnauthenticated = $derived(
    sessionErr instanceof AuthError && sessionErr.status === 401,
  );

  const callback: RefreshCallback = {
    onsuccess() {
      if (sessionErr) {
        sessionErr = undefined;
      }
    },
    onerror: (err: AuthError) => {
      if (err.status !== 401) {
        return;
      }
      clearSession();
      session = undefined;
      sessionErr = err;
    },
  };

  const connect = (signal: AbortSignal) => {
    const loading = load.begin();

    getSession(callback)
      .then((s) => {
        if (!signal.aborted) {
          session = s;
          sessionErr = undefined;
        }
      })
      .catch((err: Error) => {
        clearSession();
        if (!signal.aborted) {
          sessionErr = err;
        }
      })
      .finally(() => {
        if (!signal.aborted) {
          isLoaded = true;
        }
        loading.release();
      });
  };

  let mountSignal: AbortSignal | undefined;

  onMount((signal) => {
    mountSignal = signal;
    connect(signal);
  });

  const onLogin = (token: TokenPair) => {
    if (mountSignal) {
      const s = new AuthSession(callback, true, token);
      setSession(s);
      session = s;
      sessionErr = undefined;
    }
  };

  const reload = () => {
    if (mountSignal) {
      connect(mountSignal);
    } else {
      location.reload();
    }
  };
</script>

{#if isLoaded}
  {#if session}
    <Dashboard {session} />
  {:else if isUnauthenticated}
    <Login onsuccess={onLogin} />
  {:else}
    {#if sessionErr}
      <div>{sessionErr.message}</div>
    {/if}
    <div class={styles.status}><button onclick={reload}>Retry</button></div>
  {/if}
{/if}

<Dialog />
<LoadingBar />
