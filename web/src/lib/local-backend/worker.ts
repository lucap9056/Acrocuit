import { createRouter } from "@/lib/local-backend/handlers";
import type { LocalRequest, LocalResponse } from "@/lib/local-backend/protocol";
import { error } from "@/lib/local-backend/response";
import { openStorage } from "@/lib/local-backend/storage";

const router = openStorage().then(createRouter);

const unavailable = (err: unknown, request: LocalRequest): LocalResponse => {
  console.error(err);
  return { id: request.id, ...error(503, "local storage unavailable") };
};

const transferables = ({ body }: LocalResponse) =>
  body?.type === "binary" ? [body.data.buffer] : [];

self.onmessage = ({ data }: MessageEvent<LocalRequest>) => {
  router
    .then((r) => r.handle(data))
    .catch((err) => unavailable(err, data))
    .then((response) =>
      self.postMessage(response, { transfer: transferables(response) }),
    );
};
