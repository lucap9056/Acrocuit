<script lang="ts">
  import {
    AuthError,
    type AuthSession,
    LocalAuthSession,
    MemberAuthSession,
    type RefreshCallback,
    type TokenPair,
  } from "@/api";
  import Dialog from "@/components/dialog/dialog.svelte";
  import load from "@/components/loading-bar";
  import LoadingBar from "@/components/loading-bar/loading-bar.svelte";
  import Toast from "@/components/toast/toast.svelte";
  import { createHashRouter } from "@/lib/hash-routers";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import Dashboard from "./dashboard/dashboard.svelte";
  import Login from "./login.svelte";
  import { clearSession, getSession, setSession } from "./main";
  import styles from "./main.module.css";

  const router = createHashRouter();

  let session = $state<AuthSession>();
  let sessionErr = $state<Error>();
  let isLoaded = $state(false);
  const isUnauthenticated = $derived(
    sessionErr instanceof AuthError && sessionErr.status === 401,
  );
  let isSpaceGroupRoute = $state(false);
  const canEnterDashboard = $derived(
    session?.kind === "member" ||
      session?.kind === "local" ||
      isSpaceGroupRoute,
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
    const loading = load.begin(t("loading.connect"));

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

    const spaceGroupParams = router.on(
      "space-groups/:spaceGroupId/*",
      (params) => {
        isSpaceGroupRoute = params !== undefined;
      },
      signal,
    );
    isSpaceGroupRoute = spaceGroupParams !== undefined;

    return () => {
      router.dispose();
    };
  });

  const onLogin = (token: TokenPair) => {
    if (mountSignal) {
      const s = new MemberAuthSession(callback, true, token);
      setSession(s);
      session = s;
      sessionErr = undefined;
    }
  };

  const onLocal = () => {
    if (mountSignal) {
      const s = new LocalAuthSession();
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
  {#if session && canEnterDashboard}
    {#key session}
      <Dashboard {session} {router} onlogin={onLogin} />
    {/key}
  {:else if session || isUnauthenticated}
    <Login onsuccess={onLogin} onlocal={onLocal} />
  {:else}
    {#if sessionErr}
      <div>{sessionErr.message}</div>
    {/if}
    <div class={styles.status}>
      <button onclick={reload}>{t("common.retry")}</button>
    </div>
  {/if}
{/if}

<Dialog />
<LoadingBar />
<Toast />
