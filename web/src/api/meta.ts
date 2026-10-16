import type { Meta } from "@/api/models";
import { unwrapApiResponse } from "@/api/response";

const getMeta = () => {
  return fetch("/api/meta", { method: "GET" }).then(unwrapApiResponse<Meta>);
};

export default {
  getMeta,
};
