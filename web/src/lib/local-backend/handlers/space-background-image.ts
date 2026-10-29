import type { BackgroundImage, SpaceId } from "@/api/models";
import type { LocalRequest } from "@/lib/local-backend/protocol";
import { error, ok, type Reply } from "@/lib/local-backend/response";
import { formField, type Router } from "@/lib/local-backend/router";
import type { LocalBackend } from "@/lib/local-backend/storage/types";

const backgroundImageField = "image";
const backgroundImage = "/api/spaces/:spaceId/background-image";

const startsWith = (data: Uint8Array, signature: string, offset = 0) =>
  [...signature].every((char, i) => data[offset + i] === char.charCodeAt(0));

const detectImageType = (data: Uint8Array) => {
  if (startsWith(data, "\x89PNG\r\n\x1a\n")) {
    return "image/png";
  }
  if (startsWith(data, "\xff\xd8\xff")) {
    return "image/jpeg";
  }
  if (startsWith(data, "GIF87a") || startsWith(data, "GIF89a")) {
    return "image/gif";
  }
  if (startsWith(data, "RIFF") && startsWith(data, "WEBPVP", 8)) {
    return "image/webp";
  }
  return null;
};

const imageReply = (request: LocalRequest, image: BackgroundImage): Reply => {
  const etag = `"${image.updated_at}"`;
  const headers = {
    etag,
    "cache-control": "private, no-cache",
    "x-content-type-options": "nosniff",
  };
  if (request.headers["if-none-match"] === etag) {
    return { status: 304, headers, body: null };
  }

  return {
    status: 200,
    headers: { ...headers, "content-type": image.content_type },
    body: { type: "binary", data: image.data },
  };
};

export const spaceBackgroundImageHandlers = (
  router: Router,
  { Spaces }: LocalBackend,
) => {
  router.get(backgroundImage, ({ request, params }) =>
    Spaces.GetSpaceBackgroundImage(params.spaceId as SpaceId).then((image) =>
      imageReply(request, image),
    ),
  );

  router.put(backgroundImage, ({ request, params }) => {
    const file = formField(request, backgroundImageField);
    if (!(file instanceof Blob)) {
      return error(
        400,
        `multipart field "${backgroundImageField}" is required`,
      );
    }

    return file.arrayBuffer().then((buffer) => {
      const data = new Uint8Array(buffer);
      const contentType = detectImageType(data);
      if (!contentType) {
        return error(415, "background image must be PNG, JPEG, WebP or GIF");
      }

      return Spaces.SetSpaceBackgroundImage(
        params.spaceId as SpaceId,
        contentType,
        data,
      ).then(ok);
    });
  });

  router.delete(backgroundImage, ({ params }) =>
    Spaces.DelSpaceBackgroundImage(params.spaceId as SpaceId).then(ok),
  );
};
