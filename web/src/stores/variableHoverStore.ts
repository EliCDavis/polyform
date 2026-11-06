import { create } from "zustand";

export type VariableHoverSource = "sidebar" | "graph";

interface VariableHoverState {
  hovered: string | null;
  source: VariableHoverSource;
  setHovered: (variableKey: string | null, source: VariableHoverSource) => void;
}

export const useVariableHoverStore = create<VariableHoverState>((set) => ({
  hovered: null,
  source: "sidebar",
  setHovered: (hovered, source) => set({ hovered, source }),
}));
