import type { Snippet } from "svelte";

export interface WorkspaceParts {
  leftpart?: Snippet;
  centerpart: Snippet;
  toolbarpart?: Snippet;
  rightpart?: Snippet;
}

export type WorkspaceHandler = (parts?: WorkspaceParts) => void;

export interface WorkspaceLayer {
  update: (parts: WorkspaceParts) => void;
  release: () => void;
}

export interface Workspace {
  push: (parts: WorkspaceParts, signal: AbortSignal) => WorkspaceLayer;
  on: (handler: WorkspaceHandler) => WorkspaceParts | undefined;
  off: (handler: WorkspaceHandler) => boolean;
}

interface LayerEntry {
  parts: WorkspaceParts;
}

export const createWorkspace = (): Workspace => {
  const handlers: Set<WorkspaceHandler> = new Set();
  const layers: LayerEntry[] = [];

  const visibleParts = (): WorkspaceParts | undefined => layers.at(-1)?.parts;

  const notify = () => {
    const parts = visibleParts();
    for (const handler of handlers) {
      handler(parts);
    }
  };

  return {
    push: (parts, signal) => {
      const entry: LayerEntry = { parts };
      const isTop = () => layers.at(-1) === entry;

      const release = () => {
        const index = layers.indexOf(entry);
        if (index === -1) {
          return;
        }
        const wasTop = isTop();
        layers.splice(index, 1);
        if (wasTop) {
          notify();
        }
      };

      if (signal.aborted) {
        return { update: () => {}, release: () => {} };
      }

      layers.push(entry);
      signal.addEventListener("abort", release, { once: true });
      notify();

      return {
        update: (next) => {
          if (!layers.includes(entry)) {
            return;
          }
          entry.parts = next;
          if (isTop()) {
            notify();
          }
        },
        release,
      };
    },
    on: (handler) => {
      handlers.add(handler);
      const parts = visibleParts();
      return parts ? { ...parts } : undefined;
    },
    off: (handler) => handlers.delete(handler),
  };
};
