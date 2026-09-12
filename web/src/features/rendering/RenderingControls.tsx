import type { ReactNode } from "react";
import styles from "./RenderingControls.module.css";

interface RenderingOptionProps {
  name: string;
  description?: string;
  value: number;
  setValue: (number) => void;
}

export function RenderingOption({
  name,
  description,
  value,
  setValue,
}: RenderingOptionProps) {
  return (
    <div className="variable-row">
      <div className="variable-header">
        <span className="variable-name">{name}</span>
      </div>
      {description && <div className="variable-description">{description}</div>}
      <input
        value={value}
        onChange={(e) => setValue(parseFloat(e.target.value))}
      ></input>
    </div>
  );
}

interface RenderingToggleOptionProps {
  name: string;
  description?: string;
  value: boolean;
  setValue: (value: boolean) => void;
}

export function RenderingToggleOption({
  name,
  description,
  value,
  setValue,
}: RenderingToggleOptionProps) {
  return (
    <div className="variable-row">
      <div className="variable-header">
        <span className="variable-name">{name}</span>
        <input
          type="checkbox"
          checked={value}
          onChange={(e) => setValue(e.target.checked)}
        />
      </div>
      {description && <div className="variable-description">{description}</div>}
    </div>
  );
}

interface RenderingSliderOptionProps {
  name: string;
  description?: string;
  value: number;
  min: number;
  max: number;
  step?: number;
  unit?: string;
  setValue: (value: number) => void;
}

export function RenderingSliderOption({
  name,
  description,
  value,
  min,
  max,
  step = 1,
  unit = "",
  setValue,
}: RenderingSliderOptionProps) {
  return (
    <div className="variable-row">
      <div className="variable-header">
        <span className="variable-name">{name}</span>
        <span className={styles.sliderValue}>
          {value}
          {unit}
        </span>
      </div>
      {description && <div className="variable-description">{description}</div>}
      <input
        className={styles.slider}
        type="range"
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => setValue(parseFloat(e.target.value))}
      />
    </div>
  );
}

interface RenderingSelectOptionProps<T extends string> {
  name: string;
  description?: string;
  value: T;
  options: Array<{ label: string; value: T }>;
  setValue: (value: T) => void;
}

export function RenderingSelectOption<T extends string>({
  name,
  description,
  value,
  options,
  setValue,
}: RenderingSelectOptionProps<T>) {
  return (
    <div className="variable-row">
      <div className="variable-header">
        <span className="variable-name">{name}</span>
      </div>
      {description && <div className="variable-description">{description}</div>}
      <select value={value} onChange={(e) => setValue(e.target.value as T)}>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </div>
  );
}

interface RenderingColorOptionProps {
  name: string;
  description?: string;
  value: string;
  setValue: (color: string) => void;
}

export function RenderingColorOption({
  name,
  description,
  value,
  setValue,
}: RenderingColorOptionProps) {
  return (
    <div className="variable-row">
      <div className="variable-header">
        <span className="variable-name">{name}</span>
      </div>
      {description && <div className="variable-description">{description}</div>}
      <div style={{ display: "flex", flexDirection: "row", gap: 16, alignItems: "center" }}>
        <input
          type="color"
          value={value}
          style={{ minHeight: 25, width: 25, maxWidth: 25, padding: 0, cursor: "pointer" }}
          onChange={(e) => setValue(e.target.value)}
        />
        <span>{value}</span>
      </div>
    </div>
  );
}

interface RenderingGroupProps {
  name: string;
  description?: string;
  /** Omit both to get a plain heading rather than a switchable feature. */
  enabled?: boolean;
  setEnabled?: (enabled: boolean) => void;
  children: ReactNode;
}

export function RenderingGroup({
  name,
  description,
  enabled,
  setEnabled,
  children,
}: RenderingGroupProps) {
  const switchable = setEnabled !== undefined;
  const active = !switchable || enabled === true;
  return (
    <div className={styles.group}>
      <div className="variable-header">
        <span className={styles.groupName}>{name}</span>
        {switchable && (
          <input
            type="checkbox"
            checked={enabled === true}
            onChange={(e) => setEnabled(e.target.checked)}
          />
        )}
      </div>
      {description && <div className="variable-description">{description}</div>}
      <div className={`${styles.groupContent} ${active ? "" : styles.disabled}`}>
        {children}
      </div>
    </div>
  );
}
