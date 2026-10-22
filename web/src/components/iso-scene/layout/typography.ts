const FONT_FAMILY =
  '"IBM Plex Mono", ui-monospace, "Cascadia Mono", Consolas, "Noto Sans TC", monospace';

export const LABEL_FONT = `600 12px ${FONT_FAMILY}`;
export const TITLE_FONT = `600 13px ${FONT_FAMILY}`;
export const CAPTION_FONT = `11px ${FONT_FAMILY}`;

export const drawingFontsReady = Promise.all(
  [LABEL_FONT, CAPTION_FONT].map((font) => document.fonts.load(font)),
).catch(() => []);
