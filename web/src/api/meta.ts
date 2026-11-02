import type { Meta } from "@/api/models";
import { unwrapApiResponse } from "@/api/response";
import { appUrl } from "@/lib/app-url";

const getMeta = () => {
  return fetch(appUrl("api/meta"), { method: "GET" }).then(
    unwrapApiResponse<Meta>,
  );
};

export default {
  getMeta,
};
