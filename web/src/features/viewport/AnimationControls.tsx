import { useRef } from "react";
import { Pause, Play, Repeat } from "lucide-react";
import { useAnimationStore } from "@/stores/animationStore";
import { useEditorOptional } from "@/features/editor/EditorContext";
import styles from "./AnimationControls.module.css";

export function AnimationControls() {
  const clips = useAnimationStore((s) => s.clips);
  const active = useAnimationStore((s) => s.active);
  const playing = useAnimationStore((s) => s.playing);
  const loop = useAnimationStore((s) => s.loop);
  const time = useAnimationStore((s) => s.time);
  const editor = useEditorOptional();
  const resumeAfterScrub = useRef(false);

  const clip = clips[active];
  if (!editor || !clip) {
    return null;
  }
  const manager = editor.producerViewManager;

  const startScrub = () => {
    resumeAfterScrub.current = playing;
    manager.SetAnimationPlaying(false);
  };

  const endScrub = () => {
    if (resumeAfterScrub.current) {
      manager.SetAnimationPlaying(true);
    }
    resumeAfterScrub.current = false;
  };

  return (
    <div className={styles.controls}>
      <button
        className={`${styles.iconButton} ${styles.primary}`}
        onClick={() => manager.SetAnimationPlaying(!playing)}
        aria-label={playing ? "Pause" : "Play"}
        title={playing ? "Pause" : "Play"}
      >
        {playing ? <Pause size={14} /> : <Play size={14} />}
      </button>
      <button
        className={`${styles.iconButton} ${loop ? styles.active : ""}`}
        onClick={() => manager.SetAnimationLoop(!loop)}
        aria-label="Loop"
        aria-pressed={loop}
        title={loop ? "Looping" : "Play once"}
      >
        <Repeat size={14} />
      </button>
      <input
        className={styles.slider}
        type="range"
        min={0}
        max={clip.duration}
        step="any"
        value={time}
        onPointerDown={startScrub}
        onPointerUp={endScrub}
        onChange={(e) => manager.SeekAnimation(Number(e.target.value))}
        aria-label="Animation time"
      />
      <span className={styles.time}>
        {time.toFixed(2)}
        <span className={styles.unit}>/ {clip.duration.toFixed(2)}s</span>
      </span>
      {clips.length > 1 && (
        <select
          className={styles.select}
          value={active}
          onChange={(e) => manager.PlayAnimationClip(Number(e.target.value))}
          aria-label="Animation"
        >
          {clips.map((c, i) => (
            <option key={i} value={i}>
              {c.name}
            </option>
          ))}
        </select>
      )}
    </div>
  );
}
