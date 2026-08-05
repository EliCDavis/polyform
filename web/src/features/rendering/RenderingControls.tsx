import type { ReactNode } from "react";

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
  enabled: boolean;
  setEnabled: (enabled: boolean) => void;
  children: ReactNode;
}

export function RenderingGroup({
  name,
  description,
  enabled,
  setEnabled,
  children,
}: RenderingGroupProps) {
  return (
    <div className="rendering-group">
      <div className="variable-header">
        <span className="variable-name">{name}</span>
        <input
          type="checkbox"
          checked={enabled}
          onChange={(e) => setEnabled(e.target.checked)}
        />
      </div>
      {description && <div className="variable-description">{description}</div>}
      <div className={`rendering-group-content${enabled ? "" : " disabled"}`}>
        {children}
      </div>
    </div>
  );
}
