import i18next from "i18next";
import HttpBackend from "i18next-http-backend";
import type en from "../../../public/locales/en.json";

declare module "i18next" {
  interface CustomTypeOptions {
    resources: { translation: typeof en };
  }
}

export const baseLocale = "en";

export const supportedLocales = [baseLocale, "zh-TW"] as const;

export const ready = i18next.use(HttpBackend).init({
  lng: navigator.language,
  supportedLngs: supportedLocales,
  fallbackLng: { zh: ["zh-TW", baseLocale], default: [baseLocale] },
  load: "currentOnly",
  backend: {
    loadPath: `${import.meta.env.BASE_URL}locales/{{lng}}.json`,
  },
  interpolation: { escapeValue: false },
});

export const translationKeys = <P extends string>(prefix: P) => {
  const keys = new Set(
    i18next.languages.flatMap((locale) =>
      Object.keys(i18next.getResourceBundle(locale, "translation") ?? {}),
    ),
  );
  return [...keys].filter((key) => key.startsWith(prefix)) as Extract<
    keyof typeof en,
    `${P}${string}`
  >[];
};

export { t } from "i18next";
