interface ToggleProps {
  value: boolean;
  onChange: (value: boolean) => void;
}

export function Toggle({ value, onChange }: ToggleProps) {
  return (
    <button
      type="button"
      className="toggle"
      style={{
        flexDirection: value ? "row-reverse" : "row",
        backgroundColor: value ? "#196d6d" : "#0a2e3d",
      }}
      onClick={() => onChange(!value)}
      aria-pressed={value}
    >
      <span className="toggle-slider" />
    </button>
  );
}
