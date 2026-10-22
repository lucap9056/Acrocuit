import toast from "@/components/toast";
import {
  createMessageQueue,
  type Dismiss,
  type QueueEvent,
  type QueueEventListener,
} from "@/lib/message-queue";
import { pickFlavorText } from "./flavor-texts";

export type Release = { release: () => boolean };

export type LoadEvent = QueueEvent<string>;

export type LoadEventListener = QueueEventListener<string>;

export interface LoadManager {
  readonly current: string[];
  begin: (msg: string) => Release;
  subscribe: (listener: LoadEventListener) => void;
  unsubscribe: (listener: LoadEventListener) => void;
}

export const createLoadManager = (): LoadManager => {
  const m = createMessageQueue<string>();
  let flavor: Dismiss | undefined;

  return {
    get current() {
      return m.items.map((p) => p.payload);
    },
    begin(msg) {
      const item = m.push(msg);
      const notice = toast.push(msg);
      const flavorText = flavor ? undefined : pickFlavorText();
      if (flavorText) {
        flavor = toast.push(flavorText);
      }
      return {
        release: () => {
          notice.dismiss();
          const released = item.dismiss();
          if (m.items.length === 0) {
            flavor?.dismiss();
            flavor = undefined;
          }
          return released;
        },
      };
    },
    subscribe(listener) {
      m.subscribe(listener);
    },
    unsubscribe(listener) {
      m.unsubscribe(listener);
    },
  };
};

export default createLoadManager();
