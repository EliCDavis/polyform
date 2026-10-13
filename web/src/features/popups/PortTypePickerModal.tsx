import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";
import { X } from "lucide-react";
import { Modal, PopupButton } from "@/components/Modal";
import { formatPortTypeLabel } from "@/lib/portTypes";
import { usePortTypePickerStore } from "@/stores/portTypePickerStore";

export function PortTypePickerModal() {
  const request = usePortTypePickerStore((s) => s.request);
  const close = usePortTypePickerStore((s) => s.close);
  const confirm = usePortTypePickerStore((s) => s.confirm);
  const [selected, setSelected] = useState("");
  const [search, setSearch] = useState("");
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (request) {
      setSelected(request.current);
      setSearch("");
    }
  }, [request]);

  const matches = useMemo(() => {
    const terms = search.trim().toLowerCase().split(/\s+/).filter(Boolean);
    return (request?.options ?? []).filter((type) => {
      const label = formatPortTypeLabel(type).toLowerCase();
      return terms.every((term) => label.includes(term));
    });
  }, [request, search]);

  const active = matches.includes(selected) ? selected : matches[0];

  useEffect(() => {
    listRef.current
      ?.querySelector(".port-type-option.selected")
      ?.scrollIntoView({ block: "nearest" });
  }, [active]);

  if (!request) return null;

  const choose = () => {
    if (active !== undefined) {
      confirm(active);
    }
  };

  const onKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    const index = active === undefined ? -1 : matches.indexOf(active);
    if (e.key === "ArrowDown" && index < matches.length - 1) {
      e.preventDefault();
      setSelected(matches[index + 1]);
    } else if (e.key === "ArrowUp" && index > 0) {
      e.preventDefault();
      setSelected(matches[index - 1]);
    } else if (e.key === "Enter") {
      e.preventDefault();
      choose();
    } else if (e.key === "Escape") {
      e.preventDefault();
      close();
    }
  };

  return (
    <Modal
      title={request.title}
      open
      onClose={close}
      actions={
        <>
          <PopupButton onClick={close}>Cancel</PopupButton>
          <PopupButton variant="primary" onClick={choose} disabled={active === undefined}>
            Select
          </PopupButton>
        </>
      }
    >
      <div className="port-type-picker">
        <div className="node-catalog-search">
          <input
            type="text"
            placeholder="Search types..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={onKeyDown}
            autoFocus
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
          {matches.length === request.options.length
            ? `${matches.length} types`
            : `${matches.length} of ${request.options.length} types`}
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
              onClick={() => setSelected(type)}
              onDoubleClick={() => confirm(type)}
            >
              {formatPortTypeLabel(type)}
            </div>
          ))}
        </div>
      </div>
    </Modal>
  );
}
