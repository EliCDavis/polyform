import { useEffect, useRef, useState } from "react";
import { setVariableValue } from "@/api/variables";
import { NumberInput } from "@/components/NumberInput";
import { ListEditor } from "./ListEditor";
import styles from "./editors.module.css";

export function ColorEditor({ variableKey, value }: { variableKey: string; value: string }) {
  const [color, setColor] = useState(value);
  return (
    <div style={{ display: "flex", flexDirection: "row", gap: 16, alignItems: "center" }}>
      <input
        type="color"
        value={color}
        style={{ minHeight: 25, width: 25, maxWidth: 25, padding: 0, cursor: "pointer" }}
        onChange={(e) => {
          setColor(e.target.value);
          void setVariableValue(variableKey, e.target.value);
        }}
      />
      <span>{color}</span>
    </div>
  );
}

// Alpha is dropped: <input type="color"> only speaks #rrggbb.
function ColorSwatch({ value, onChange }: { value: string; onChange: (hex: string) => void }) {
  return (
    <input
      type="color"
      value={(value ?? "#000000").slice(0, 7)}
      style={{ minHeight: 25, width: 25, maxWidth: 25, padding: 0, cursor: "pointer" }}
      onChange={(e) => onChange(e.target.value)}
    />
  );
}

export function ColorListEditor({ variableKey, value }: { variableKey: string; value: string[] }) {
  return (
    <ListEditor<string>
      variableKey={variableKey}
      value={value}
      blank={(items, index) => items[index - 1] ?? "#ffffff"}
      renderItem={(item, update) => (
        <div style={{ display: "flex", flexDirection: "row", gap: 16, alignItems: "center" }}>
          <ColorSwatch value={item} onChange={update} />
          <span>{item}</span>
        </div>
      )}
    />
  );
}

interface GradientKey {
  time: number;
  value: string;
}

export interface GradientValue {
  keys: GradientKey[];
}

function gradientCSS(keys: GradientKey[]): string {
  if (keys.length === 0) return "transparent";
  if (keys.length === 1) return keys[0].value;
  return `linear-gradient(to right, ${keys.map((k) => `${k.value} ${k.time * 100}%`).join(", ")})`;
}

// The backend renormalizes key times to span 0..1 on every write, so the
// first and last stops are pinned and only the ones between them move.
export function GradientEditor({ variableKey, value }: { variableKey: string; value: GradientValue }) {
  const [keys, setKeys] = useState<GradientKey[]>(value?.keys ?? []);
  const pendingWrites = useRef(0);

  useEffect(() => {
    if (pendingWrites.current > 0) return;
    setKeys(value?.keys ?? []);
  }, [value]);

  const commit = (next: GradientKey[]) => {
    const sorted = [...next].sort((a, b) => a.time - b.time);
    setKeys(sorted);
    pendingWrites.current += 1;
    void setVariableValue(variableKey, { keys: sorted }).finally(() => {
      pendingWrites.current -= 1;
    });
  };

  const insertAfter = (index: number) => {
    const left = keys[index];
    const right = keys[index + 1] ?? left;
    const next = [...keys];
    next.splice(index + 1, 0, { time: (left.time + right.time) / 2, value: left.value });
    commit(next);
  };

  const pinned = (i: number) => keys.length > 1 && (i === 0 || i === keys.length - 1);

  return (
    <div className="variable-inputs">
      <div style={{ height: 18, borderRadius: 3, background: gradientCSS(keys) }} />
      {keys.map((key, i) => (
        <div key={i} className={styles.pointRow}>
          <div style={{ display: "flex", flexDirection: "row", gap: 8, alignItems: "center" }}>
            <ColorSwatch value={key.value} onChange={(hex) => commit(keys.map((k, j) => (j === i ? { ...k, value: hex } : k)))} />
            <NumberInput
              step="0.01"
              min={0}
              max={1}
              value={key.time}
              disabled={pinned(i)}
              onCommit={(t) => commit(keys.map((k, j) => (j === i ? { ...k, time: t } : k)))}
            />
            <button type="button" title="Insert a stop after this one" onClick={() => insertAfter(i)}>
              +
            </button>
            <button type="button" onClick={() => commit(keys.filter((_, j) => j !== i))}>
              Delete
            </button>
          </div>
        </div>
      ))}
      {keys.length === 0 && (
        <button type="button" onClick={() => commit([{ time: 0, value: "#000000" }, { time: 1, value: "#ffffff" }])}>
          Add
        </button>
      )}
    </div>
  );
}
