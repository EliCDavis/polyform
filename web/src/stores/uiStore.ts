import { create } from "zustand";

export interface UiFlags {
  hideStats: boolean;
  hideInfo: boolean;
  hideGraph: boolean;
  hideFileControls: boolean;
  hideExportControls: boolean;
  hideWatermark: boolean;
  hideImportSubGraphs: boolean;
  canEdit: boolean;
}

interface UiState extends UiFlags {
  sidebarWidthPercent: number;
  viewportHeightPercent: number;
  showNewGraphPopup: boolean;
  setShowNewGraphPopup: (show: boolean) => void;
  sidebarCollapsed: boolean;
  setSidebarCollapsed: (collapsed: boolean) => void;
  compactLayout: boolean;
  graphTitle: string;
  setGraphTitle: (title: string) => void;
}

const SIDEBAR_COLLAPSED_KEY = "polyform.sidebarCollapsed";

const compactQuery = window.matchMedia("(max-width: 768px)");

function loadSidebarCollapsed(): boolean {
  try {
    return localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === "true";
  } catch {
    return false;
  }
}

function parseUrlFlags(): UiFlags {
  const params = new URLSearchParams(window.location.search);
  return {
    hideStats: params.get("hide-stats") === "true",
    hideInfo: params.get("hide-info") === "true",
    hideGraph: params.get("hide-graph") === "true",
    hideFileControls: params.get("hide-file-controls") === "true",
    hideExportControls: params.get("hide-export-controls") === "true",
    hideWatermark: params.get("hide-watermark") === "true",
    hideImportSubGraphs: params.get("hide-import-subgraphs") === "true",
    canEdit: params.get("can-edit") !== "false",
  };
}

export const useUiStore = create<UiState>((set, get) => ({
  ...parseUrlFlags(),
  sidebarWidthPercent: 20,
  viewportHeightPercent: 40,
  showNewGraphPopup: globalThis.RenderingConfiguration?.ShowNewGraphPopup ?? false,
  setShowNewGraphPopup: (show) => set({ showNewGraphPopup: show }),
  compactLayout: compactQuery.matches,
  sidebarCollapsed: compactQuery.matches || loadSidebarCollapsed(),
  setSidebarCollapsed: (collapsed) => {
    if (!get().compactLayout) {
      try {
        localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(collapsed));
      } catch {}
    }
    set({ sidebarCollapsed: collapsed });
  },
  graphTitle: document.title || "Polyform",
  setGraphTitle: (title) => {
    const shown = title || "Polyform";
    document.title = shown;
    set({ graphTitle: shown });
  },
}));

compactQuery.addEventListener("change", (e) => {
  useUiStore.setState(
    e.matches
      ? { compactLayout: true, sidebarCollapsed: true }
      : { compactLayout: false, sidebarCollapsed: loadSidebarCollapsed() },
  );
});
