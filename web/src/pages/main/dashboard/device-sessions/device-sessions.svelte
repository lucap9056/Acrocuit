<script lang="ts">
  import api, { type DeviceSession, type MemberAuthSession } from "@/api";
  import Spinner from "@/components/spinner/spinner.svelte";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import styles from "./device-sessions.module.css";

  interface Props {
    session: MemberAuthSession;
  }

  const { session }: Props = $props();

  let deviceSessions = $state<DeviceSession[]>();
  let loadError = $state<string>();

  const formatUpdatedAt = (unixSeconds: number) =>
    new Date(unixSeconds * 1000).toLocaleString();

  onMount((signal) => {
    api.auth
      .listSessions(session)
      .then((result) => {
        if (!signal.aborted) {
          deviceSessions = result;
        }
      })
      .catch((err: Error) => {
        console.error(err);
        if (!signal.aborted) {
          loadError = err.message;
        }
      });
  });
</script>

<div class={styles.device_sessions}>
  <div class={styles.title}>{t("auth.deviceSessions")}</div>
  {#if loadError !== undefined}
    <div class={styles.error}>{loadError}</div>
  {:else if deviceSessions === undefined}
    <div class={styles.loading} role="status" aria-label={t("common.loading")}>
      <Spinner />
    </div>
  {:else}
    <ul class={styles.rows}>
      {#each deviceSessions as deviceSession (deviceSession.device_id)}
        <li class={styles.row}>
          <div class={styles.info}>
            <span class={styles.name}>{deviceSession.device_name}</span>
            <span class={styles.updated_at}>
              {formatUpdatedAt(deviceSession.updated_at)}
            </span>
          </div>
          {#if deviceSession.current}
            <span class={styles.current}>{t("auth.currentDevice")}</span>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>
