import type { LocalRequest, LocalResponse } from "@/lib/local-backend/protocol";
import { error, failure, type Reply } from "@/lib/local-backend/response";

export type Context = {
  request: LocalRequest;
  params: Record<string, string>;
};

export type Handler = (context: Context) => Reply | Promise<Reply>;

type Route = { method: string; segments: string[]; handler: Handler };

const splitPath = (path: string) => path.split("/").filter(Boolean);

const matchSegments = (pattern: string[], path: string[]) => {
  if (pattern.length !== path.length) {
    return null;
  }

  const params: Record<string, string> = {};
  for (const [i, segment] of pattern.entries()) {
    if (segment.startsWith(":")) {
      params[segment.slice(1)] = decodeURIComponent(path[i]);
    } else if (segment !== path[i]) {
      return null;
    }
  }
  return params;
};

export class Router {
  private readonly routes: Route[] = [];

  private on(method: string, pattern: string, handler: Handler) {
    this.routes.push({ method, segments: splitPath(pattern), handler });
  }

  get(pattern: string, handler: Handler) {
    this.on("GET", pattern, handler);
  }

  post(pattern: string, handler: Handler) {
    this.on("POST", pattern, handler);
  }

  put(pattern: string, handler: Handler) {
    this.on("PUT", pattern, handler);
  }

  patch(pattern: string, handler: Handler) {
    this.on("PATCH", pattern, handler);
  }

  delete(pattern: string, handler: Handler) {
    this.on("DELETE", pattern, handler);
  }

  handle(request: LocalRequest): Promise<LocalResponse> {
    return this.dispatch(request).then((reply) => ({
      id: request.id,
      ...reply,
    }));
  }

  private dispatch(request: LocalRequest): Promise<Reply> {
    const path = splitPath(request.path);

    for (const route of this.routes) {
      if (route.method !== request.method) {
        continue;
      }
      const params = matchSegments(route.segments, path);
      if (!params) {
        continue;
      }
      return Promise.resolve({ request, params })
        .then(route.handler)
        .catch(failure);
    }

    return Promise.resolve(error(404, "not found"));
  }
}

export const json = <T>(request: LocalRequest) =>
  (request.body?.type === "json" ? request.body.value : {}) as T;

export const formField = (request: LocalRequest, name: string) =>
  request.body?.type === "form" ? request.body.fields[name] : undefined;

export const include = (request: LocalRequest) => request.query.include ?? "";
