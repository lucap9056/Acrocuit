<script lang="ts">
  import PlusIcon from "phosphor-svelte/lib/PlusIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import type { AuthSession, Device, Position, Space } from "@/api";
  import dialog from "@/components/dialog";
  import load from "@/components/loading-bar";
  import { tooltip } from "@/components/tooltip";
  import { addDevice, removeDevice } from "@/lib/cache/devices";
  import { t } from "@/lib/i18n";
  import {
    deviceKey,
    type SpaceFocus,
  } from "@/pages/main/dashboard/space-group-view/space-view/space-focus.svelte";
  import styles from "./device-list.module.css";

  interface Props {
    session: AuthSession;
    space: Space;
    devices?: Device[];
    focus: SpaceFocus;
    onedit: (device: Device) => void;
  }

  let {
    session,
    space,
    devices = $bindable(),
    focus,
    onedit,
  }: Props = $props();

  const INITIAL_POSITION: Position = { x: 0, y: 0, z: 0 };

  let deviceName = $state("");

  const withSpace = (device: Device) => ({
    ...device,
    space: { id: space.id },
  });

  const showNameDialog = (text: string, name: string = ""): Promise<string> => {
    deviceName = name;

    const { promise, resolve, reject } = Promise.withResolvers<string>();

    dialog.push({
      text: "",
      content: setDeviceName,
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => {
            reject();
            return true;
          },
        },
        {
          text,
          type: "",
          onclick: () => {
            resolve(deviceName);
            return true;
          },
        },
      ],
    });

    return promise;
  };

  const confirmAction = (text: string, action: string): Promise<void> => {
    const { promise, resolve, reject } = Promise.withResolvers<void>();

    dialog.push({
      text,
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => {
            reject();
            return true;
          },
        },
        {
          text: action,
          type: "",
          onclick: () => {
            resolve();
            return true;
          },
        },
      ],
    });

    return promise;
  };

  const update = (message: string, request: Promise<Device[]>) => {
    const loading = load.begin(message);
    request
      .then((ds) => {
        devices = ds;
      })
      .catch((err) => {
        console.error(err);
      })
      .finally(() => {
        loading.release();
      });
  };

  const create = () => {
    showNameDialog(t("common.create")).then(
      (name) => {
        update(
          t("loading.createDevice"),
          addDevice(session, space.id, name, INITIAL_POSITION),
        );
      },
      () => {},
    );
  };

  const remove = (device: Device) => {
    confirmAction(
      t("devices.confirmDelete", { name: device.name }),
      t("common.delete"),
    ).then(
      () => {
        update(
          t("loading.removeDevice"),
          removeDevice(session, withSpace(device)),
        );
      },
      () => {},
    );
  };
</script>

{#snippet setDeviceName()}
  <input type="text" bind:value={deviceName}>
{/snippet}

<section class={styles.section}>
  <div class={styles.heading}>
    <h2 class={styles.title}>{t("devices.title")}</h2>
    <button
      class="icon_button"
      aria-label={t("devices.new")}
      onclick={create}
      {@attach tooltip(t("devices.new"))}
    >
      <PlusIcon />
    </button>
  </div>
  {#if devices === undefined}
    <p class={styles.message} role="status">{t("common.loading")}</p>
  {:else if devices.length === 0}
    <p class={styles.message}>{t("common.noneYet")}</p>
  {:else}
    <ul class={styles.list}>
      {#each devices as device (device.id)}
        {@const key = deviceKey(device.id)}
        <li
          class={[styles.item, focus.hoveredKey === key && styles.focused]}
          onpointerenter={() => {
            focus.enter(key);
            focus.scheduleReveal(key);
          }}
          onpointerleave={() => {
            focus.leave(key);
            focus.cancelReveal();
          }}
        >
          <button
            class={styles.name}
            onclick={() => onedit(device)}
            onfocus={() => focus.scheduleReveal(key)}
            onblur={() => focus.cancelReveal()}
          >
            {device.name}
          </button>
          <div class={styles.actions}>
            <button
              class="icon_button {styles.delete}"
              aria-label={t("common.delete")}
              onclick={() => remove(device)}
              {@attach tooltip(t("common.delete"))}
            >
              <TrashIcon />
            </button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</section>
