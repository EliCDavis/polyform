import { useCallback, useRef, useState, type ReactNode } from "react";

interface ResizablePanelsProps {
  direction: "horizontal" | "vertical";
  first: ReactNode;
  second: ReactNode;
  initialFirstPercent?: number;
  firstCollapsedSize?: string;
  collapsed?: boolean;
}

export function ResizablePanels({
  direction,
  first,
  second,
  initialFirstPercent = 40,
  firstCollapsedSize,
  collapsed = false,
}: ResizablePanelsProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const firstRef = useRef<HTMLDivElement>(null);
  const percentRef = useRef(initialFirstPercent);
  const [dragging, setDragging] = useState(false);

  const onPointerDown = useCallback(
    (e: React.PointerEvent) => {
      e.preventDefault();
      const container = containerRef.current;
      const firstEl = firstRef.current;
      if (!container || !firstEl) return;

      const rect = container.getBoundingClientRect();
      const startX = e.clientX;
      const startY = e.clientY;
      const startSize =
        direction === "horizontal" ? firstEl.offsetWidth : firstEl.offsetHeight;
      const totalSize =
        direction === "horizontal" ? rect.width : rect.height;

      const onMove = (ev: PointerEvent) => {
        const dx = ev.clientX - startX;
        const dy = ev.clientY - startY;
        const delta = direction === "horizontal" ? dx : dy;
        const newSize = startSize + delta;
        percentRef.current = Math.min(Math.max((newSize / totalSize) * 100, 5), 95);
        if (direction === "horizontal") {
          firstEl.style.width = `${percentRef.current}%`;
        } else {
          firstEl.style.height = `${percentRef.current}%`;
        }
      };

      const onUp = () => {
        document.removeEventListener("pointermove", onMove);
        document.removeEventListener("pointerup", onUp);
        document.removeEventListener("pointercancel", onUp);
        document.body.style.cursor = "";
        setDragging(false);
      };

      document.body.style.cursor =
        direction === "horizontal" ? "col-resize" : "row-resize";
      document.addEventListener("pointermove", onMove);
      document.addEventListener("pointerup", onUp);
      document.addEventListener("pointercancel", onUp);
      setDragging(true);
    },
    [direction]
  );

  const isHorizontal = direction === "horizontal";
  const firstSize = collapsed && firstCollapsedSize ? firstCollapsedSize : `${percentRef.current}%`;

  return (
    <div
      ref={containerRef}
      style={{
        display: "flex",
        flexDirection: isHorizontal ? "row" : "column",
        flex: 1,
        height: "100%",
        minHeight: 0,
        width: "100%",
      }}
    >
      <div
        ref={firstRef}
        style={
          isHorizontal
            ? { width: firstSize, height: "100%", minWidth: 0, flexShrink: 0 }
            : { height: firstSize, width: "100%", minHeight: 0, flexShrink: 0 }
        }
      >
        {first}
      </div>
      <div
        className={`resizer ${dragging ? "dragging" : ""}`}
        data-direction={direction}
        data-collapsed={collapsed}
        onPointerDown={collapsed ? undefined : onPointerDown}
      />
      <div style={{ flex: 1, minHeight: 0, minWidth: 0, display: "flex", flexDirection: isHorizontal ? "row" : "column" }}>
        {second}
      </div>
    </div>
  );
}
