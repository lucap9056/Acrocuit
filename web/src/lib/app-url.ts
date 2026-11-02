const chunkDirectoryParent = "../";

export const appBaseUrl = import.meta.env.DEV
  ? new URL("/", location.href)
  : new URL(chunkDirectoryParent, import.meta.url);

export const appUrl = (path: string) => new URL(path, appBaseUrl);
