import { devicesHandlers } from "@/lib/local-backend/handlers/devices";
import { spaceBackgroundImageHandlers } from "@/lib/local-backend/handlers/space-background-image";
import { spaceGroupsHandlers } from "@/lib/local-backend/handlers/space-groups";
import { spacesHandlers } from "@/lib/local-backend/handlers/spaces";
import { switchPanelsHandlers } from "@/lib/local-backend/handlers/switch-panels";
import { switchesHandlers } from "@/lib/local-backend/handlers/switches";
import { Router } from "@/lib/local-backend/router";
import type { LocalBackend } from "@/lib/local-backend/storage/types";

export const createRouter = (storage: LocalBackend) => {
  const router = new Router();
  spaceGroupsHandlers(router, storage);
  spacesHandlers(router, storage);
  spaceBackgroundImageHandlers(router, storage);
  switchPanelsHandlers(router, storage);
  switchesHandlers(router, storage);
  devicesHandlers(router, storage);
  return router;
};
