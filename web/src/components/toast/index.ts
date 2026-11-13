import {
  createMessageQueue,
  type Dismiss,
  type MessageQueue,
  type QueueEvent,
} from "@/lib/message-queue";

export type ToastEvent = QueueEvent<string>;

const { push, ...m } = createMessageQueue<string>();

type ToastManager = MessageQueue<string> & {
  push: (text: string, durationMs?: number) => Dismiss;
};

const toastManager: ToastManager = {
  ...m,
  push: (text: string, durationMs: number = 0) => {
    const item = push(text);
    if (durationMs > 0) {
      setTimeout(item.dismiss, durationMs);
    }
    return item;
  },
};

export default toastManager;
