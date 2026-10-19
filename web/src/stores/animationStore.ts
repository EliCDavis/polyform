import { create } from "zustand";

export interface AnimationClipInfo {
  name: string;
  duration: number;
}

interface AnimationPlayback {
  clips: AnimationClipInfo[];
  active: number;
  playing: boolean;
  loop: boolean;
  time: number;
}

export const useAnimationStore = create<AnimationPlayback>(() => ({
  clips: [],
  active: 0,
  playing: true,
  loop: true,
  time: 0,
}));

export const animationActions = {
  set(playback: Partial<AnimationPlayback>): void {
    useAnimationStore.setState(playback);
  },
};
