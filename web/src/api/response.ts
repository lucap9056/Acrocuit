type Body<T> =
  | { success: true; data: T }
  | { success: false; data: null; error: string };

export class ApiError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export const unwrapApiResponse = async <T>(r: Response): Promise<T> => {
  const body = await r.json().then(
    (b) => b as Body<T>,
    () => null,
  );

  if (r.ok && body?.success) {
    return body.data;
  }

  const message =
    body && !body.success && typeof body.error === "string"
      ? body.error
      : r.statusText || `HTTP ${r.status}`;

  throw new ApiError(message, r.status);
};
