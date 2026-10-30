import { create } from "zustand";

interface ProfileState {
  activeProfile: string | null;
  setActiveProfile: (name: string | null) => void;
}

const storageKey = () => `polyform.activeProfile.${document.title}`;

function loadActiveProfile(): string | null {
  try {
    return localStorage.getItem(storageKey());
  } catch {
    return null;
  }
}

export const useProfileStore = create<ProfileState>((set) => ({
  activeProfile: loadActiveProfile(),
  setActiveProfile: (name) => {
    try {
      if (name === null) localStorage.removeItem(storageKey());
      else localStorage.setItem(storageKey(), name);
    } catch {}
    set({ activeProfile: name });
  },
}));
