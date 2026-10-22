import type { HashRouter, RouteParams } from "@/lib/hash-routers";
import { t } from "@/lib/i18n";
import type { DialogButton, DialogMessage } from "./index";
import dialog from "./index";

export type RouteDialogMessage = Omit<DialogMessage, "buttons"> & {
  actions?: DialogButton[];
};

export interface RouteDialogOptions {
  router: HashRouter;
  path: string;
  closeTo: string;
  open: (
    params: RouteParams,
    signal: AbortSignal,
  ) => Promise<RouteDialogMessage | undefined>;
  onopened?: (signal: AbortSignal) => void;
  signal: AbortSignal;
}

const dismissFrom = (token: symbol) => {
  const items = dialog.items;
  const index = items.findIndex((item) => item.token === token);
  if (index === -1) {
    return;
  }
  for (const item of items.slice(index).reverse()) {
    dialog.dismiss(item.token);
  }
};

export const routeDialog = ({
  router,
  path,
  closeTo,
  open,
  onopened,
  signal,
}: RouteDialogOptions) => {
  let active: AbortController | undefined;

  const close = () => {
    active?.abort();
    active = undefined;
  };

  const handle = (params?: RouteParams) => {
    close();
    if (params === undefined || signal.aborted) {
      return;
    }

    const controller = new AbortController();
    active = controller;
    open(params, controller.signal)
      .then((message) => {
        if (controller.signal.aborted) {
          return;
        }
        if (message === undefined) {
          router.back(closeTo);
          return;
        }

        const { actions = [], ...content } = message;
        const { token } = dialog.push({
          ...content,
          buttons: [
            ...actions,
            {
              text: t("common.close"),
              type: "",
              onclick: () => {
                router.close(closeTo);
                return false;
              },
            },
          ],
        });
        controller.signal.addEventListener("abort", () => dismissFrom(token), {
          once: true,
        });
        onopened?.(controller.signal);
      })
      .catch((err) => {
        console.error(err);
      });
  };

  signal.addEventListener("abort", close, { once: true });
  handle(router.on(path, handle, signal));
};
