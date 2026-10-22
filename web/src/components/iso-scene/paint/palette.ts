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
};
