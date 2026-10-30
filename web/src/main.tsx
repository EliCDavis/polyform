import { createRoot } from "react-dom/client";
import { setTheme } from "@elicdavis/node-flow";
import { App, setupGlobalHooks } from "./App";
import "@fortawesome/fontawesome-free/css/all.min.css";
import "./styles/global.css";
import "./styles/components.css";

setTheme({
  FontFamily: "monospace",
  Minimap: {
    BackgroundColor: "rgba(15, 49, 60, 0.9)",
    BorderColor: "#154050",
  },
});
setupGlobalHooks();

const root = document.getElementById("root");
if (root) {
  createRoot(root).render(<App />);
}
