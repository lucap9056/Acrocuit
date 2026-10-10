import { mount } from "svelte";
import "@/app.css";
import App from "@/pages/callback/callback.svelte";

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
