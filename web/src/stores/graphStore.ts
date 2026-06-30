import { create } from "zustand";
import type { GraphInstance } from "@/types/schema";

interface GraphState {
  graph: GraphInstance | null;
  modelVersion: number;
  setGraph: (graph: GraphInstance) => void;
  setModelVersion: (version: number) => void;
}

export const useGraphStore = create<GraphState>((set) => ({
  graph: null,
  modelVersion: 0,
  setGraph: (graph) => set({ graph }),
  setModelVersion: (modelVersion) => set({ modelVersion }),
}));
