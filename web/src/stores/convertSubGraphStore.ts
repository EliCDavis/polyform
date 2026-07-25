import { create } from "zustand";

interface ConvertSubGraphState {
  open: boolean;
  nodeIds: string[];
  scope: string;
  openConvert: (nodeIds: string[], scope: string) => void;
  close: () => void;
}

export const useConvertSubGraphStore = create<ConvertSubGraphState>((set) => ({
  open: false,
  nodeIds: [],
  scope: "root",
  openConvert: (nodeIds, scope) => set({ open: true, nodeIds, scope }),
  close: () => set({ open: false, nodeIds: [], scope: "root" }),
}));
