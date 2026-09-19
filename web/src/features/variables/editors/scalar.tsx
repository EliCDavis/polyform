import { useState } from "react";
import { setVariableValue } from "@/api/variables";
import { NumberInput } from "@/components/NumberInput";
import { ListEditor } from "./ListEditor";

export function BoolEditor({ variableKey, value }: { variableKey: string; value: boolean }) {
  const [checked, setChecked] = useState(!!value);
  return (
    <input
      type="checkbox"
      checked={checked}
      onChange={(e) => {
        setChecked(e.target.checked);
        void setVariableValue(variableKey, e.target.checked);
      }}
    />
  );
}

export function TextEditor({ variableKey, value }: { variableKey: string; value: string }) {
  const [text, setText] = useState(value);
  return (
    <input
      type="text"
      value={text}
      onChange={(e) => setText(e.target.value)}
      onBlur={() => void setVariableValue(variableKey, text)}
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

export function NumberListEditor({
  variableKey,
  value,
  step,
  parse,
}: {
  variableKey: string;
  value: number[];
  step: string;
  parse: (s: string) => number;
}) {
  return (
    <ListEditor<number>
      variableKey={variableKey}
      value={value}
      blank={(items, index) => items[index - 1] ?? 0}
      renderItem={(item, update) => (
        <NumberInput step={step} integer={parse !== parseFloat} value={item} onCommit={update} />
      )}
    />
  );
}

export function StringListEditor({ variableKey, value }: { variableKey: string; value: string[] }) {
  return (
    <ListEditor<string>
      variableKey={variableKey}
      value={value}
      blank={() => ""}
      renderItem={(item, update) => (
        <input
          type="text"
          defaultValue={item}
          onBlur={(e) => {
            if (e.target.value !== item) update(e.target.value);
          }}
        />
      )}
    />
  );
}
