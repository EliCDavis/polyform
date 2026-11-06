import { useEffect, useRef } from "react";
import { setVariableValue } from "@/api/variables";
import { NumberInput } from "@/components/NumberInput";
import { useLiveValue } from "@/components/useLiveValue";
import { ListEditor } from "./ListEditor";
import styles from "./editors.module.css";

export function BoolEditor({ variableKey, value }: { variableKey: string; value: boolean }) {
  const [checked, , commit] = useLiveValue(!!value, (v) => setVariableValue(variableKey, v));
  return <input type="checkbox" checked={checked} onChange={(e) => commit(e.target.checked)} />;
}

export function TextEditor({ variableKey, value }: { variableKey: string; value: string }) {
  const [text, setText, commit] = useLiveValue(value, (v) => setVariableValue(variableKey, v));
  return (
    <input
      type="text"
      value={text}
      onChange={(e) => setText(e.target.value)}
      onBlur={() => commit(text)}
    />
  );
}

export function NumberEditor({
  variableKey,
  value,
  step,
  parse,
}: {
  variableKey: string;
  value: number;
  step: string;
  parse: (s: string) => number;
}) {
  return (
    <NumberInput
      step={step}
      integer={parse !== parseFloat}
      value={value ?? 0}
      onCommit={(n) => void setVariableValue(variableKey, n)}
    />
  );
}

export function NormEditor({ variableKey, value, min }: { variableKey: string; value: number; min: number }) {
  const [shown, set, commit] = useLiveValue(value ?? 0, (v) => setVariableValue(variableKey, v));
  const slider = useRef<HTMLInputElement>(null);
  const commitRef = useRef(commit);
  commitRef.current = commit;

  // React's onChange is the input event; the native change event is the one fired on release.
  useEffect(() => {
    const el = slider.current;
    if (!el) return;
    const release = () => commitRef.current(parseFloat(el.value));
    el.addEventListener("change", release);
    return () => el.removeEventListener("change", release);
  }, []);

  return (
    <div className={styles.normRow}>
      <input
        ref={slider}
        className={styles.normSlider}
        type="range"
        min={min}
        max={1}
        step={0.01}
        value={shown}
        onChange={(e) => set(parseFloat(e.target.value))}
      />
      <div className={styles.normValue}>
        <NumberInput step="0.01" min={min} max={1} value={shown} onCommit={commit} />
      </div>
    </div>
  );
}

export function NumberListEditor({
  variableKey,
  value,
  open,
  step,
  parse,
}: {
  variableKey: string;
  value: number[];
  open: boolean;
  step: string;
  parse: (s: string) => number;
}) {
  return (
    <ListEditor<number>
      variableKey={variableKey}
      value={value}
      open={open}
      blank={(items, index) => items[index - 1] ?? 0}
      renderItem={(item, update) => (
        <NumberInput step={step} integer={parse !== parseFloat} value={item} onCommit={update} />
      )}
    />
  );
}

export function StringListEditor({ variableKey, value, open }: { variableKey: string; value: string[]; open: boolean }) {
  return (
    <ListEditor<string>
      variableKey={variableKey}
      value={value}
      open={open}
      blank={() => ""}
      renderItem={(item, update) => <StringListItem value={item} onCommit={update} />}
    />
  );
}

function StringListItem({ value, onCommit }: { value: string; onCommit: (next: string) => void }) {
  const [text, setText, commit] = useLiveValue(value, async (next) => onCommit(next));
  return (
    <input
      type="text"
      value={text}
      onChange={(e) => setText(e.target.value)}
      onBlur={() => {
        if (text !== value) commit(text);
      }}
    />
  );
}
