import { useState, useRef, useEffect } from "react";

export type MenuItem =
  | { label: string; onClick: () => void }
  | { label: string; items: MenuItem[] };

interface DropdownMenuProps {
  items: MenuItem[];
}

function MenuItems({ items, depth, close }: { items: MenuItem[]; depth: number; close: () => void }) {
  const [openGroup, setOpenGroup] = useState<string | null>(null);
  const indent = { paddingLeft: 12 + depth * 14 };

  return (
    <>
      {items.map((item) =>
        "items" in item ? (
          <div key={item.label}>
            <button
              type="button"
              className={`dropdown-item dropdown-group ${openGroup === item.label ? "open" : ""}`}
              style={indent}
              aria-expanded={openGroup === item.label}
              onClick={() => setOpenGroup(openGroup === item.label ? null : item.label)}
            >
              <span className="dropdown-group-chevron">▶</span>
              {item.label}
            </button>
            {openGroup === item.label && <MenuItems items={item.items} depth={depth + 1} close={close} />}
          </div>
        ) : (
          <button
            key={item.label}
            type="button"
            className="dropdown-item"
            style={indent}
            onClick={() => {
              close();
              item.onClick();
            }}
          >
            {item.label}
          </button>
        )
      )}
    </>
  );
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
      {open && (
        <div className="dropdown-content">
          <MenuItems items={items} depth={0} close={() => setOpen(false)} />
        </div>
      )}
    </div>
  );
}
