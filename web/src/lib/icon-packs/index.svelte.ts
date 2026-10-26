import { currentLocale } from "@/lib/i18n";

export type IconRef = string;

type LocalizedText = Record<string, string>;

interface IconPackManifest {
  schema: 1;
  id: string;
  name: string;
  version: string;
  license: string;
  viewBox: [number, number, number, number];
  style?: { strokeWidth?: number; lineCap?: string; lineJoin?: string };
  assets: { bundle: string };
  categories?: Record<string, LocalizedText>;
  icons: {
    name: string;
    category?: string;
    labels?: LocalizedText;
    tags?: string[];
  }[];
}

type IconPackBundle = Record<string, { strokes?: string[]; fills?: string[] }>;

export interface PackIcon {
  id: IconRef;
  name: string;
  category?: string;
  labels: LocalizedText;
  tags: string[];
  strokes: string[];
  fills: string[];
}

export interface IconPackStyle {
  viewBox: string;
  strokeWidth: number;
  lineCap: "butt" | "round" | "square";
  lineJoin: "miter" | "round" | "bevel";
}

export interface ResolvedIcon {
  pack: IconPack;
  icon: PackIcon;
}

export interface IconPack {
  id: string;
  name: string;
  license: string;
  source: string;
  builtIn: boolean;
  style: IconPackStyle;
  categories: Record<string, LocalizedText>;
  icons: PackIcon[];
}

const BUILT_IN_MANIFESTS = ["icon-packs/default/manifest.json"];

const SOURCES_STORAGE_KEY = "acrocuit:icon-pack-sources";

const PACK_ID_PATTERN = /^[a-z0-9-]+$/;

const MAX_MANIFEST_BYTES = 512 * 1024;

const MAX_BUNDLE_BYTES = 4 * 1024 * 1024;

const MAX_ICONS = 2000;

const ICON_NAME_PATTERN = /^[A-Za-z0-9_-]+$/;

const BUILT_IN_REF_PATTERN = /^([a-z0-9-]+):([A-Za-z0-9_-]+)$/;

const PATH_DATA_PATTERN = /^[\d\s.,eE+\-MmLlHhVvCcSsQqTtAaZz]*$/;

const MAX_PATH_LENGTH = 8192;

const LINE_CAPS = ["butt", "round", "square"] as const;

const LINE_JOINS = ["miter", "round", "bevel"] as const;

const isAllowedSource = (url: URL) =>
  url.protocol === "https:" || url.origin === location.origin;

export type ParsedIconRef =
  | { kind: "builtIn"; packId: string; name: string }
  | { kind: "source"; source: string; name: string };

export const parseIconRef = (value: string): ParsedIconRef | undefined => {
  const builtIn = BUILT_IN_REF_PATTERN.exec(value);
  if (builtIn) {
    return { kind: "builtIn", packId: builtIn[1], name: builtIn[2] };
  }
  let url: URL;
  try {
    url = new URL(value);
  } catch {
    return undefined;
  }
  const name = url.hash.slice(1);
  if (!isAllowedSource(url) || !ICON_NAME_PATTERN.test(name)) {
    return undefined;
  }
  url.hash = "";
  return { kind: "source", source: url.href, name };
};

export const isIconRef = (value: string) => parseIconRef(value) !== undefined;

const pick = <T extends string>(
  value: string | undefined,
  options: readonly T[],
  fallback: T,
): T => options.find((option) => option === value) ?? fallback;

const validPaths = (paths: unknown): string[] =>
  Array.isArray(paths)
    ? paths.filter(
        (d): d is string =>
          typeof d === "string" &&
          d.length <= MAX_PATH_LENGTH &&
          PATH_DATA_PATTERN.test(d),
      )
    : [];

export type IconPackErrorKind =
  | "invalidUrl"
  | "unreachable"
  | "unsupported"
  | "tooLarge";

export class IconPackError extends Error {
  readonly kind: IconPackErrorKind;

  constructor(kind: IconPackErrorKind, message: string) {
    super(message);
    this.name = "IconPackError";
    this.kind = kind;
  }
}

const fetchJson = async <T>(url: URL, maxBytes: number): Promise<T> => {
  const response = await fetch(url).catch((err: unknown) => {
    throw new IconPackError("unreachable", `${url}: ${(err as Error).message}`);
  });
  if (!response.ok) {
    throw new IconPackError("unreachable", `${url}: HTTP ${response.status}`);
  }
  const text = await response.text();
  if (text.length > maxBytes) {
    throw new IconPackError("tooLarge", `${url}: exceeds ${maxBytes} bytes`);
  }
  try {
    return JSON.parse(text) as T;
  } catch {
    throw new IconPackError("unsupported", `${url}: invalid JSON`);
  }
};

