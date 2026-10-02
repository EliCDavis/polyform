import { setVariableValue } from "@/api/variables";
import { NumberInput } from "@/components/NumberInput";
import { useLiveValue } from "@/components/useLiveValue";
import { ListEditor } from "./ListEditor";

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
