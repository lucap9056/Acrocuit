import { mount } from "svelte";
import "@/app.css";
import { drawingFontsReady } from "@/components/iso-scene/layout/typography";
import { ready as i18nReady } from "@/lib/i18n";
import App from "@/pages/main/main.svelte";

await Promise.all([i18nReady, drawingFontsReady]);

const target =
  document.getElementById("app") ??
  (() => {
    const div = document.createElement("div");
    div.id = "app";
    document.body.appendChild(div);
    return div;
  })();

const app = mount(App, { target });

export default app;
