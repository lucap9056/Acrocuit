import type {
  LocalRequest,
  LocalRequestBody,
  LocalResponse,
} from "@/lib/local-backend/protocol";
import { error } from "@/lib/local-backend/response";
import LocalBackendWorker from "@/lib/local-backend/worker?worker";

type Pending = {
  resolve(response: LocalResponse): void;
  reject(reason: unknown): void;
};

const localOrigin = "http://acrocuit.local";
const jsonContentType = "application/json; charset=utf-8";

const toRequest = (input: RequestInfo | URL, init: RequestInit) =>
  input instanceof Request
    ? new Request(input, init)
    : new Request(new URL(input, localOrigin), init);

const readBody = (
  request: Request,
  contentType: string,
): Promise<LocalRequestBody> => {
  if (contentType.startsWith("application/json")) {
    return request.json().then((value) => ({ type: "json", value }));
  }
  if (contentType.startsWith("multipart/form-data")) {
    return request
      .formData()
      .then((form) => ({ type: "form", fields: Object.fromEntries(form) }));
  }
  return Promise.resolve(null);
};

const toLocalRequest = (id: number, request: Request) => {
  const url = new URL(request.url);
  const headers = Object.fromEntries(request.headers);

  return readBody(request, headers["content-type"] ?? "").then(
    (body): LocalRequest => ({
      id,
      method: request.method,
      path: url.pathname,
      query: Object.fromEntries(url.searchParams),
      headers,
      body,
    }),
  );
};

const toResponse = ({ status, headers, body }: Omit<LocalResponse, "id">) => {
  if (body?.type === "json") {
    return new Response(JSON.stringify(body.value), {
      status,
      headers: { ...headers, "content-type": jsonContentType },
    });
  }
  return new Response(body?.data ?? null, { status, headers });
};

const invalidBody = (err: unknown) =>
  toResponse(
    error(400, err instanceof Error ? err.message : "invalid request body"),
  );

export class LocalAuthSession {
  readonly kind = "local";

  private readonly worker = new LocalBackendWorker();
  private readonly pending = new Map<number, Pending>();
  private nextId = 0;

  constructor() {
    this.worker.onmessage = ({ data }: MessageEvent<LocalResponse>) => {
      this.pending.get(data.id)?.resolve(data);
      this.pending.delete(data.id);
    };
    this.worker.onerror = (event) => {
      this.rejectAll(new Error(event.message || "local backend crashed"));
    };
  }

  public fetch(input: RequestInfo | URL, init: RequestInit = {}) {
    return toLocalRequest(this.nextId++, toRequest(input, init)).then(
      (request) => this.send(request).then(toResponse),
      invalidBody,
    );
  }

  private send(request: LocalRequest) {
    return new Promise<LocalResponse>((resolve, reject) => {
      this.pending.set(request.id, { resolve, reject });
      this.worker.postMessage(request);
    });
  }

  private rejectAll(reason: Error) {
    for (const { reject } of this.pending.values()) {
      reject(reason);
    }
    this.pending.clear();
  }
}
