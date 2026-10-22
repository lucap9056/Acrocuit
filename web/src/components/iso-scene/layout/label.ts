import type { Point } from "@/components/iso-scene/camera/projection";
import { measureTextWidth } from "./text-metrics";
import { LABEL_FONT } from "./typography";

const LEADER_LENGTH_PX = 28;
const TEXT_GAP_PX = 6;
const MAX_TEXT_LENGTH = 24;
const LABEL_HALF_HEIGHT_PX = 9;

export interface LabelLayout {
  anchor: Point;
  leaderEndX: number;
  textX: number;
  text: string;
  left: number;
  top: number;
  bottom: number;
}

const truncate = (text: string) =>
  text.length > MAX_TEXT_LENGTH
    ? `${text.slice(0, MAX_TEXT_LENGTH - 1)}…`
    : text;

export const layoutLabel = (rawText: string, anchor: Point): LabelLayout => {
  const leaderEndX = anchor.x - LEADER_LENGTH_PX;
  const textX = leaderEndX - TEXT_GAP_PX;
  const text = truncate(rawText).toUpperCase();

  return {
    anchor,
    leaderEndX,
    textX,
    text,
    left: textX - measureTextWidth(text, LABEL_FONT),
    top: anchor.y - LABEL_HALF_HEIGHT_PX,
    bottom: anchor.y + LABEL_HALF_HEIGHT_PX,
  };
};
