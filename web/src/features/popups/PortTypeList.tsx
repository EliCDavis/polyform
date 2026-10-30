import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";
import { X } from "lucide-react";
import { formatPortTypeLabel } from "@/lib/portTypes";

interface PortTypeListProps {
  options: string[];
  selected: string;
  onSelect: (type: string) => void;
  onConfirm?: (type: string) => void;
  onCancel?: () => void;
  label?: (type: string) => string;
  autoFocus?: boolean;
}

export function PortTypeList({
  options,
  selected,
  onSelect,
  onConfirm,
  onCancel,
  label = formatPortTypeLabel,
  autoFocus,
}: PortTypeListProps) {
  const [search, setSearch] = useState("");
  const listRef = useRef<HTMLDivElement>(null);

  const matches = useMemo(() => {
    const terms = search.trim().toLowerCase().split(/\s+/).filter(Boolean);
    return options.filter((type) => {
      const text = `${label(type)} ${formatPortTypeLabel(type)}`.toLowerCase();
      return terms.every((term) => text.includes(term));
    });
  }, [options, search, label]);

  const active = matches.includes(selected) ? selected : matches[0];

  useEffect(() => {
    const next = active ?? "";
    if (next !== selected) onSelect(next);
  }, [active, selected, onSelect]);

  useEffect(() => {
    listRef.current
      ?.querySelector(".port-type-option.selected")
      ?.scrollIntoView({ block: "nearest" });
  }, [active]);

  const onKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    const index = active === undefined ? -1 : matches.indexOf(active);
    if (e.key === "ArrowDown" && index < matches.length - 1) {
      e.preventDefault();
      onSelect(matches[index + 1]);
    } else if (e.key === "ArrowUp" && index > 0) {
      e.preventDefault();
      onSelect(matches[index - 1]);
    } else if (e.key === "Enter" && onConfirm) {
      e.preventDefault();
      if (active !== undefined) onConfirm(active);
    } else if (e.key === "Escape" && onCancel) {
      e.preventDefault();
      onCancel();
    }
  };

  return (
    <div className="port-type-picker">
      <div className="node-catalog-search">
        <input
          type="text"
          placeholder="Search types..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          onKeyDown={onKeyDown}
          autoFocus={autoFocus}
        />
        {search.length > 0 && (
          <button
            type="button"
            className="icon-button node-catalog-search-clear"
            onClick={() => setSearch("")}
            aria-label="Clear search"
          >
            <X size={14} />
          </button>
        )}
      </div>
      <div className="node-catalog-result-count">
        {matches.length === options.length
          ? `${matches.length} types`
          : `${matches.length} of ${options.length} types`}
      </div>
      <div className="port-type-list" role="listbox" ref={listRef}>
        {matches.length === 0 && (
          <div className="variable-description">No types match your search.</div>
        )}
        {matches.map((type) => (
          <div
            key={type}
            role="option"
            aria-selected={type === active}
            className={type === active ? "port-type-option selected" : "port-type-option"}
            onClick={() => onSelect(type)}
            onDoubleClick={() => onConfirm?.(type)}
          >
            {label(type)}
          </div>
        ))}
      </div>
    </div>
  );
}
