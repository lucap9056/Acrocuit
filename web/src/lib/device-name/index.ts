const browsers: [marker: string, name: string][] = [
  ["Edg/", "Edge"],
  ["OPR/", "Opera"],
  ["Firefox/", "Firefox"],
  ["Chrome/", "Chrome"],
  ["Safari/", "Safari"],
];

const systems: [marker: string, name: string][] = [
  ["Windows", "Windows"],
  ["Android", "Android"],
  ["iPhone", "iOS"],
  ["iPad", "iPadOS"],
  ["Mac OS X", "macOS"],
  ["CrOS", "ChromeOS"],
  ["Linux", "Linux"],
];

const find = (table: [string, string][], userAgent: string) =>
  table.find(([marker]) => userAgent.includes(marker))?.[1];

export const describeDevice = (userAgent = navigator.userAgent) => {
  const browser = find(browsers, userAgent) ?? "Browser";
  const system = find(systems, userAgent);
  return system ? `${browser} on ${system}` : browser;
};
