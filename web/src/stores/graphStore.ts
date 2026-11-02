import { create } from "zustand";
import type { GraphInstance } from "@/types/schema";

interface GraphState {
  graph: GraphInstance | null;
  revision: number;
  setGraph: (graph: GraphInstance) => void;
  setRevision: (revision: number) => void;
}

export const useGraphStore = create<GraphState>((set) => ({
  graph: null,
  revision: 0,
  setGraph: (graph) => set({ graph }),
  setRevision: (revision) => set({ revision }),
}));
