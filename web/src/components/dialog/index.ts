import type { Snippet } from "svelte";
import { createMessageQueue, type QueueEvent } from "@/lib/message-queue";

export type DialogButtonHandler = () =>
  | undefined
  | boolean
  | Promise<undefined | boolean>;

export interface DialogButton {
  text: string;
  type: string;
  align?: "start";
  onclick: DialogButtonHandler;
}

export interface DialogContentProps {
  dismiss: () => boolean;
}

export interface DialogMessage {
  text?: string;
  content?: Snippet<[DialogContentProps]>;
  buttons: DialogButton[];
  maxWidth?: string;
  maxHeight?: string;
}

export type DialogEvent = QueueEvent<DialogMessage>;

export default createMessageQueue<DialogMessage>();
