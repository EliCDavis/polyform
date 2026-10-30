import { create } from "zustand";
import type { CameraOrientation } from "@elicdavis/node-flow";
import { ROOT_SCOPE, subgraphScope, type GraphScope } from "@/lib/portTypes";

export enum GraphTabKind {
  Root = "root",
  SubGraph = "subgraph",
  Preview = "preview",
}

export type GraphTab = {
  id: string;
  kind: GraphTabKind;
  label: string;
  camera?: CameraOrientation;
  // A preview tab shows one node output as a model instead of a canvas.
  nodeId?: string;
  port?: string;
  scope?: string | null;
};

export type NodeFocus = { tabId: string; nodeId: string };

interface GraphTabState {
  tabs: GraphTab[];
  activeTabId: string;
  pendingFocus: NodeFocus | null;
  openSubGraphTab: (id: string, label: string) => void;
  openPreviewTab: (nodeId: string, port: string, label: string, scope: string | null) => void;
  focusNodeInTab: (tabId: string, label: string, nodeId: string) => void;
  clearPendingFocus: () => void;
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
  pendingFocus: null,
  focusNodeInTab: (tabId, label, nodeId) => {
    const { tabs } = get();
    const opened = tabId === "root" || tabs.some((t) => t.id === tabId);
    set({
      tabs: opened
        ? tabs
        : [...tabs, { id: tabId, kind: GraphTabKind.SubGraph, label }],
      activeTabId: tabId,
      pendingFocus: { tabId, nodeId },
    });
  },
  clearPendingFocus: () => set({ pendingFocus: null }),
  openPreviewTab: (nodeId, port, label, scope) => {
    const id = `preview:${scope ?? "root"}:${nodeId}:${port}`;
    const { tabs } = get();
    set({
      tabs: tabs.some((t) => t.id === id)
        ? updateTab(tabs, id, { label })
        : [...tabs, { id, kind: GraphTabKind.Preview, label, nodeId, port, scope }],
      activeTabId: id,
    });
  },
  openSubGraphTab: (id, label) => {
    const { tabs } = get();
    const existing = tabs.find((t) => t.id === id);
    if (existing) {
      set({
        tabs: updateTab(tabs, id, { label }),
        activeTabId: id,
      });
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
