import { SvelteMap } from "svelte/reactivity";
import type { SwitchPanel, Device, Position } from "@/api";
import type { FocusKey } from "./space-focus.svelte";

export type PositionDraft =
  | { kind: "switch_panel"; target: SwitchPanel; position: Position }
  | { kind: "device"; target: Device; position: Position };

export class PositionDrafts {
  readonly entries = new SvelteMap<FocusKey, PositionDraft>();
  isSaving = $state(false);

  get size() {
    return this.entries.size;
  }

  get(key: FocusKey) {
    return this.entries.get(key);
  }

  set(key: FocusKey, draft: PositionDraft) {
    this.entries.set(key, draft);
  }

  clear() {
    this.entries.clear();
  }
}
