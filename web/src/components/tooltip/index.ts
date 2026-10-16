import type { Attachment } from "svelte/attachments";
import styles from "./tooltip.module.css";

const ANCHOR_GAP_PX = 6;
const VIEWPORT_MARGIN_PX = 4;

let bubble: HTMLDivElement | undefined;
let currentAnchor: HTMLElement | undefined;

const getBubble = () => {
  if (!bubble) {
    bubble = document.createElement("div");
    bubble.className = styles.tooltip;
    bubble.popover = "manual";
    bubble.setAttribute("role", "tooltip");
    document.body.append(bubble);
  }
  return bubble;
};

const clamp = (value: number, min: number, max: number) =>
  Math.min(Math.max(value, min), max);

const placeNear = (target: HTMLDivElement, anchor: HTMLElement) => {
  const rect = anchor.getBoundingClientRect();
  const { offsetWidth: width, offsetHeight: height } = target;

  const left = clamp(
    rect.left + rect.width / 2 - width / 2,
    VIEWPORT_MARGIN_PX,
    window.innerWidth - width - VIEWPORT_MARGIN_PX,
  );
  const below = rect.bottom + ANCHOR_GAP_PX;
  const fitsBelow = below + height <= window.innerHeight - VIEWPORT_MARGIN_PX;
  const top = fitsBelow ? below : rect.top - ANCHOR_GAP_PX - height;

  target.style.left = `${left}px`;
  target.style.top = `${top}px`;
};

export const tooltip =
  (text: string): Attachment<HTMLElement> =>
  (anchor) => {
    anchor.setAttribute("aria-label", text);

    const show = () => {
      const target = getBubble();
      currentAnchor = anchor;
      target.textContent = text;
      if (!target.matches(":popover-open")) {
        target.showPopover();
      }
      placeNear(target, anchor);
    };

    const hide = () => {
      if (currentAnchor !== anchor) {
        return;
      }
      currentAnchor = undefined;
      bubble?.hidePopover();
    };

    anchor.addEventListener("pointerenter", show);
    anchor.addEventListener("pointerleave", hide);
    anchor.addEventListener("focus", show);
    anchor.addEventListener("blur", hide);
    anchor.addEventListener("click", hide);

    return () => {
      anchor.removeEventListener("pointerenter", show);
      anchor.removeEventListener("pointerleave", hide);
      anchor.removeEventListener("focus", show);
      anchor.removeEventListener("blur", hide);
      anchor.removeEventListener("click", hide);
      hide();
    };
  };
