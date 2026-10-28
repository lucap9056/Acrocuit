<script lang="ts">
  import GlobeIcon from "phosphor-svelte/lib/GlobeIcon";
  import LockIcon from "phosphor-svelte/lib/LockIcon";
  import PlusIcon from "phosphor-svelte/lib/PlusIcon";
  import TrashIcon from "phosphor-svelte/lib/TrashIcon";
  import api, {
    type AuthSession,
    type SpaceGroup,
    type SpaceGroupMember,
    type SpaceGroupMemberRole,
    type SpaceGroupMembers,
    type SpaceGroupRole,
  } from "@/api";
  import dialog from "@/components/dialog";
  import CollapsibleSection from "@/components/collapsible-section/collapsible-section.svelte";
  import load from "@/components/loading-bar";
  import Select, { type SelectOption } from "@/components/select/select.svelte";
  import Spinner from "@/components/spinner/spinner.svelte";
  import Textbox from "@/components/textbox/textbox.svelte";
  import { tooltip } from "@/components/tooltip";
  import { t } from "@/lib/i18n";
  import { onMount } from "@/lib/lifecycle";
  import { canManageMembers } from "@/lib/permissions";
  import styles from "./space-group-editor.module.css";

  interface Props {
    session: AuthSession;
    spaceGroup: SpaceGroup;
    onrename: (name: string) => Promise<unknown>;
    ontransfer: (email: string) => Promise<unknown>;
  }

  const { session, spaceGroup, onrename, ontransfer }: Props = $props();

  const isMemberManager = $derived(canManageMembers(session, spaceGroup));

  const roleOptions: SelectOption<SpaceGroupMemberRole>[] = [
    { value: "viewer", label: t("spaceGroups.role.viewer") },
    { value: "editor", label: t("spaceGroups.role.editor") },
  ];

  const memberRoleOptions: SelectOption<SpaceGroupRole>[] = [
    ...roleOptions,
    { value: "owner", label: t("spaceGroups.role.owner") },
  ];

  let members = $state<SpaceGroupMember[]>();
  let isPublic = $state<boolean>();
  let isPending = $state(false);
  let email = $state("");
  let role = $state<SpaceGroupMemberRole>("viewer");
  let roleSelectVersion = $state(0);

  const canAdd = $derived(email.trim() !== "" && !isPending);

  const fetchMembers = () => api.spaceGroups.getMembers(session, spaceGroup.id);

  const showMembers = (result: SpaceGroupMembers) => {
    members = result.members;
    isPublic = result.public;
  };

  onMount((signal) => {
    if (!isMemberManager) {
      return;
    }
    fetchMembers()
      .then((result) => {
        if (!signal.aborted) {
          showMembers(result);
        }
      })
      .catch((err) => {
        console.error(err);
      });
  });

  const update = (message: string, request: () => Promise<unknown>) => {
    isPending = true;
    const loading = load.begin(message);
    return request()
      .then(fetchMembers)
      .then(showMembers)
      .finally(() => {
        isPending = false;
        loading.release();
      });
  };

  const add = () => {
    if (!canAdd) {
      return;
    }
    const target = email.trim();
    update(t("loading.addSpaceGroupMember"), () =>
      api.spaceGroups.addMember(session, spaceGroup.id, target, role),
    )
      .then(() => {
        email = "";
      })
      .catch((err) => {
        console.error(err);
      });
  };

  const onemailkeydown = (event: KeyboardEvent) => {
    if (event.key === "Enter" && !event.isComposing) {
      event.preventDefault();
      add();
    }
  };

  const transfer = (member: SpaceGroupMember) => {
    dialog.push({
      text: t("spaceGroupEditor.confirmTransfer", { email: member.email }),
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => {
            roleSelectVersion++;
            return true;
          },
        },
        {
          text: t("spaceGroupEditor.transfer"),
          type: "danger",
          onclick: () => {
            isPending = true;
            ontransfer(member.email)
              .catch((err) => {
                console.error(err);
                roleSelectVersion++;
              })
              .finally(() => {
                isPending = false;
              });
            return true;
          },
        },
      ],
    });
  };

  const onmemberrolechange = (
    member: SpaceGroupMember,
    nextRole: SpaceGroupRole,
  ) => {
    if (nextRole === "owner") {
      transfer(member);
    } else {
      changeRole(member, nextRole);
    }
  };

  const changeRole = (
    member: SpaceGroupMember,
    nextRole: SpaceGroupMemberRole,
  ) => {
    update(t("loading.updateSpaceGroupMember"), () =>
      api.spaceGroups.updateMemberRole(
        session,
        spaceGroup.id,
        member.email,
        nextRole,
      ),
    ).catch((err) => {
      console.error(err);
    });
  };

  const togglePublic = () => {
    if (isPublic === undefined) {
      return;
    }
    const nextPublic = !isPublic;
    dialog.push({
      text: nextPublic
        ? t("spaceGroupEditor.confirmPublic")
        : t("spaceGroupEditor.confirmPrivate"),
      buttons: [
        {
          text: t("common.cancel"),
          type: "",
          onclick: () => true,
        },
        {
          text: nextPublic
            ? t("spaceGroupEditor.publish")
            : t("spaceGroupEditor.unpublish"),
          type: "primary",
          onclick: () => {
            update(t("loading.updateSpaceGroupPublic"), () =>
              api.spaceGroups.setPublic(session, spaceGroup.id, nextPublic),
            ).catch((err) => {
              console.error(err);
            });
            return true;
          },
        },
      ],
    });
  };

  const remove = (member: SpaceGroupMember) => {
    update(t("loading.removeSpaceGroupMember"), () =>
      api.spaceGroups.deleteMember(session, spaceGroup.id, member.email),
    ).catch((err) => {
      console.error(err);
    });
  };
