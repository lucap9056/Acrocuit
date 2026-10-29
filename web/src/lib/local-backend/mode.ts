const LOCAL_MODE_STORAGE_KEY = "acrocuit:local-mode";

export const readLocalMode = (): boolean => {
  try {
    return localStorage.getItem(LOCAL_MODE_STORAGE_KEY) === "true";
  } catch {
    return false;
  }
};

export const writeLocalMode = (enabled: boolean) => {
  try {
    if (enabled) {
      localStorage.setItem(LOCAL_MODE_STORAGE_KEY, "true");
    } else {
      localStorage.removeItem(LOCAL_MODE_STORAGE_KEY);
    }
  } catch (err) {
    console.error(err);
  }
};
