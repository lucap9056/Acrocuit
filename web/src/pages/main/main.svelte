<script lang="ts">
  import { type AuthError, AuthSession, type RefreshCallback } from "@/api";
  import cache from "@/lib/cache";
  import { onMount } from "@/lib/lifecycle";
  import load from "@/lib/load";
  import Dashboard from "./dashboard/dashboard.svelte";
  import Login from "./login.svelte";

  let session = $state<AuthSession | undefined>(undefined);
  let sessionErr = $state<AuthError | undefined>(undefined);
  let isLoaded = $state(false);
  const isUnauthenticated = $derived(sessionErr?.status === 401);

  const getSession = () => {
    return (
      cache.auth ||
      (() => {
        const callback: RefreshCallback = {
          onsuccess() {},
          onerror(err) {
            console.error("Refresh error:", err);
          },
        };
        const r = AuthSession.New(callback);
        cache.auth = r;
        return r;
      })()
    );
  };

  onMount((signal) => {
    const loading = load.begin();

    getSession()
      .then((s) => {
        if (!signal.aborted) {
          session = s;
        }
      })
      .catch((err: AuthError) => {
        sessionErr = err;
      })
      .finally(() => {
        if (!signal.aborted) {
          isLoaded = true;
        }
        loading.release();
      });
  });
</script>

{#if isLoaded}
  {#if session}
    <Dashboard {session} />
  {:else if isUnauthenticated}
    <Login />
  {:else}
    <div>{sessionErr?.message ?? "Unknown error"}</div>
  {/if}
{:else}
  <div>Loading...</div>
{/if}
