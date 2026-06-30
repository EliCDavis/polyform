import { Toggle } from "./Toggle";

interface GizmoToggleProps {
  value: boolean;
  onChange: (value: boolean) => void;
}

export function GizmoToggle({ value, onChange }: GizmoToggleProps) {
  return (
    <div style={{ display: "flex", flexDirection: "row", alignItems: "center", gap: 8 }}>
      <i className="fa-solid fa-eye" style={{ color: "#196d6d" }} />
      <span>Gizmo</span>
      <span style={{ flex: 1 }} />
      <Toggle value={value} onChange={onChange} />
    </div>
  );
}
