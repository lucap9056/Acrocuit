import {
  createMessageQueue,
  type QueueEvent,
  type QueueEventListener,
} from "@/lib/message-queue";

export type Release = { release: () => boolean };

export type LoadEvent = QueueEvent<string>;

export type LoadEventListener = QueueEventListener<string>;

export interface LoadManager {
  readonly current: string[];
  begin: () => Release;
  subscribe: (listener: LoadEventListener) => void;
  unsubscribe: (listener: LoadEventListener) => void;
}

export const createLoadManager = (): LoadManager => {
  const m = createMessageQueue<string>();

  return {
    get current() {
      return m.items.map((p) => p.payload);
    },
    begin(msg: string = "") {
      const item = m.push(msg);
      return {
        release: () => {
          const r = item.dismiss();
          return r;
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
