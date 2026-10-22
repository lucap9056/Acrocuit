import { mount } from "svelte";
import "@/app.css";
import { ready as i18nReady } from "@/lib/i18n";
import App from "@/pages/main/main.svelte";

await i18nReady;

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
