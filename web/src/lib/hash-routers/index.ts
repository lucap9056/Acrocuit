export type RouteParams = Record<string, string>;

export type RouteHandler = (params?: RouteParams) => void;

export interface HashRouter {
  on: (
    path: string,
    handler: RouteHandler,
    signal: AbortSignal,
  ) => RouteParams | undefined;
  to: (path: string) => void;
  back: (fallbackPath?: string) => void;
  backTo: (path: string) => void;
  child: (prefix: string) => HashRouter;
}

export interface RootHashRouter extends HashRouter {
  dispose: () => void;
}

interface CompiledRoute {
  regex: RegExp;
  paramNames: string[];
}

interface Registration {
  route: CompiledRoute;
  handler: RouteHandler;
}

const trimPath = (path: string): string =>
  path.replace(/^#/, "").replace(/^\/+|\/+$/g, "");

const joinPath = (...paths: string[]): string =>
  `/${paths
    .map(trimPath)
    .filter((p) => p !== "")
    .join("/")}`;

const escapeRegExp = (text: string): string =>
  text.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");

const compileRoute = (pattern: string): CompiledRoute => {
  const segments = trimPath(pattern)
    .split("/")
    .filter((s) => s !== "");
  const paramNames: string[] = [];

  const body = segments
    .map((segment, index) => {
      if (segment.startsWith(":")) {
        paramNames.push(segment.slice(1));
        return "/([^/]+)";
      }
      if (segment === "*") {
        return index === segments.length - 1 ? "(?:/.*)?" : "/[^/]+";
      }
      return `/${escapeRegExp(segment)}`;
    })
    .join("");

  return { regex: new RegExp(`^${body || "/"}$`), paramNames };
};

const matchRoute = (
  route: CompiledRoute,
  path: string,
): RouteParams | undefined => {
  const match = route.regex.exec(path);
  if (!match) {
    return undefined;
  }

  const params: RouteParams = {};
  route.paramNames.forEach((name, index) => {
    params[name] = match[index + 1];
  });
  return params;
};

const isSameParams = (a?: RouteParams, b?: RouteParams): boolean => {
  if (a === undefined || b === undefined) {
    return a === b;
  }
  const keys = Object.keys(a);
  return (
    keys.length === Object.keys(b).length && keys.every((k) => a[k] === b[k])
  );
};

const getCurrentPath = (): string =>
  typeof window === "undefined" ? "/" : joinPath(window.location.hash);

const getHistoryIndex = (): number | undefined => {
  const index = window.history.state?.hashRouterIndex;
  return typeof index === "number" ? index : undefined;
};

const getHistoryPreviousPath = (): string | undefined => {
  const path = window.history.state?.hashRouterPreviousPath;
  return typeof path === "string" ? path : undefined;
};

interface HistoryEntry {
  index: number;
  previousPath?: string;
}

const stampHistoryEntry = ({ index, previousPath }: HistoryEntry) => {
  window.history.replaceState(
    {
      ...window.history.state,
      hashRouterIndex: index,
      hashRouterPreviousPath: previousPath,
    },
    "",
  );
};

export const createHashRouter = (): RootHashRouter => {
  const registrations = new Set<Registration>();
  const listenerController = new AbortController();
  let previousPath = getCurrentPath();
  let replacedHistoryEntry: HistoryEntry | undefined;

  const handleHashChange = () => {
    const prevPath = previousPath;
    const currPath = getCurrentPath();
    previousPath = currPath;

    for (const registration of [...registrations]) {
      if (!registrations.has(registration)) {
        continue;
      }

      const { route, handler } = registration;
      const prevParams = matchRoute(route, prevPath);
      const currParams = matchRoute(route, currPath);
      if (isSameParams(prevParams, currParams)) {
        continue;
      }

      try {
        handler(currParams);
      } catch (error) {
        console.error(
          `[HashRouter] Error executing handler for route "${route.regex}":`,
          error,
        );
      }
    }
  };

  if (typeof window !== "undefined") {
    if (getHistoryIndex() === undefined) {
      stampHistoryEntry({ index: 0 });
    }

    let lastHistoryIndex = getHistoryIndex() ?? 0;
    window.addEventListener(
      "hashchange",
      () => {
        const existingIndex = getHistoryIndex();
        if (existingIndex === undefined) {
          const entry = replacedHistoryEntry ?? {
            index: lastHistoryIndex + 1,
            previousPath,
          };
          replacedHistoryEntry = undefined;
          lastHistoryIndex = entry.index;
          stampHistoryEntry(entry);
        } else {
          lastHistoryIndex = existingIndex;
        }
        handleHashChange();
      },
      { signal: listenerController.signal },
    );
  } else {
    console.warn(
      "[HashRouter] Running in a non-browser environment; window object is undefined.",
    );
  }

  const register = (
    path: string,
    handler: RouteHandler,
    signal: AbortSignal,
  ): RouteParams | undefined => {
    const route = compileRoute(path);
    const params = matchRoute(route, getCurrentPath());
    if (signal.aborted) {
      return params;
    }

    const registration: Registration = { route, handler };
    registrations.add(registration);
    signal.addEventListener(
      "abort",
      () => {
        registrations.delete(registration);
      },
      { once: true },
    );
    return params;
  };

  const navigate = (path: string) => {
    if (getCurrentPath() !== path) {
      window.location.hash = path;
    }
  };

  const back = (fallbackPath: string) => {
    if ((getHistoryIndex() ?? 0) > 0) {
      window.history.back();
    } else {
      replacedHistoryEntry = {
        index: getHistoryIndex() ?? 0,
        previousPath: getHistoryPreviousPath(),
      };
      window.location.replace(`#${fallbackPath}`);
    }
  };

  const backTo = (path: string) => {
    if ((getHistoryIndex() ?? 0) > 0 && getHistoryPreviousPath() === path) {
      window.history.back();
    } else {
      navigate(path);
    }
  };

  const createScope = (prefix: string): HashRouter => ({
    on: (path, handler, signal) =>
      register(joinPath(prefix, path), handler, signal),
    to: (path) => navigate(joinPath(prefix, path)),
    back: (fallbackPath = "") => back(joinPath(prefix, fallbackPath)),
    backTo: (path) => backTo(joinPath(prefix, path)),
    child: (childPrefix) => createScope(joinPath(prefix, childPrefix)),
  });

  return {
    ...createScope(""),
    dispose: () => {
      listenerController.abort();
      registrations.clear();
    },
  };
};

export const field = (p: RouteParams, k: string, d: string): string => {
  if (typeof p[k] === "string") {
    return p[k];
  }
  return d;
};
