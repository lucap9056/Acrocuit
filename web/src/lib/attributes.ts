import type en from "@public/locales/en.json";
import type { Attributes } from "@/api";
import { t } from "@/lib/i18n";
import { isIconRef } from "@/lib/icon-packs/index.svelte";
import { solvePower } from "@/lib/power";

export const ATTRIBUTES_MAX_BYTES = 2048;

export type AttributeKey = keyof (typeof en)["attributes"]["fields"];

export type AttributeOption = { value: string; label: () => string };

type BaseField = {
  key: AttributeKey;
  unit: string | null;
  estimate?: (fixed: AttributesForm["fixed"]) => string | undefined;
};

export type AttributeField =
  | (BaseField & { type: "text" | "integer" | "icon" })
  | (BaseField & { type: "number"; min?: number; max?: number })
  | (BaseField & { type: "enum"; options: readonly AttributeOption[] });

export const LINK_ATTRIBUTE_FIELDS: readonly AttributeField[] = [
  { key: "wire_size", type: "number", unit: "mm²", min: 0 },
  { key: "wire_length", type: "number", unit: "m", min: 0 },
  { key: "wire_type", type: "text", unit: null },
  { key: "wire_cores", type: "integer", unit: null },
  { key: "rated_current", type: "number", unit: "A", min: 0 },
];

const estimateRatedCurrent = (fixed: AttributesForm["fixed"]) => {
  const { current } = solvePower(
    fixed.rated_power ?? "",
    fixed.rated_voltage ?? "",
    fixed.rated_current ?? "",
    fixed.power_factor ?? "",
  );
  return Number.isFinite(current)
    ? String(Number(current.toFixed(2)))
    : undefined;
};

export const DEVICE_ATTRIBUTE_FIELDS: readonly AttributeField[] = [
  { key: "icon", type: "icon", unit: null },
  { key: "rated_power", type: "number", unit: "W", min: 0 },
  { key: "rated_voltage", type: "number", unit: "V", min: 0 },
  {
    key: "rated_current",
    type: "number",
    unit: "A",
    min: 0,
    estimate: estimateRatedCurrent,
  },
  { key: "power_factor", type: "number", unit: null, min: 0, max: 1 },
];

export type Phase = "single" | "three";

const PHASE_OPTIONS: readonly AttributeOption[] = [
  { value: "single", label: () => t("attributes.options.phase.single") },
  { value: "three", label: () => t("attributes.options.phase.three") },
];

const SWITCH_TYPE_OPTIONS: readonly AttributeOption[] = [
  { value: "nfb", label: () => t("attributes.options.switch_type.nfb") },
  { value: "elcb", label: () => t("attributes.options.switch_type.elcb") },
  { value: "switch", label: () => t("attributes.options.switch_type.switch") },
];

export const SWITCH_ATTRIBUTE_FIELDS: readonly AttributeField[] = [
  { key: "circuit_label", type: "text", unit: null },
  {
    key: "switch_type",
    type: "enum",
    unit: null,
    options: SWITCH_TYPE_OPTIONS,
  },
  { key: "max_current", type: "number", unit: "A", min: 0 },
  { key: "rated_voltage", type: "number", unit: "V", min: 0 },
  { key: "phase", type: "enum", unit: null, options: PHASE_OPTIONS },
  { key: "poles", type: "integer", unit: "P" },
  { key: "leakage_current", type: "number", unit: "mA", min: 0 },
];

export const SWITCH_PANEL_ATTRIBUTE_FIELDS: readonly AttributeField[] = [
  { key: "icon", type: "icon", unit: null },
  { key: "switch_capacity", type: "integer", unit: null },
];

export const switchCapacity = (attributes: Attributes | null) => {
  const value = attributes?.switch_capacity;
  return typeof value === "number" && Number.isInteger(value) && value > 0
    ? value
    : undefined;
};

export const attributePhase = (
  attributes: Attributes | null,
): Phase | undefined => {
  const value = attributes?.phase;
  return value === "single" || value === "three" ? value : undefined;
};

export const attributeIcon = (attributes: Attributes | null) => {
  const value = attributes?.icon;
  return typeof value === "string" && isIconRef(value) ? value : undefined;
};

export type CustomAttribute = { key: string; value: string };

export type AttributesForm = {
  fixed: Partial<Record<AttributeKey, string>>;
  custom: CustomAttribute[];
};

export type AttributesIssue =
  | { kind: "invalidNumber"; key: AttributeKey }
  | { kind: "invalidInteger"; key: AttributeKey }
  | { kind: "outOfRange"; key: AttributeKey; min?: number; max?: number }
  | { kind: "invalidOption"; key: AttributeKey }
  | { kind: "keyRequired" }
  | { kind: "duplicateKey"; key: string }
  | { kind: "tooLarge" };

