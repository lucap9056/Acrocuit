import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vite";

const acrocuitUrl = process.env.ACROCUIT_URL ?? "http://localhost:8080";

// https://vite.dev/config/
export default defineConfig({
  base: "./",
  plugins: [svelte()],
  resolve: {
    tsconfigPaths: true,
  },
  optimizeDeps: {
    exclude: ["@sqlite.org/sqlite-wasm"],
  },
  worker: {
    format: "es",
  },
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      "/api": acrocuitUrl,
      "/auth": acrocuitUrl,
      "/dex": acrocuitUrl,
    },
  },
});
