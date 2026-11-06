import type { ReactNode } from "react";
import { NumberInput } from "@/components/NumberInput";
import styles from "./editors.module.css";

export const axes = ["x", "y", "z"] as const;
export const XYZ = ["X", "Y", "Z"];

export function AxisRow({
  labels,
  values,
  onCommit,
  step,
  integer,
}: {
  labels: string[];
  values: number[];
  onCommit: (index: number, value: number) => void;
  step?: string;
  integer?: boolean;
}) {
  return (
    <div className={styles.axisRow}>
      {labels.map((label, i) => (
        <label key={label} className={styles.axisField}>
          <span className={styles.axisLabel}>{label}</span>
          <NumberInput step={step} integer={integer} value={values[i]} onCommit={(n) => onCommit(i, n)} />
        </label>
      ))}
    </div>
  );
}

export function PartRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className={styles.partRow}>
      <span className={styles.partLabel}>{label}</span>
      {children}
    </div>
  );
}
