const ONE_YEAR_SECONDS = 60 * 60 * 24 * 365;

export const readCookie = (name: string): string | undefined => {
  const prefix = `${encodeURIComponent(name)}=`;
  const entry = document.cookie
    .split("; ")
    .find((item) => item.startsWith(prefix));
  return entry === undefined
    ? undefined
    : decodeURIComponent(entry.slice(prefix.length));
};

export const writeCookie = (
  name: string,
  value: string,
  maxAgeSeconds = ONE_YEAR_SECONDS,
) => {
  document.cookie = [
    `${encodeURIComponent(name)}=${encodeURIComponent(value)}`,
    `path=${import.meta.env.BASE_URL}`,
    `max-age=${maxAgeSeconds}`,
    "SameSite=Lax",
  ].join("; ");
};
