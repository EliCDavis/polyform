import { useEffect, useLayoutEffect, useRef, useState, type KeyboardEvent, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { ChevronRight } from "lucide-react";

export type MenuItem =
  | { label: string; onClick: () => void; icon?: ReactNode }
  | { label: string; items: MenuItem[]; icon?: ReactNode }
  | "separator";

interface DropdownMenuProps {
  items: MenuItem[];
}

type Placement = { kind: "below"; anchor: DOMRect } | { kind: "side"; anchor: DOMRect; parent: DOMRect };

interface OpenGroup {
  label: string;
  placement: Placement;
  viaKeyboard: boolean;
}

const MARGIN = 4;
const SUBMENU_DELAY_MS = 120;

function menuItems(panel: HTMLElement): HTMLButtonElement[] {
  return Array.from(panel.querySelectorAll<HTMLButtonElement>(".ctx-menu-item")).filter(
    (item) => item.closest(".ctx-menu") === panel,
  );
}

function focusSibling(panel: HTMLElement, step: number) {
  const items = menuItems(panel);
  if (items.length === 0) return;
  const index = items.indexOf(document.activeElement as HTMLButtonElement);
  items[(index + step + items.length) % items.length].focus();
}

function MenuPanel({
  items,
  placement,
  autoFocus,
  close,
  onBack,
}: {
  items: MenuItem[];
  placement: Placement;
  autoFocus: boolean;
  close: () => void;
  onBack?: () => void;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const timer = useRef<number | undefined>(undefined);
  const [openGroup, setOpenGroup] = useState<OpenGroup | null>(null);
  const [position, setPosition] = useState({ left: -9999, top: -9999 });
  const showIcons = items.some((item) => item !== "separator" && item.icon);

  useLayoutEffect(() => {
    const panel = ref.current!.getBoundingClientRect();
    const maxLeft = window.innerWidth - panel.width - MARGIN;
    const maxTop = window.innerHeight - panel.height - MARGIN;
    let left: number;
    let top: number;
    if (placement.kind === "below") {
      left = placement.anchor.right - panel.width;
      if (left < MARGIN) left = placement.anchor.left;
      top = placement.anchor.bottom + MARGIN;
      if (top > maxTop) top = placement.anchor.top - panel.height - MARGIN;
    } else {
      left = placement.parent.right - 2;
      if (left > maxLeft) left = placement.parent.left - panel.width + 2;
      top = placement.anchor.top - 5;
    }
    setPosition({
      left: Math.max(MARGIN, Math.min(left, maxLeft)),
      top: Math.max(MARGIN, Math.min(top, maxTop)),
    });
    if (autoFocus) menuItems(ref.current!)[0]?.focus({ preventScroll: true });
  }, [placement, autoFocus]);

  useEffect(() => () => window.clearTimeout(timer.current), []);

  const scheduleGroup = (next: OpenGroup | null) => {
    window.clearTimeout(timer.current);
    timer.current = window.setTimeout(() => setOpenGroup(next), SUBMENU_DELAY_MS);
  };

  const openGroupFrom = (label: string, row: HTMLElement, viaKeyboard: boolean): OpenGroup => ({
    label,
    viaKeyboard,
    placement: { kind: "side", anchor: row.getBoundingClientRect(), parent: ref.current!.getBoundingClientRect() },
  });

  const onKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
    if (e.key === "ArrowDown") focusSibling(ref.current!, 1);
    else if (e.key === "ArrowUp") focusSibling(ref.current!, -1);
    else if (e.key === "ArrowLeft" && onBack) onBack();
    else return;
    e.preventDefault();
    e.stopPropagation();
  };

  return (
    <div ref={ref} className="ctx-menu" role="menu" style={position} onKeyDown={onKeyDown}>
      {items.map((item, i) => {
        if (item === "separator") {
          return <div key={`separator-${i}`} className="ctx-menu-separator" role="separator" />;
        }
        const icon = showIcons && <span className="ctx-menu-icon">{item.icon}</span>;
        if ("items" in item) {
          const open = openGroup?.label === item.label;
          return (
            <div key={item.label} style={{ display: "contents" }} onMouseEnter={() => window.clearTimeout(timer.current)}>
              <button
                type="button"
                role="menuitem"
                aria-haspopup="menu"
                aria-expanded={open}
                className={`ctx-menu-item ${open ? "open" : ""}`}
                onMouseEnter={(e) => scheduleGroup(openGroupFrom(item.label, e.currentTarget, false))}
                onClick={(e) => {
                  window.clearTimeout(timer.current);
                  setOpenGroup(openGroupFrom(item.label, e.currentTarget, e.detail === 0));
                }}
                onKeyDown={(e) => {
                  if (e.key !== "ArrowRight") return;
                  e.preventDefault();
                  e.stopPropagation();
                  setOpenGroup(openGroupFrom(item.label, e.currentTarget, true));
                }}
              >
                {icon}
                <span>{item.label}</span>
                <ChevronRight className="ctx-menu-chevron" size={14} />
              </button>
              {open && (
                <MenuPanel
                  items={item.items}
                  placement={openGroup.placement}
                  autoFocus={openGroup.viaKeyboard}
                  close={close}
                  onBack={() => {
                    setOpenGroup(null);
                    menuItems(ref.current!).find((row) => row.getAttribute("aria-expanded") === "true")?.focus();
                  }}
                />
              )}
            </div>
          );
        }
        return (
          <button
            key={item.label}
            type="button"
            role="menuitem"
            className="ctx-menu-item"
            onMouseEnter={() => scheduleGroup(null)}
            onClick={() => {
              close();
              item.onClick();
            }}
          >
            {icon}
            <span>{item.label}</span>
          </button>
        );
      })}
    </div>
  );
}

export function DropdownMenu({ items }: DropdownMenuProps) {
  const trigger = useRef<HTMLButtonElement>(null);
  const [placement, setPlacement] = useState<Placement | null>(null);

  const close = (refocus = false) => {
    setPlacement(null);
    if (refocus) trigger.current?.focus();
  };

  useEffect(() => {
    if (!placement) return;
    const onPointerDown = (e: PointerEvent) => {
      const target = e.target as HTMLElement;
      if (target.closest(".ctx-menu") || trigger.current?.contains(target)) return;
      close();
    };
    const onKey = (e: globalThis.KeyboardEvent) => {
      if (e.key === "Escape") close(true);
    };
    const onLayoutChange = () => close();
    document.addEventListener("pointerdown", onPointerDown);
    document.addEventListener("keydown", onKey);
    window.addEventListener("resize", onLayoutChange);
    window.addEventListener("scroll", onLayoutChange, true);
    return () => {
      document.removeEventListener("pointerdown", onPointerDown);
      document.removeEventListener("keydown", onKey);
      window.removeEventListener("resize", onLayoutChange);
      window.removeEventListener("scroll", onLayoutChange, true);
    };
  }, [placement]);

  return (
    <>
      <button
        ref={trigger}
        type="button"
        className={`icon-button ${placement ? "active" : ""}`}
        aria-label="Menu"
        aria-haspopup="menu"
        aria-expanded={placement !== null}
        onClick={(e) =>
          setPlacement(placement ? null : { kind: "below", anchor: e.currentTarget.getBoundingClientRect() })
        }
      >
        <i className="fa-solid fa-ellipsis-vertical" />
      </button>
      {placement &&
        createPortal(
          <MenuPanel items={items} placement={placement} autoFocus close={() => close()} />,
          document.body,
        )}
    </>
  );
}
