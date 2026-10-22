import type { SwitchPanelId, DeviceId } from "@/api";

export type FocusKey = `switch-panel-${number}` | `device-${number}`;

export const switchPanelKey = (id: SwitchPanelId): FocusKey =>
  `switch-panel-${id}`;

export const deviceKey = (id: DeviceId): FocusKey => `device-${id}`;

export interface RevealRequest {
  key: FocusKey;
}

const REVEAL_DELAY_MS = 150;

export class SpaceFocus {
  hoveredKey = $state<FocusKey>();
  revealRequest = $state.raw<RevealRequest>();
  #revealTimer: ReturnType<typeof setTimeout> | undefined;

  scheduleReveal(key: FocusKey) {
    this.cancelReveal();
    this.#revealTimer = setTimeout(() => {
      this.#revealTimer = undefined;
      this.revealRequest = { key };
    }, REVEAL_DELAY_MS);
  }

  cancelReveal() {
    clearTimeout(this.#revealTimer);
    this.#revealTimer = undefined;
  }

  enter(key: FocusKey) {
    this.hoveredKey = key;
  }

  leave(key: FocusKey) {
    if (this.hoveredKey === key) {
      this.hoveredKey = undefined;
    }
  }
}
