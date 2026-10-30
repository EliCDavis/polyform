import { Fragment, useEffect, useRef, useState, type ReactNode } from "react";
import { setVariableValue } from "@/api/variables";
import styles from "./editors.module.css";

export interface ListHandle<T> {
  items: () => T[];
  commit: (next: T[]) => void;
}

export function ListToggle({ label, open, onToggle }: { label: string; open: boolean; onToggle: () => void }) {
  return (
    <button
      type="button"
      className={`${styles.listToggle} ${open ? styles.listToggleOpen : ""}`}
      aria-expanded={open}
      onClick={onToggle}
    >
      <span className={styles.listToggleIcon}>▶</span>
      {label}
    </button>
  );
}

export function RemoveButton({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <button type="button" className={styles.remove} title={label} aria-label={label} onClick={onClick}>
      −
    </button>
  );
}

export function ListItem({
  index,
  focused,
  onFocus,
  onRemove,
  children,
}: {
  index: ReactNode;
  focused?: boolean;
  onFocus?: (focused: boolean) => void;
  onRemove: () => void;
  children: ReactNode;
}) {
  return (
    <div
      className={`${styles.listItem} ${focused ? styles.pointRowFocused : ""}`}
      onPointerEnter={() => onFocus?.(true)}
      onPointerLeave={() => onFocus?.(false)}
    >
      <span className={styles.listIndex}>{index}</span>
      <div className={styles.listBody}>{children}</div>
      <RemoveButton label={`Remove ${index}`} onClick={onRemove} />
    </div>
  );
}

// Owns the list, its writes, and the insert/delete controls; the caller only
// says how one item is drawn.
export function ListEditor<T>({
  variableKey,
  value,
  blank,
  renderItem,
  handle,
  onItems,
  onFocus,
  onInsertPreview,
  footer,
}: {
  variableKey: string;
  value: T[];
  blank: (items: T[], index: number) => T;
  renderItem: (item: T, update: (next: T) => void) => React.ReactNode;
  // Lets a gizmo owned by the caller read and write the list. The list's
  // state lives here, so the caller is told about every change explicitly
  // rather than relying on a re-render it will not get.
  handle?: React.MutableRefObject<ListHandle<T> | null>;
  onItems?: (items: T[]) => void;
  onFocus?: (index: number | null) => void;
  onInsertPreview?: (index: number | null) => void;
  footer?: React.ReactNode;
}) {
  const [items, setItems] = useState<T[]>(value ?? []);
  const [open, setOpen] = useState(false);
  const [focused, setFocusedState] = useState<number | null>(null);
  const itemsRef = useRef(items);
  itemsRef.current = items;
  const pendingWrites = useRef(0);

  const setFocused = (index: number | null) => {
    setFocusedState(index);
    onFocus?.(index);
  };

  const previewInsert = (index: number | null) => onInsertPreview?.(index);

  useEffect(() => {
    if (pendingWrites.current > 0) return;
    setItems(value ?? []);
    onItems?.(value ?? []);
  }, [value]);

  const commit = (next: T[]) => {
    itemsRef.current = next;
    setItems(next);
    onItems?.(next);
    pendingWrites.current += 1;
    void setVariableValue(variableKey, next).finally(() => {
      pendingWrites.current -= 1;
    });
  };

  if (handle) {
    handle.current = { items: () => itemsRef.current, commit };
  }

  const insertItem = (index: number) => {
    const next = [...items];
    next.splice(index, 0, blank(items, index));
    previewInsert(null);
    setFocused(index);
    commit(next);
  };

  const insertControl = (index: number, label: string) => (
    <button
      type="button"
      className={styles.insert}
      title={label}
      onClick={() => insertItem(index)}
      onPointerEnter={() => previewInsert(index)}
      onPointerLeave={() => previewInsert(null)}
    >
      <span className={styles.insertRule} />
      <span className={styles.insertIcon}>+</span>
      <span className={styles.insertRule} />
    </button>
  );

  return (
    <div className="variable-inputs">
      <ListToggle label={`${items.length} items`} open={open} onToggle={() => setOpen(!open)} />
      {open && (
        <>
          {items.map((item, i) => (
            <Fragment key={i}>
              {insertControl(i, `Insert before ${i}`)}
              <ListItem
                index={i}
                focused={focused === i}
                onFocus={(on) => setFocused(on ? i : null)}
                onRemove={() => {
                  setFocused(null);
                  commit(items.filter((_, j) => j !== i));
                }}
              >
                {renderItem(item, (next) => commit(items.map((it, j) => (j === i ? next : it))))}
              </ListItem>
            </Fragment>
          ))}
          {insertControl(items.length, "Add at the end")}
        </>
      )}
      {footer}
    </div>
  );
}
