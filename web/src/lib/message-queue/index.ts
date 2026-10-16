export type Message<T> = {
  readonly token: symbol;
  readonly payload: T;
};

export type Dismiss = { dismiss: () => boolean };

export type QueueEvent<T> = { value: { messages: Message<T>[] } };

export type QueueEventListener<T> = (e: QueueEvent<T>) => void;

export interface MessageQueue<T> {
  readonly items: Message<T>[];
  push: (payload: T) => Dismiss;
  dismiss: (token: symbol) => boolean;
  subscribe: (listener: QueueEventListener<T>) => void;
  unsubscribe: (listener: QueueEventListener<T>) => void;
}

export const createMessageQueue = <T>(): MessageQueue<T> => {
  let items: readonly Message<T>[] = [];
  const listeners = new Set<QueueEventListener<T>>();

  const update = (next: readonly Message<T>[]) => {
    items = next;
    const messages = [...next];
    for (const listener of listeners) {
      listener({ value: { messages } });
    }
  };

  const dismiss = (token: symbol) => {
    const next = items.filter((m) => m.token !== token);
    if (next.length === items.length) {
      return false;
    }
    update(next);
    return true;
  };

  return {
    get items() {
      return [...items];
    },
    push(payload) {
      const token = Symbol();
      update([...items, { token, payload }]);
      return {
        dismiss: () => dismiss(token),
      };
    },
    dismiss,
    subscribe(listener) {
      listeners.add(listener);
    },
    unsubscribe(listener) {
      listeners.delete(listener);
    },
  };
};
