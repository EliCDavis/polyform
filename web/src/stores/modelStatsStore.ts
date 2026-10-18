import { create } from "zustand";
import type { ModelStats } from "@/lib/ProducerView/model_stats";
import type { HierarchyEntry } from "@/lib/ProducerView/model_hierarchy";

interface ModelStatsState {
  stats: ModelStats | null;
  hierarchy: HierarchyEntry[];
  setStats: (stats: ModelStats | null, hierarchy: HierarchyEntry[]) => void;
}

export const useModelStatsStore = create<ModelStatsState>((set) => ({
  stats: null,
  hierarchy: [],
  setStats: (stats, hierarchy) => set({ stats, hierarchy }),
}));

/** Imperative API for non-React code (e.g. ProducerViewManager). */
export const modelStatsActions = {
  set(stats: ModelStats | null, hierarchy: HierarchyEntry[]): void {
    useModelStatsStore.getState().setStats(stats, hierarchy);
  },
};
