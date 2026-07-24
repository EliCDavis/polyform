import { create } from "zustand";
import type { CameraOrientation } from "@elicdavis/node-flow";
import { ROOT_SCOPE, subgraphScope, type GraphScope } from "@/lib/portTypes";

export enum GraphTabKind {
  Root = "root",
  SubGraph = "subgraph",
}

export type GraphTab = {
  id: string;
  kind: GraphTabKind;
  label: string;
  camera?: CameraOrientation;
};

interface GraphTabState {
  tabs: GraphTab[];
  activeTabId: string;
  openSubGraphTab: (id: string, label: string) => void;
  renameSubGraphTab: (id: string, label: string) => void;
  setActiveTab: (tabId: string) => void;
  closeTab: (tabId: string) => void;
  setTabCamera: (tabId: string, camera: CameraOrientation) => void;
  getTab: (tabId: string) => GraphTab | undefined;
}

const ROOT_TAB: GraphTab = {
  id: "root",
  kind: GraphTabKind.Root,
  label: "Main Graph",
};

function copyCamera(camera: CameraOrientation): CameraOrientation {
  return {
    position: { x: camera.position.x, y: camera.position.y },
    zoom: camera.zoom,
  };
}

function updateTab(
  tabs: GraphTab[],
  tabId: string,
  patch: Partial<GraphTab>
): GraphTab[] {
  return tabs.map((tab) => (tab.id === tabId ? { ...tab, ...patch } : tab));
}

export const useGraphTabStore = create<GraphTabState>((set, get) => ({
  tabs: [ROOT_TAB],
  activeTabId: "root",
  openSubGraphTab: (id, label) => {
    const { tabs } = get();
    const existing = tabs.find((t) => t.id === id);
    if (existing) {
      set({ activeTabId: id });
      return;
    }
    set({
      tabs: [...tabs, { id, kind: GraphTabKind.SubGraph, label }],
      activeTabId: id,
    });
  },
  renameSubGraphTab: (id, label) => {
    set({ tabs: updateTab(get().tabs, id, { label }) });
  },
  setActiveTab: (tabId) => set({ activeTabId: tabId }),
  closeTab: (tabId) => {
    if (tabId === "root") return;
    const { tabs, activeTabId } = get();
    const nextTabs = tabs.filter((t) => t.id !== tabId);
    set({
      tabs: nextTabs.length ? nextTabs : [ROOT_TAB],
      activeTabId: activeTabId === tabId ? "root" : activeTabId,
    });
  },
  setTabCamera: (tabId, camera) => {
    set({ tabs: updateTab(get().tabs, tabId, { camera: copyCamera(camera) }) });
  },
  getTab: (tabId) => get().tabs.find((t) => t.id === tabId),
}));

export function activeGraphScope(activeTabId: string): GraphScope {
  if (activeTabId === "root") return ROOT_SCOPE;
  return subgraphScope(activeTabId);
}
