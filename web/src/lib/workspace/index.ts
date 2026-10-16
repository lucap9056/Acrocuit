import type { Snippet } from "svelte";

export interface WorkspaceParts {
  leftpart?: Snippet;
  centerpart: Snippet;
  rightpart?: Snippet;
}

export type WorkspaceHandler = (parts?: WorkspaceParts) => void;

export interface Workspace {
  set: (parts?: WorkspaceParts) => void;
  on: (handler: WorkspaceHandler) => WorkspaceParts | undefined;
  off: (handler: WorkspaceHandler) => boolean;
}

export const createWorkspace = (): Workspace => {
  const handlers: Set<WorkspaceHandler> = new Set();
  let current: WorkspaceParts | undefined;

  return {
    set: (parts?: WorkspaceParts) => {
      current = parts;
      for (const handler of handlers) {
        handler(parts);
      }
    },
    on: (handler: WorkspaceHandler): WorkspaceParts | undefined => {
      handlers.add(handler);
      if (current) {
        return { ...current };
      }
      return undefined;
    },
    off: (handler: WorkspaceHandler) => handlers.delete(handler),
  };
};
