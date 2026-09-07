import { create } from "zustand";
import type { ModelStats } from "@/lib/ProducerView/model_stats";

interface ModelStatsState {
  stats: ModelStats | null;
  setStats: (stats: ModelStats | null) => void;
}

export const useModelStatsStore = create<ModelStatsState>((set) => ({
  stats: null,
  setStats: (stats) => set({ stats }),
}));

/** Imperative API for non-React code (e.g. ProducerViewManager). */
export const modelStatsActions = {
  set(stats: ModelStats | null): void {
    useModelStatsStore.getState().setStats(stats);
  },
};
