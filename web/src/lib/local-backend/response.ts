import type { LocalResponse } from "@/lib/local-backend/protocol";

export type Reply = Omit<LocalResponse, "id">;

type StorageError = { status: number; message: string };

const json = (status: number, value: unknown): Reply => ({
  status,
  headers: {},
  body: { type: "json", value },
});

export const ok = (data: unknown) => json(200, { success: true, data });

export const empty = () => json(200, { success: true, data: null });

export const error = (status: number, message: string) =>
  json(status, { success: false, data: null, error: message });

const isStorageError = (err: unknown): err is StorageError =>
  typeof err === "object" &&
  err !== null &&
  typeof (err as StorageError).status === "number" &&
  typeof (err as StorageError).message === "string";

export const failure = (err: unknown) => {
  if (isStorageError(err)) {
    return error(err.status, err.message);
  }
  console.error(err);
  return error(500, "internal server error");
};
