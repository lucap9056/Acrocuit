import type { AuthSession, SpaceGroup } from "@/api";

export const canEdit = (sg: SpaceGroup) => sg.role !== "viewer";

export const canManage = (sg: SpaceGroup) => sg.role === "owner";

export const canManageMembers = (session: AuthSession, sg: SpaceGroup) =>
  session.kind === "member" && session.authRequired && canManage(sg);
