export type RouteParams = Record<string, string>;

export type RouteHandler = (params?: RouteParams) => void;

interface HashRouter {
  on: (path: string, callback: RouteHandler) => RouteParams | undefined;
  off: (path: string, callback: RouteHandler) => void;
  to: (path: string) => void;
  back: (fallbackPath?: string) => void;
}

enum RouteTransitionType {
  OUT_OF_REGION,
  PARAM_CHANGE,
  NONE,
}

const matchRoutePattern = (hash: string, pattern: string): boolean => {
  if (typeof hash !== "string" || typeof pattern !== "string") {
    throw new Error(
      "[HashRouter] Invalid arguments: hash and pattern must be strings.",
    );
  }

  const cleanHash = `/${hash.replace(/^#/, "").replace(/^\//, "")}`;
  const cleanPattern = `/${pattern.replace(/^\//, "")}`;

  const regexPath = cleanPattern
    .replace(/:([^/]+)/g, "([^/]+)")
    .replace(/\*/g, ".*");

  const regex = new RegExp(`^${regexPath}$`);
  return regex.test(cleanHash);
};

const getRouteTransitionType = (
  previousHash: string,
  currentHash: string,
  path: string,
): RouteTransitionType => {
  const wasMatched = matchRoutePattern(previousHash, path);
  const isMatched = matchRoutePattern(currentHash, path);

  if (!wasMatched && isMatched) {
    return RouteTransitionType.PARAM_CHANGE;
  }

  if (wasMatched && !isMatched) {
    return RouteTransitionType.OUT_OF_REGION;
  }

  if (wasMatched && isMatched && previousHash !== currentHash) {
    return RouteTransitionType.PARAM_CHANGE;
  }

  return RouteTransitionType.NONE;
};

const extractRouteParameters = (
  transitionType: RouteTransitionType,
  template: string,
  currentHash: string,
): RouteParams | undefined => {
  if (transitionType === RouteTransitionType.OUT_OF_REGION) {
    return undefined;
  }

  const cleanHash = `/${currentHash.replace(/^#/, "").replace(/^\//, "")}`;
  const cleanTemplate = `/${template.replace(/^\//, "")}`;

  const paramNames: string[] = [];
  const regexPattern = cleanTemplate
    .replace(/:([^/]+)/g, (_, paramName) => {
      paramNames.push(paramName);
      return "([^/]+)";
    })
    .replace(/\*/g, "(.*)");

  const regex = new RegExp(`^${regexPattern}$`);
  const match = cleanHash.match(regex);

  if (!match) {
    return undefined;
  }

  const params: RouteParams = {};
  paramNames.forEach((name, index) => {
    params[name] = match[index + 1];
  });

  return params;
};

const getHistoryIndex = (): number | undefined => {
  const index = window.history.state?.hashRouterIndex;
  return typeof index === "number" ? index : undefined;
};

const stampHistoryIndex = (index: number) => {
  window.history.replaceState(
    { ...window.history.state, hashRouterIndex: index },
    "",
  );
};

export const createHashRouter = (): HashRouter => {
  const routeHandlers: Map<string, Set<RouteHandler>> = new Map();
  let previousHash = "";

  const getCurrentHash = (): string => {
    if (typeof window === "undefined") {
      return "";
    }
    return window.location.hash.replace(/^#/, "") || "";
  };

  const handleHashChange = () => {
    const prevHash = previousHash;
    const currHash = getCurrentHash();
    previousHash = currHash;

    for (const [path, handlers] of routeHandlers) {
      const transitionType = getRouteTransitionType(prevHash, currHash, path);
      if (transitionType !== RouteTransitionType.NONE) {
        const params = extractRouteParameters(transitionType, path, currHash);
        for (const handler of handlers) {
          try {
            handler(params);
          } catch (error) {
            console.error(
              `[HashRouter] Error executing handler for route "${path}":`,
              error,
            );
          }
        }
      }
    }
  };

  if (typeof window !== "undefined") {
    if (getHistoryIndex() === undefined) {
      stampHistoryIndex(0);
    }

    let lastHistoryIndex = getHistoryIndex() ?? 0;
    window.addEventListener("hashchange", () => {
      const existingIndex = getHistoryIndex();
      if (existingIndex === undefined) {
        lastHistoryIndex += 1;
        stampHistoryIndex(lastHistoryIndex);
      } else {
        lastHistoryIndex = existingIndex;
      }
      handleHashChange();
    });
  } else {
    console.warn(
      "[HashRouter] Running in a non-browser environment; window object is undefined.",
    );
  }

  handleHashChange();

  return {
    on: (path: string, callback: RouteHandler): RouteParams | undefined => {
      if (path.trim() === "") {
        throw new Error("[HashRouter] Route path must be a non-empty string.");
      }

      const currHash = getCurrentHash();

      const handlersSet =
        routeHandlers.get(path) ||
        (() => {
          const newSet = new Set<RouteHandler>();
          routeHandlers.set(path, newSet);
          return newSet;
        })();

      handlersSet.add(callback);

      const transitionType = getRouteTransitionType("", currHash, path);
      return extractRouteParameters(transitionType, path, currHash);
    },
    off: (path: string, callback: RouteHandler) => {
      const handlersSet = routeHandlers.get(path);
      if (handlersSet !== undefined) {
        handlersSet.delete(callback);
        if (handlersSet.size === 0) {
          routeHandlers.delete(path);
        }
      }
    },
    to: (path: string) => {
      const targetHash = `/${path.replace(/^#/, "").replace(/^\//, "")}`;
      if (getCurrentHash() !== targetHash) {
        window.location.hash = targetHash;
      }
    },
    back: (fallbackPath: string = "") => {
      if ((getHistoryIndex() ?? 0) > 0) {
        window.history.back();
      } else {
        const targetHash = `/${fallbackPath.replace(/^#/, "").replace(/^\//, "")}`;
        window.location.replace(`#${targetHash}`);
      }
    },
  };
};

export default createHashRouter();

export const field = (p: RouteParams, k: string, d: string): string => {
  if (typeof p[k] === "string") {
    return k;
  }
  return d;
};
