import { setVariableValue } from "@/api/variables";
import { useLiveValue } from "@/components/useLiveValue";

export function EnumEditor({
  variableKey,
  value,
  values,
}: {
  variableKey: string;
  value: string;
  values: string[];
}) {
  const [selected, , commit] = useLiveValue(value, (v) => setVariableValue(variableKey, v));
  return (
    <select value={selected} onChange={(e) => commit(e.target.value)}>
      {values.map((name) => (
        <option key={name} value={name}>
          {name}
        </option>
      ))}
    </select>
  );
}
