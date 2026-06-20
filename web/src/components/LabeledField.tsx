import type { ReactNode } from "react";

export function LabeledField({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="labeled-field">
      <label>{label}</label>
      {children}
    </div>
  );
}