</script>

<div class={styles.editor}>
  <Textbox
    class={styles.title}
    label={t("spaceGroupEditor.name")}
    value={spaceGroup.name}
    disabled={spaceGroup.role === "viewer"}
    onsave={onrename}
  />
  {#if isMemberManager}
    <div class={styles.visibility}>
      <span class={styles.status} data-public={isPublic}>
        {#if isPublic}
          <GlobeIcon />
          <span>{t("spaceGroupEditor.public")}</span>
        {:else}
          <LockIcon />
          <span>{t("spaceGroupEditor.private")}</span>
        {/if}
      </span>
      <button
        class={styles.toggle}
        disabled={isPublic === undefined || isPending}
        onclick={togglePublic}
      >
        {isPublic
          ? t("spaceGroupEditor.makePrivate")
          : t("spaceGroupEditor.makePublic")}
      </button>
    </div>
    <CollapsibleSection title={t("spaceGroupEditor.members")} open>
      <ul class={styles.rows}>
        <li class={styles.row}>
          <input
            class={styles.email}
            type="email"
            aria-label={t("spaceGroupEditor.email")}
            placeholder={t("spaceGroupEditor.email")}
            disabled={isPending}
            bind:value={email}
            onkeydown={onemailkeydown}
          >
          <div class={styles.role}>
            <Select
              label={t("spaceGroupEditor.role")}
              options={roleOptions}
              disabled={isPending}
              bind:value={role}
            />
          </div>
          <button
            class="icon_button {styles.add}"
            aria-label={t("spaceGroupEditor.addMember")}
            disabled={!canAdd}
            onclick={add}
            {@attach tooltip(t("spaceGroupEditor.addMember"))}
          >
            <PlusIcon />
          </button>
        </li>
        {#if members === undefined}
          <li
            class={styles.loading}
            role="status"
            aria-label={t("common.loading")}
          >
            <Spinner />
          </li>
        {:else}
          {#each members as member (member.email)}
            <li class={styles.row}>
              <span class={styles.name}>{member.email}</span>
              {#if member.role === "owner"}
                <span class={styles.owner}>{t("spaceGroups.role.owner")}</span>
              {:else}
                <div class={styles.role}>
                  {#key roleSelectVersion}
                    <Select
                      label={t("spaceGroupEditor.role")}
                      options={memberRoleOptions}
                      value={member.role}
                      disabled={isPending}
                      onchange={(nextRole) => onmemberrolechange(member, nextRole)}
                    />
                  {/key}
                </div>
                <button
                  class="icon_button {styles.delete}"
                  aria-label={t("common.remove")}
                  disabled={isPending}
                  onclick={() => remove(member)}
                  {@attach tooltip(t("common.remove"))}
                >
                  <TrashIcon />
                </button>
              {/if}
            </li>
          {/each}
        {/if}
      </ul>
    </CollapsibleSection>
  {/if}
</div>
