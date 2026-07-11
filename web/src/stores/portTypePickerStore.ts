import { create } from "zustand";

export interface PortTypePickerRequest {
  title: string;
  options: string[];
  current: string;
  onSelect: (type: string) => void;
  onCancel?: () => void;
}

interface PortTypePickerState {
  request: PortTypePickerRequest | null;
  open: (request: PortTypePickerRequest) => void;
  close: () => void;
  confirm: (type: string) => void;
}

export const usePortTypePickerStore = create<PortTypePickerState>((set, get) => ({
  request: null,
  open: (request) => set({ request }),
  close: () => {
    const request = get().request;
    set({ request: null });
    request?.onCancel?.();
  },
  confirm: (type) => {
    const request = get().request;
    set({ request: null });
    request?.onSelect(type);
  },
}));

/** Imperative API for node-flow widgets and other non-React code. */
export const portTypePickerActions = {
  show(config: PortTypePickerRequest): void {
    usePortTypePickerStore.getState().open(config);
  },
};