export type AttributesResult =
  | { ok: true; attributes: Attributes }
  | { ok: false; issue: AttributesIssue };

const toText = (value: unknown) =>
  typeof value === "string" ? value : JSON.stringify(value);

export const toAttributesForm = (
  fields: readonly AttributeField[],
  attributes: Attributes | null,
): AttributesForm => {
  const source = attributes ?? {};
  const fixedKeys = new Set<string>(fields.map(({ key }) => key));
  const fixed = Object.fromEntries(
    fields.map(({ key }) => [
      key,
      source[key] == null ? "" : toText(source[key]),
    ]),
  );
  const custom = Object.entries(source)
    .filter(([key]) => !fixedKeys.has(key))
    .map(([key, value]) => ({ key, value: toText(value) }));
  return { fixed, custom };
};

const parseFixedField = (
  field: AttributeField,
  text: string,
): { value: unknown } | { issue: AttributesIssue } => {
  switch (field.type) {
    case "text":
      return { value: text };
    case "icon":
      return isIconRef(text)
        ? { value: text }
        : { issue: { kind: "invalidOption", key: field.key } };
    case "enum":
      return field.options.some(({ value }) => value === text)
        ? { value: text }
        : { issue: { kind: "invalidOption", key: field.key } };
    case "integer": {
      const value = Number(text);
      if (!Number.isFinite(value)) {
        return { issue: { kind: "invalidNumber", key: field.key } };
      }
      return Number.isInteger(value)
        ? { value }
        : { issue: { kind: "invalidInteger", key: field.key } };
    }
    case "number": {
      const value = Number(text);
      if (!Number.isFinite(value)) {
        return { issue: { kind: "invalidNumber", key: field.key } };
      }
      const { min, max } = field;
      if (
        (min !== undefined && value < min) ||
        (max !== undefined && value > max)
      ) {
        return { issue: { kind: "outOfRange", key: field.key, min, max } };
      }
      return { value };
    }
  }
};

export const fromAttributesForm = (
  fields: readonly AttributeField[],
  form: AttributesForm,
): AttributesResult => {
  const attributes: Attributes = {};
  const fixedKeys = new Set<string>(fields.map(({ key }) => key));

  for (const field of fields) {
    const text = (form.fixed[field.key] ?? "").trim();
    if (text === "") {
      continue;
    }
    const parsed = parseFixedField(field, text);
    if ("issue" in parsed) {
      return { ok: false, issue: parsed.issue };
    }
    attributes[field.key] = parsed.value;
  }

  for (const row of form.custom) {
    const key = row.key.trim();
    if (key === "" && row.value.trim() === "") {
      continue;
    }
    if (key === "") {
      return { ok: false, issue: { kind: "keyRequired" } };
    }
    if (Object.hasOwn(attributes, key) || fixedKeys.has(key)) {
      return { ok: false, issue: { kind: "duplicateKey", key } };
    }
    attributes[key] = row.value;
  }

  const size = new TextEncoder().encode(JSON.stringify(attributes)).length;
  if (size > ATTRIBUTES_MAX_BYTES) {
    return { ok: false, issue: { kind: "tooLarge" } };
  }

  return { ok: true, attributes };
};

const describeRange = (min?: number, max?: number) => {
  if (min !== undefined && max !== undefined) {
    return t("attributes.range.between", { min, max });
  }
  return min !== undefined
    ? t("attributes.range.atLeast", { min })
    : t("attributes.range.atMost", { max });
};

export const describeAttributesIssue = (issue: AttributesIssue) => {
  switch (issue.kind) {
    case "invalidNumber":
      return t("attributes.invalidNumber", {
        field: t(`attributes.fields.${issue.key}`),
      });
    case "invalidInteger":
      return t("attributes.invalidInteger", {
        field: t(`attributes.fields.${issue.key}`),
      });
    case "outOfRange":
      return t("attributes.outOfRange", {
        field: t(`attributes.fields.${issue.key}`),
        range: describeRange(issue.min, issue.max),
      });
    case "invalidOption":
      return t("attributes.invalidOption", {
        field: t(`attributes.fields.${issue.key}`),
      });
    case "keyRequired":
      return t("attributes.keyRequired");
    case "duplicateKey":
      return t("attributes.duplicateKey", { key: issue.key });
    case "tooLarge":
      return t("attributes.tooLarge", { max: ATTRIBUTES_MAX_BYTES });
  }
};
