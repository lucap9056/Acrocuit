import {
  createMessageQueue,
  type MessageQueue,
  type QueueEvent,
} from "@/lib/message-queue";

export type ToastEvent = QueueEvent<string>;

const { push, ...m } = createMessageQueue<string>();

const toastManager: MessageQueue<string> = {
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
