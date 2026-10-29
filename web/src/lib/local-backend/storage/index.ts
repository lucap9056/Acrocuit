import sqlite3InitModule from "@sqlite.org/sqlite-wasm";
import wasmUrl from "@wasm/acrocuit.wasm?url";
import wasmExecUrl from "@wasm/wasm_exec.js?url";
import type { LocalBackend } from "@/lib/local-backend/storage/types";

type GoRuntime = {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
};

declare global {
  var Go: new () => GoRuntime;
  var acrocuitBoot: {
    sqlite3: unknown;
    db: unknown;
    resolve(storage: LocalBackend): void;
    reject(message: string): void;
  };
}

const databaseFile = "/acrocuit.sqlite3";

export const openStorage = async () => {
  const sqlite3 = await sqlite3InitModule();
  const pool = await sqlite3.installOpfsSAHPoolVfs({});
  const db = new pool.OpfsSAHPoolDb(databaseFile);

  await import(/* @vite-ignore */ wasmExecUrl);
  const go = new Go();
  const { instance } = await WebAssembly.instantiateStreaming(
    fetch(wasmUrl),
    go.importObject,
  );

  return new Promise<LocalBackend>((resolve, reject) => {
    globalThis.acrocuitBoot = {
      sqlite3,
      db,
      resolve,
      reject: (message) => reject(new Error(message)),
    };
    go.run(instance);
  });
};
