<script lang="ts">
  import CloudSlashIcon from "phosphor-svelte/lib/CloudSlashIcon";
  import SignInIcon from "phosphor-svelte/lib/SignInIcon";
  import SignOutIcon from "phosphor-svelte/lib/SignOutIcon";
  import api, {
    type AuthSession,
    type SpaceGroup,
    type SpaceGroupId,
    type TokenPair,
  } from "@/api";
  import BackButton from "@/components/back-button/back-button.svelte";
  import dialog, { type DialogContentProps } from "@/components/dialog";
  import HeaderTitle from "@/components/header-title/header-title.svelte";
  import load from "@/components/loading-bar";
  import { tooltip } from "@/components/tooltip";
  import WorkspaceView from "@/components/workspace-view/workspace-view.svelte";
  import { fetchSpaceGroups } from "@/lib/cache/space-groups";
  import type {
    HashRouter,
    RouteHandler,
    RouteParams,
  } from "@/lib/hash-routers";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import { writeLocalMode } from "@/lib/local-backend/mode";
  import { createWorkspace } from "@/lib/workspace";
  import Login from "../login.svelte";
  import styles from "./dashboard.module.css";
  import SpaceGroupSelector from "./space-group-selector.svelte";
  import SpaceGroupView from "./space-group-view/space-group-view.svelte";

  interface Props {
    session: AuthSession;
    router: HashRouter;
    onlogin: (token: TokenPair) => void;
  }
  const { session, router, onlogin }: Props = $props();

  const workspace = $state(createWorkspace());
  let spaceGroup = $state<SpaceGroup>();

  const loadSpaceGroup = (spaceGroupId: SpaceGroupId) => {
    if (session.kind === "guest") {
      return api.spaceGroups.getSpaceGroup(session, spaceGroupId);
    }
    return fetchSpaceGroups(session).then(
      (spaceGroups) =>
        spaceGroups.find((sg) => sg.id === spaceGroupId) ??
        api.spaceGroups.getSpaceGroup(session, spaceGroupId),
    );
  };

  const onSpaceGroupSelected = (
    signal: AbortSignal,
    spaceGroupIdStr: string,
  ) => {
    const loading = load.begin(t("loading.fetchSpaceGroups"));
    loadSpaceGroup(spaceGroupIdStr as SpaceGroupId)
      .then((sg) => {
        if (!signal.aborted) {
          spaceGroup = sg;
        }
      })
      .catch((err) => {
        console.error(err);
        if (session.kind === "guest" && !signal.aborted) {
          router.back("");
        }
      })
      .finally(() => {
        loading.release();
      });
  };

  onMount((signal) => {
    const paramName = "spaceGroupId";
    const routePath = `space-groups/:${paramName}/*`;

    const onSpaceGroupSelectedRouteHandle: RouteHandler = (
      params?: RouteParams,
    ) => {
      if (params !== undefined && typeof params[paramName] === "string") {
        const spaceGroupId = params[paramName];
        onSpaceGroupSelected(signal, spaceGroupId);
      } else {
        if (!signal.aborted) {
          spaceGroup = undefined;
        }
      }
    };
    const params = router.on(
      routePath,
      onSpaceGroupSelectedRouteHandle,
      signal,
    );
    onSpaceGroupSelectedRouteHandle(params);
  });

  const showLogin = () => {
    dialog.push({
      content: loginContent,
      buttons: [],
      maxWidth: "50rem",
    });
  };

  const openLocalMode = () => {
    dialog.push({
      text: t("auth.enterLocalModeConfirm"),
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => true,
        },
        {
          text: t("common.confirm"),
          type: "primary",
          onclick: () => {
            writeLocalMode(true);
            location.assign("/");
            return true;
          },
        },
      ],
    });
  };

  const logout = () => {
    if (session.kind === "local") {
      dialog.push({
        text: t("auth.exitLocalModeConfirm"),
        buttons: [
          {
            text: t("common.cancel"),
            type: "",
            onclick: () => true,
          },
          {
            text: t("common.confirm"),
            type: "primary",
            onclick: () => {
              writeLocalMode(false);
              location.assign("/");
              return true;
            },
          },
        ],
      });
      return;
    }

    const loading = load.begin(t("loading.signOut"));
    api.auth
      .logout()
      .then(() => {
        location.assign("/");
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
      });
  };
</script>

{#snippet loginContent({
  dismiss,
}: DialogContentProps)}
  <Login
    embedded
    onsuccess={(token) => {
      dismiss();
      onlogin(token);
    }}
    onclose={dismiss}
  />
{/snippet}

<div class={styles.dashboard}>
  <div class={styles.header}>
    <div class={styles.left}>
      <BackButton />
    </div>
    <div class={styles.center}>
      <HeaderTitle />
    </div>
    <div class={styles.right}>
      {#if session.kind !== "local"}
        <button
          class="icon_button"
          onclick={openLocalMode}
          {@attach tooltip(t("auth.localMode"))}
        >
          <CloudSlashIcon />
        </button>
      {:else}
        <span class={styles.local_mode_hint}>{t("auth.inLocalMode")}</span>
      {/if}
      {#if session.kind !== "guest"}
        <button
          class="icon_button"
          onclick={logout}
          {@attach tooltip(t("auth.signOut"))}
        >
          <SignOutIcon />
        </button>
      {:else}
        <button
          class="icon_button"
          onclick={showLogin}
          {@attach tooltip(t("auth.signIn"))}
        >
          <SignInIcon />
        </button>
      {/if}
    </div>
  </div>
  {#each spaceGroup
    ? [spaceGroup]
    : [] as currentSpaceGroup (currentSpaceGroup.id)}
    <WorkspaceView {workspace} />
    <SpaceGroupView
      {session}
      {workspace}
      {router}
      spaceGroup={currentSpaceGroup}
    />
  {:else}
    {#if session.kind !== "guest"}
      <div class={styles.space_groups}>
        <SpaceGroupSelector {session} {router} />
      </div>
    {/if}
  {/each}
</div>
