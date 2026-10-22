import type { FloorPlanTone } from "@/lib/iso-scene/floor-plan-image";

export interface HighlightStyle {
  line: string;
  glow?: string;
}

export interface ScenePalette {
  paper: string;
  gridMinor: string;
  gridMajor: string;
  plane: string;
  line: string;
  lineMuted: string;
  lineFaint: string;
  highlight: HighlightStyle;
  trace: HighlightStyle;
  groups: string[];
}

export const BLUEPRINT_PALETTE: ScenePalette = {
  paper: "#1b4a85",
  gridMinor: "rgb(255 255 255 / 6%)",
  gridMajor: "rgb(255 255 255 / 14%)",
  plane: "rgb(22 64 118 / 78%)",
  line: "#e9f2ff",
  lineMuted: "rgb(233 242 255 / 55%)",
  lineFaint: "rgb(233 242 255 / 28%)",
  highlight: {
    line: "#7fe3ff",
    glow: "rgb(127 227 255 / 70%)",
  },
  trace: {
    line: "#ffd166",
    glow: "rgb(255 209 102 / 60%)",
  },
  groups: ["#4cc47c", "#b08cf2", "#f27474"],
};

export const GRAY_PALETTE: ScenePalette = {
  paper: "#c9cdd2",
  gridMinor: "rgb(42 48 56 / 6%)",
  gridMajor: "rgb(42 48 56 / 14%)",
  plane: "rgb(255 255 255 / 70%)",
  line: "#2a3038",
  lineMuted: "rgb(42 48 56 / 60%)",
  lineFaint: "rgb(42 48 56 / 30%)",
  highlight: {
    line: "#d97706",
    glow: "rgb(217 119 6 / 45%)",
  },
  trace: {
    line: "#2563eb",
    glow: "rgb(37 99 235 / 40%)",
  },
  groups: ["#1baf7a", "#a03fb8", "#008300", "#e87ba4"],
};

export const paletteForTone = (tone: FloorPlanTone): ScenePalette =>
  tone === "original" ? GRAY_PALETTE : BLUEPRINT_PALETTE;

export const groupColor = (
  { groups }: ScenePalette,
  group: number | undefined,
): string | undefined => (group === undefined ? undefined : groups[group]);
