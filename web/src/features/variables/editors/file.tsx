import { useState } from "react";
import { setBinaryVariableValue } from "@/api/variables";
import styles from "./editors.module.css";

function formatFileSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  const size = bytes / Math.pow(k, i);
  return `${size.toFixed(size < 10 && i > 0 ? 1 : 0)} ${units[i]}`;
}

export function ImageEditor({ variableKey }: { variableKey: string }) {
  const [src, setSrc] = useState(`./variable/value/${variableKey}?t=${Date.now()}`);
  const [missing, setMissing] = useState(false);
  return (
    <div className={styles.fileRow}>
      <a href={src} target="_blank" rel="noopener noreferrer" className={styles.thumb}>
        {!missing && <img src={src} alt="" onError={() => setMissing(true)} />}
      </a>
      <button
        type="button"
        className={styles.setButton}
        onClick={() =>
          setBinaryVariableValue(variableKey, () => {
            setMissing(false);
            setSrc(`./variable/value/${variableKey}?t=${Date.now()}`);
          })
        }
      >
        Set…
      </button>
    </div>
  );
}

export function FileEditor({ variableKey, size }: { variableKey: string; size: number }) {
  return (
    <div className={styles.fileRow}>
      <span className={styles.fileSize}>{formatFileSize(size)}</span>
      <button type="button" className={styles.setButton} onClick={() => setBinaryVariableValue(variableKey)}>
        Set…
      </button>
    </div>
  );
}
