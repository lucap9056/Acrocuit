<script lang="ts">
  import {
    AuthError,
    type AuthSession,
    MemberAuthSession,
    type RefreshCallback,
    type TokenPair,
  } from "@/api";
  import dialog from "@/components/dialog";
  import Dialog from "@/components/dialog/dialog.svelte";
  import load from "@/components/loading-bar";
  import LoadingBar from "@/components/loading-bar/loading-bar.svelte";
  import Toast from "@/components/toast/toast.svelte";
  import { createHashRouter } from "@/lib/hash-routers";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import { writeLocalMode } from "@/lib/local-backend/mode";
  import Dashboard from "./dashboard/dashboard.svelte";
  import Login from "./login.svelte";
  import { clearSession, getSession, setSession } from "./main";

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
        if (signal.aborted) {
          return;
        }
        sessionErr = err;
        if (err instanceof AuthError && err.status === 401) {
          return;
        }
        dialog.push({
          text: err.message,
          buttons: [
            {
              text: t("auth.localMode"),
              type: "",
              align: "start",
              onclick: () => {
                onLocal();
                return true;
              },
            },
            {
              text: t("common.retry"),
              type: "primary",
              onclick: () => {
                reload();
                return true;
              },
            },
          ],
        });
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
    writeLocalMode(true);
    location.reload();
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
  {/if}
{/if}

<Dialog />
<LoadingBar />
<Toast />
