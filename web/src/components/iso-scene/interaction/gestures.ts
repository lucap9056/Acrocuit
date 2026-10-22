import { panBy, zoomAt } from "@/components/iso-scene/camera/navigate";
import type { Point } from "@/components/iso-scene/camera/projection";
import type { ViewportState } from "@/components/iso-scene/camera/viewport";

const WHEEL_ZOOM_SPEED = 0.0015;
const TAP_SLOP_PX = 6;

interface TapCandidate {
  pointerId: number;
  start: Point;
}

export interface GestureOptions {
  toLocal: (event: MouseEvent) => Point;
  getViewport: () => ViewportState;
  onviewportchange: (viewport: ViewportState) => void;
  onpanningchange: (isPanning: boolean) => void;
  ontap?: (event: PointerEvent) => void;
  onhover?: (event: PointerEvent) => void;
}

export interface Gestures {
  onwheel: (event: WheelEvent) => void;
  onpointerdown: (event: PointerEvent) => void;
  onpointermove: (event: PointerEvent) => void;
  onpointerup: (event: PointerEvent) => void;
  onpointercancel: (event: PointerEvent) => void;
}

const midpoint = (a: Point, b: Point): Point => ({
  x: (a.x + b.x) / 2,
  y: (a.y + b.y) / 2,
});

const distance = (a: Point, b: Point) => Math.hypot(a.x - b.x, a.y - b.y);

const pinch = (
  viewport: ViewportState,
  previous: Point,
  current: Point,
  other: Point,
): ViewportState => {
  const previousMid = midpoint(previous, other);
  const currentMid = midpoint(current, other);
  const previousDistance = distance(previous, other);
  const zoomed =
    previousDistance > 0
      ? zoomAt(
          viewport,
          distance(current, other) / previousDistance,
          currentMid,
        )
      : viewport;

  return panBy(
    zoomed,
    currentMid.x - previousMid.x,
    currentMid.y - previousMid.y,
  );
};

export const createGestures = ({
  toLocal,
  getViewport,
  onviewportchange,
  onpanningchange,
  ontap,
  onhover,
}: GestureOptions): Gestures => {
  const pointers = new Map<number, Point>();
  let tapCandidate: TapCandidate | undefined;

  const release = (pointerId: number) => {
    pointers.delete(pointerId);
    onpanningchange(pointers.size > 0);
    tapCandidate = undefined;
  };

  const capture = (event: PointerEvent) => {
    const target = event.currentTarget;
    if (
      target instanceof Element &&
      !target.hasPointerCapture(event.pointerId)
    ) {
      target.setPointerCapture(event.pointerId);
    }
  };

  const move = (event: PointerEvent, previous: Point) => {
    const current = toLocal(event);
    const other = [...pointers].find(([id]) => id !== event.pointerId)?.[1];
    pointers.set(event.pointerId, current);

    if (tapCandidate && distance(current, tapCandidate.start) > TAP_SLOP_PX) {
      tapCandidate = undefined;
    }
    if (!tapCandidate) {
      capture(event);
    }

    const viewport = getViewport();
    onviewportchange(
      other
        ? pinch(viewport, previous, current, other)
        : panBy(viewport, current.x - previous.x, current.y - previous.y),
    );
  };

  return {
    onwheel: (event) => {
      event.preventDefault();
      const point = toLocal(event);
      const factor = Math.exp(-event.deltaY * WHEEL_ZOOM_SPEED);
      onviewportchange(zoomAt(getViewport(), factor, point));
    },
    onpointerdown: (event) => {
      if (event.pointerType === "mouse" && event.button !== 0) {
        return;
      }
      const point = toLocal(event);
      tapCandidate =
        pointers.size === 0
          ? { pointerId: event.pointerId, start: point }
          : undefined;
      pointers.set(event.pointerId, point);
      onpanningchange(true);
    },
    onpointermove: (event) => {
      const previous = pointers.get(event.pointerId);
      if (previous) {
        move(event, previous);
      } else {
        onhover?.(event);
      }
    },
    onpointerup: (event) => {
      const isTap = tapCandidate?.pointerId === event.pointerId;
      release(event.pointerId);
      if (isTap) {
        ontap?.(event);
      } else if (pointers.size === 0) {
        onhover?.(event);
      }
    },
    onpointercancel: (event) => {
      release(event.pointerId);
    },
  };
};
