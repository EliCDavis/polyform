import { useState } from "react";
import { setBinaryVariableValue } from "@/api/variables";

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
  return (
    <div className="variable-inputs">
      <img src={src} alt={variableKey} style={{ maxWidth: "100%" }} />
      <button type="button" onClick={() => setBinaryVariableValue(variableKey, () => setSrc(`./variable/value/${variableKey}?t=${Date.now()}`))}>
        Set Image
      </button>
    </div>
  );
}

export function FileEditor({ variableKey, size }: { variableKey: string; size: number }) {
  const [displaySize] = useState(formatFileSize(size));
  return (
    <div className="variable-inputs">
      <span>{displaySize}</span>
      <button
        type="button"
        onClick={() =>
          setBinaryVariableValue(variableKey, () => {
            /* size updates on schema refresh */
          })
        }
      >
        Set File
      </button>
    </div>
  );
}
