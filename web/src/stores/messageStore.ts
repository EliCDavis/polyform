import { create } from "zustand";

interface MessageState {
  errors: Record<string, string>;
  info: string | null;
  showError: (key: string, message: string) => void;
  clearError: (key: string) => void;
  showInfo: (message: string) => void;
  clearInfo: () => void;
}

export const useMessageStore = create<MessageState>((set) => ({
  errors: {},
  info: null,
  showError: (key, message) =>
    set((state) => ({
      errors: { ...state.errors, [key]: message },
    })),
  clearError: (key) =>
    set((state) => {
      const next = { ...state.errors };
      delete next[key];
      return { errors: next };
    }),
  showInfo: (message) => set({ info: message }),
  clearInfo: () => set({ info: null }),
}));

/** Imperative API for non-React code (e.g. ProducerViewManager). */
export const messageActions = {
  showError(key: string, message: string): void {
    useMessageStore.getState().showError(key, message);
  },
  clearError(key: string): void {
    useMessageStore.getState().clearError(key);
  },
  showInfo(message: string): void {
    useMessageStore.getState().showInfo(message);
  },
  clearInfo(): void {
    useMessageStore.getState().clearInfo();
  },
};