const loadPack = async (
  manifestUrl: URL,
  builtIn: boolean,
): Promise<IconPack> => {
  const manifest = await fetchJson<IconPackManifest>(
    manifestUrl,
    MAX_MANIFEST_BYTES,
  );
  if (
    manifest?.schema !== 1 ||
    typeof manifest.id !== "string" ||
    !PACK_ID_PATTERN.test(manifest.id) ||
    typeof manifest.assets?.bundle !== "string" ||
    !Array.isArray(manifest.icons) ||
    !Array.isArray(manifest.viewBox)
  ) {
    throw new IconPackError("unsupported", `${manifestUrl}: unsupported`);
  }
  if (manifest.icons.length > MAX_ICONS) {
    throw new IconPackError("tooLarge", `${manifestUrl}: too many icons`);
  }
  const bundle = await fetchJson<IconPackBundle>(
    new URL(manifest.assets.bundle, manifestUrl),
    MAX_BUNDLE_BYTES,
  );

  const icons = manifest.icons.flatMap((meta): PackIcon[] => {
    const name = meta?.name;
    const data =
      typeof name === "string" && Object.hasOwn(bundle, name)
        ? bundle[name]
        : undefined;
    if (!ICON_NAME_PATTERN.test(name) || !data) {
      return [];
    }
    const id = builtIn
      ? `${manifest.id}:${name}`
      : `${manifestUrl.href}#${name}`;
    const strokes = validPaths(data.strokes);
    const fills = validPaths(data.fills);
    if (strokes.length === 0 && fills.length === 0) {
      return [];
    }
    return [
      {
        id,
        name: meta.name,
        category: typeof meta.category === "string" ? meta.category : undefined,
        labels: meta.labels ?? {},
        tags: Array.isArray(meta.tags)
          ? meta.tags.filter((tag) => typeof tag === "string")
          : [],
        strokes,
        fills,
      },
    ];
  });

  return {
    id: manifest.id,
    name: typeof manifest.name === "string" ? manifest.name : manifest.id,
    license: typeof manifest.license === "string" ? manifest.license : "",
    source: manifestUrl.href,
    builtIn,
    style: {
      viewBox: manifest.viewBox.map(Number).join(" "),
      strokeWidth: Number(manifest.style?.strokeWidth) || 1.5,
      lineCap: pick(manifest.style?.lineCap, LINE_CAPS, "round"),
      lineJoin: pick(manifest.style?.lineJoin, LINE_JOINS, "round"),
    },
    categories: manifest.categories ?? {},
    icons,
  };
};

export const localize = (text: LocalizedText, fallback: string) =>
  text[currentLocale()] ?? text.en ?? fallback;

export const iconLabel = (icon: PackIcon) => localize(icon.labels, icon.name);

const appBaseUrl = () => new URL(import.meta.env.BASE_URL, location.href);

const readStoredSources = (): string[] => {
  try {
    const stored = JSON.parse(
      localStorage.getItem(SOURCES_STORAGE_KEY) ?? "[]",
    );
    return Array.isArray(stored)
      ? stored.filter((source) => typeof source === "string")
      : [];
  } catch {
    return [];
  }
};

const rememberSource = (source: string) => {
  const stored = readStoredSources();
  if (stored.includes(source)) {
    return;
  }
  try {
    localStorage.setItem(
      SOURCES_STORAGE_KEY,
      JSON.stringify([...stored, source]),
    );
  } catch (err) {
    console.error(err);
  }
};

export const toManifestUrl = (input: string): URL => {
  let url: URL;
  try {
    url = new URL(input.trim());
  } catch {
    throw new IconPackError("invalidUrl", input);
  }
  if (!isAllowedSource(url)) {
    throw new IconPackError("invalidUrl", input);
  }
  url.hash = "";
  if (url.pathname.endsWith(".json")) {
    return url;
  }
  const directory = url.pathname.endsWith("/") ? url : new URL(`${url.href}/`);
  return new URL("manifest.json", directory);
};

class IconRegistry {
  packs = $state.raw<IconPack[]>([]);

  #sources = new Map<string, Promise<IconPack>>();

  #failed = new Set<string>();

  #icons = $derived(
    new Map(
      this.packs.flatMap((pack) =>
        pack.icons.map((icon) => [icon.id, { pack, icon }] as const),
      ),
    ),
  );

  #loadSource(source: string, builtIn: boolean): Promise<IconPack> {
    const existing = this.#sources.get(source);
    if (existing) {
      return existing;
    }
    const loading = loadPack(new URL(source), builtIn).then(
      (pack) => {
        if (!this.packs.some((loaded) => loaded.source === pack.source)) {
          this.packs = [...this.packs, pack].sort(
            (a, b) => Number(b.builtIn) - Number(a.builtIn),
          );
        }
        if (!builtIn) {
          rememberSource(pack.source);
        }
        return pack;
      },
      (err: unknown) => {
        this.#failed.add(source);
        throw err;
      },
    );
    this.#sources.set(source, loading);
    return loading;
  }

  load() {
    return Promise.allSettled([
      ...BUILT_IN_MANIFESTS.map((path) =>
        this.#loadSource(new URL(path, appBaseUrl()).href, true),
      ),
      ...readStoredSources().map((source) => this.#loadSource(source, false)),
    ]).then((results) => {
      for (const result of results) {
        if (result.status === "rejected") {
          console.error(result.reason);
        }
      }
    });
  }

  addSource(input: string): Promise<IconPack> {
    const source = toManifestUrl(input).href;
    if (this.#failed.delete(source)) {
      this.#sources.delete(source);
    }
    return this.#loadSource(source, false);
  }

  get(ref: string | undefined): ResolvedIcon | undefined {
    if (ref === undefined) {
      return undefined;
    }
    const entry = this.#icons.get(ref);
    if (entry) {
      return entry;
    }
    const parsed = parseIconRef(ref);
    if (parsed?.kind === "source" && !this.#sources.has(parsed.source)) {
      this.#loadSource(parsed.source, false).catch((err) => {
        console.error(err);
      });
    }
    return undefined;
  }
}

export const iconRegistry = new IconRegistry();
