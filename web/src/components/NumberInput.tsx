import { useEffect, useRef, useState } from "react";

interface NumberInputProps {
  value: number;
  onCommit: (value: number) => void;
  integer?: boolean;
  step?: string;
  min?: number;
  max?: number;
  disabled?: boolean;
}

export function NumberInput({ value, onCommit, integer, step, min, max, disabled }: NumberInputProps) {
  const [text, setText] = useState(String(value));
  const [editing, setEditing] = useState(false);
  const lastCommitted = useRef(value);

  useEffect(() => {
    if (editing) return;
    setText(String(value));
    lastCommitted.current = value;
  }, [value, editing]);

  const parse = (s: string) => (integer ? parseInt(s, 10) : parseFloat(s));

  const tryCommit = (s: string) => {
    const n = parse(s);
    if (!Number.isFinite(n) || n === lastCommitted.current) return;
    lastCommitted.current = n;
    onCommit(n);
  };

  return (
    <input
      type="number"
      step={step ?? (integer ? "1" : "any")}
      min={min}
      max={max}
      disabled={disabled}
      value={text}
      onFocus={() => setEditing(true)}
      onChange={(e) => {
        setText(e.target.value);
        tryCommit(e.target.value);
      }}
      onBlur={() => {
        setEditing(false);
        const n = parse(text);
        setText(String(Number.isFinite(n) ? n : value));
      }}
      onKeyDown={(e) => {
        if (e.key === "Enter") (e.target as HTMLInputElement).blur();
      }}
    />
  );
}
