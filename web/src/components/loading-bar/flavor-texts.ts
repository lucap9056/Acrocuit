import { t, translationKeys } from "@/lib/i18n";

export const pickFlavorText = () => {
  const keys = translationKeys("flavor_");
  if (keys.length === 0) {
    return undefined;
  }
  return t(keys[Math.floor(Math.random() * keys.length)]);
};
