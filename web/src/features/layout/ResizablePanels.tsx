import { useCallback, useRef, type ReactNode } from "react";

interface ResizablePanelsProps {
  direction: "horizontal" | "vertical";
  first: ReactNode;
  second: ReactNode;
  initialFirstPercent?: number;
}

export function ResizablePanels({
  direction,
  first,
  second,
  initialFirstPercent = 40,
}: ResizablePanelsProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const firstRef = useRef<HTMLDivElement>(null);
  const percentRef = useRef(initialFirstPercent);

  const onMouseDown = useCallback(
    (e: React.MouseEvent) => {
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

      const onMove = (ev: MouseEvent) => {
        const dx = ev.clientX - startX;
        const dy = ev.clientY - startY;
        const delta = direction === "horizontal" ? dx : dy;
        const newSize = startSize + delta;
        percentRef.current = (newSize / totalSize) * 100;
        if (direction === "horizontal") {
          firstEl.style.width = `${percentRef.current}%`;
        } else {
          firstEl.style.height = `${percentRef.current}%`;
        }
      };

      const onUp = () => {
        document.removeEventListener("mousemove", onMove);
        document.removeEventListener("mouseup", onUp);
        document.body.style.cursor = "";
      };

      document.body.style.cursor =
        direction === "horizontal" ? "col-resize" : "row-resize";
      document.addEventListener("mousemove", onMove);
      document.addEventListener("mouseup", onUp);
    },
    [direction]
  );

  const isHorizontal = direction === "horizontal";

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
            ? { width: `${initialFirstPercent}%`, height: "100%", minWidth: 0 }
            : { height: `${initialFirstPercent}%`, width: "100%", minHeight: 0 }
        }
      >
        {first}
      </div>
      <div
        className="resizer"
        data-direction={direction}
        onMouseDown={onMouseDown}
      />
      <div style={{ flex: 1, minHeight: 0, minWidth: 0, display: "flex", flexDirection: isHorizontal ? "row" : "column" }}>
        {second}
      </div>
    </div>
  );
}
