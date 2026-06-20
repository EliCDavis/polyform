import { createRoot } from "react-dom/client";
import { App, setupGlobalHooks } from "./App";
import "@fortawesome/fontawesome-free/css/all.min.css";
import "./styles/global.css";
import "./styles/components.css";

setupGlobalHooks();

const root = document.getElementById("root");
if (root) {
  createRoot(root).render(<App />);
}
