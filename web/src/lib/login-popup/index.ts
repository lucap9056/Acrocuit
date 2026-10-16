import type { LoginChallenge, TokenPair, Verifier } from "@/api";
import api from "@/api";
import { describeDevice } from "@/lib/device-name";

const channelName = "acrocuit:login";
const messageType = "login-callback";

type LoginCallbackMessage = {
  type: typeof messageType;
  search: string;
};

const popupName = "acrocuit-login";
const popupFeatures = "popup,width=480,height=680";

const isLoginCallbackMessage = (data: unknown): data is LoginCallbackMessage =>
  typeof data === "object" &&
  data !== null &&
  (data as LoginCallbackMessage).type === messageType &&
  typeof (data as LoginCallbackMessage).search === "string";

export type PopupCallback = {
  verifier?: Verifier;
  search: URLSearchParams;
};

export type Popup = {
  waitCallback: (signal: AbortSignal) => Promise<PopupCallback>;
};

export const openPopup = ({ url, verifier }: LoginChallenge): Popup => {
  const state = new URL(url, location.href).searchParams.get("state");
  if (!state) {
    throw new Error("The sign-in URL has no state parameter.");
  }

  const { promise, resolve, reject } = Promise.withResolvers<PopupCallback>();

  const popup = window.open("about:blank", popupName, popupFeatures);
  if (!popup) {
    throw new Error(
      "The sign-in window was blocked. Allow pop-ups for this site and try again.",
    );
  }

  popup.location.assign(url);

  return {
    waitCallback: (signal: AbortSignal) => {
      const channel = new BroadcastChannel(channelName);

      channel.onmessage = (e: MessageEvent) => {
        if (!isLoginCallbackMessage(e.data)) {
          return;
        }
        const search = new URLSearchParams(e.data.search);
        if (search.get("state") !== state) {
          return;
        }
        resolve({
          search,
          verifier,
        });
      };

      const onabort = () => {
        reject(new Error("Sign-in was cancelled."));
      };

      promise.finally(() => {
        channel.close();
        signal.removeEventListener("abort", onabort);
        popup.close();
      });

      signal.addEventListener("abort", onabort);

      if (signal.aborted) {
        onabort();
      }
      return promise;
    },
  };
};

export const notifyLoginCallback = (search: string) => {
  const channel = new BroadcastChannel(channelName);
  const message: LoginCallbackMessage = { type: messageType, search };
  channel.postMessage(message);
  channel.close();
};

export const assertNoProviderError = (
  callback: PopupCallback,
): PopupCallback => {
  const providerError =
    callback.search.get("error_description") ?? callback.search.get("error");

  if (providerError) {
    throw new Error(providerError);
  }

  return callback;
};

export const exchangeCallbackForToken = ({
  verifier,
  search,
}: PopupCallback): Promise<TokenPair> => {
  return api.auth.loginCallback(search, {
    deviceName: describeDevice(),
    verifier,
  });
};
