import { mount } from "svelte";
import "@/app.css";
import { appUrl } from "@/lib/app-url";
import { ready as i18nReady } from "@/lib/i18n";

const isLoginCallback = location.pathname === appUrl("callback").pathname;

const loadMainPage = () =>
  Promise.all([
    import("@/pages/main/main.svelte"),
    import("@/components/iso-scene/layout/typography").then(
      (typography) => typography.drawingFontsReady,
    ),
  ]).then(([page]) => page.default);

const loadCallbackPage = () =>
  import("@/pages/callback/callback.svelte").then((page) => page.default);

const [Page] = await Promise.all([
  isLoginCallback ? loadCallbackPage() : loadMainPage(),
  i18nReady,
]);

const target =
  document.getElementById("app") ??
  (() => {
    const div = document.createElement("div");
    div.id = "app";
    document.body.appendChild(div);
    return div;
  })();

const app = mount(Page, { target });

export default app;
