import { useState, useRef, useEffect } from "react";

interface DropdownMenuProps {
  items: Array<{ label: string; onClick: () => void }>;
}

export function DropdownMenu({ items }: DropdownMenuProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  return (
    <div className={`dropdown-menu ${open ? "open" : ""}`} ref={ref}>
      <button
        type="button"
        className="icon-button"
        onClick={() => setOpen(!open)}
        aria-label="Menu"
      >
        <i className="fa-solid fa-ellipsis-vertical" />
      </button>
      <div className="dropdown-content">
        {items.map((item) => (
          <button
            key={item.label}
            type="button"
            className="dropdown-item"
            onClick={() => {
              setOpen(false);
              item.onClick();
            }}
          >
            {item.label}
          </button>
        ))}
      </div>
    </div>
  );
}
