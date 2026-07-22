import { create } from "zustand";

interface ConvertSubGraphState {
  open: boolean;
  nodeIds: string[];
  scope: string;
  pendingCenterOnGraph: boolean;
  openConvert: (nodeIds: string[], scope: string) => void;
  close: () => void;
  requestCenterOnGraph: () => void;
  consumePendingCenterOnGraph: () => boolean;
}

export const useConvertSubGraphStore = create<ConvertSubGraphState>((set, get) => ({
  open: false,
  nodeIds: [],
  scope: "root",
  pendingCenterOnGraph: false,
  openConvert: (nodeIds, scope) =>
    set({ open: true, nodeIds, scope, pendingCenterOnGraph: false }),
  close: () => set({ open: false, nodeIds: [], scope: "root" }),
  requestCenterOnGraph: () => set({ pendingCenterOnGraph: true }),
  consumePendingCenterOnGraph: () => {
    if (!get().pendingCenterOnGraph) return false;
    set({ pendingCenterOnGraph: false });
    return true;
  },
}));
