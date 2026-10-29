import type { SpaceGroup, SpaceGroupId } from "@/api/models";
import { empty, ok } from "@/lib/local-backend/response";
import { include, json, type Router } from "@/lib/local-backend/router";
import type { LocalBackend } from "@/lib/local-backend/storage/types";

type NameBody = { name: string };
type AddMemberBody = { email: string; role: string };
type MemberRoleBody = { role: string };
type TransferOwnerBody = { email: string };
type PublicBody = { public: boolean };

const memberEmailSelf = "me";
const base = "/api/space-groups";
const spaceGroup = `${base}/:spaceGroupId`;

export const spaceGroupsHandlers = (
  router: Router,
  { SpaceGroups, Spaces }: LocalBackend,
) => {
  router.get(base, ({ request }) =>
    SpaceGroups.GetSpaceGroups(include(request)).then(ok),
  );

  router.post(base, ({ request }) => {
    const { name } = json<NameBody>(request);
    return SpaceGroups.AddSpaceGroup(name).then((id) =>
      ok({ id, name, role: "owner" } satisfies SpaceGroup),
    );
  });

  router.get(spaceGroup, ({ request, params }) =>
    SpaceGroups.GetSpaceGroup(
      params.spaceGroupId as SpaceGroupId,
      include(request),
    ).then(ok),
  );

  router.put(spaceGroup, ({ request, params }) =>
    SpaceGroups.EditSpaceGroupName(
      params.spaceGroupId as SpaceGroupId,
      json<NameBody>(request).name,
    ).then(ok),
  );

  router.delete(spaceGroup, ({ params }) =>
    SpaceGroups.DelSpaceGroup(params.spaceGroupId as SpaceGroupId).then(empty),
  );

  router.get(`${spaceGroup}/spaces`, ({ request, params }) => {
    const spaceGroupId = params.spaceGroupId as SpaceGroupId;
    return SpaceGroups.GetSpaceGroup(spaceGroupId, "id")
      .then(() => Spaces.GetSpaces(spaceGroupId, include(request)))
      .then(ok);
  });

  router.post(`${spaceGroup}/spaces`, ({ request, params }) =>
    Spaces.AddSpace(
      params.spaceGroupId as SpaceGroupId,
      json<NameBody>(request).name,
    ).then(ok),
  );

  router.get(`${spaceGroup}/members`, ({ params }) =>
    SpaceGroups.GetSpaceGroupMembers(params.spaceGroupId as SpaceGroupId).then(
      ok,
    ),
  );

  router.post(`${spaceGroup}/members`, ({ request, params }) => {
    const { email, role } = json<AddMemberBody>(request);
    return SpaceGroups.AddSpaceGroupMember(
      params.spaceGroupId as SpaceGroupId,
      email,
      role,
    ).then(empty);
  });

  router.put(`${spaceGroup}/members/:memberEmail`, ({ request, params }) =>
    SpaceGroups.SetSpaceGroupMemberRole(
      params.spaceGroupId as SpaceGroupId,
      params.memberEmail,
      json<MemberRoleBody>(request).role,
    ).then(empty),
  );

  router.delete(`${spaceGroup}/members/:memberEmail`, ({ params }) => {
    const spaceGroupId = params.spaceGroupId as SpaceGroupId;
    const removal =
      params.memberEmail === memberEmailSelf
        ? SpaceGroups.LeaveSpaceGroup(spaceGroupId)
        : SpaceGroups.DelSpaceGroupMember(spaceGroupId, params.memberEmail);
    return removal.then(empty);
  });

  router.put(`${spaceGroup}/owner`, ({ request, params }) =>
    SpaceGroups.TransferSpaceGroupOwner(
      params.spaceGroupId as SpaceGroupId,
      json<TransferOwnerBody>(request).email,
    ).then(empty),
  );

  router.put(`${spaceGroup}/public`, ({ request, params }) =>
    SpaceGroups.SetSpaceGroupPublic(
      params.spaceGroupId as SpaceGroupId,
      json<PublicBody>(request).public,
    ).then(empty),
  );
};
