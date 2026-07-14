import { create } from "zustand";
import { ROOT_SCOPE, subgraphScope, type GraphScope } from "@/lib/portTypes";

export type GraphTab =
  | { id: "root"; label: string }
  | { id: string; kind: "subgraph"; label: string };

interface GraphTabState {
  tabs: GraphTab[];
  activeTabId: string;
  openSubGraphTab: (id: string, label: string) => void;
  renameSubGraphTab: (id: string, label: string) => void;
  setActiveTab: (tabId: string) => void;
  closeTab: (tabId: string) => void;
}

const ROOT_TAB: GraphTab = { id: "root", label: "Main Graph" };

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
      tabs: [...tabs, { id, kind: "subgraph", label }],
      activeTabId: id,
    });
  },
  renameSubGraphTab: (id, label) => {
    set({
      tabs: get().tabs.map((tab) =>
        tab.id === id && tab.id !== "root" ? { ...tab, label } : tab
      ),
    });
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
}));

export function activeGraphScope(activeTabId: string): GraphScope {
  if (activeTabId === "root") return ROOT_SCOPE;
  return subgraphScope(activeTabId);
}
