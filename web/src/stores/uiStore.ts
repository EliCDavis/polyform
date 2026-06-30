import { create } from "zustand";

export interface UiFlags {
  hideStats: boolean;
  hideInfo: boolean;
  hideGraph: boolean;
  hideFileControls: boolean;
  hideWatermark: boolean;
  canEdit: boolean;
}

interface UiState extends UiFlags {
  sidebarWidthPercent: number;
  viewportHeightPercent: number;
  showNewGraphPopup: boolean;
  setShowNewGraphPopup: (show: boolean) => void;
}

function parseUrlFlags(): UiFlags {
  const params = new URLSearchParams(window.location.search);
  return {
    hideStats: params.get("hide-stats") === "true",
    hideInfo: params.get("hide-info") === "true",
    hideGraph: params.get("hide-graph") === "true",
    hideFileControls: params.get("hide-file-controls") === "true",
    hideWatermark: params.get("hide-watermark") === "true",
    canEdit: params.get("can-edit") !== "false",
  };
}

export const useUiStore = create<UiState>((set) => ({
  ...parseUrlFlags(),
  sidebarWidthPercent: 20,
  viewportHeightPercent: 40,
  showNewGraphPopup: globalThis.RenderingConfiguration?.ShowNewGraphPopup ?? false,
  setShowNewGraphPopup: (show) => set({ showNewGraphPopup: show }),
}));
