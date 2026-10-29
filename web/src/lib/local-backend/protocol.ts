export type LocalRequestBody =
  | { type: "json"; value: unknown }
  | { type: "form"; fields: Record<string, FormDataEntryValue> }
  | null;

export type LocalResponseBody =
  | { type: "json"; value: unknown }
  | { type: "binary"; data: Uint8Array<ArrayBuffer> }
  | null;

export type LocalRequest = {
  id: number;
  method: string;
  path: string;
  query: Record<string, string>;
  headers: Record<string, string>;
  body: LocalRequestBody;
};

export type LocalResponse = {
  id: number;
  status: number;
  headers: Record<string, string>;
  body: LocalResponseBody;
};
