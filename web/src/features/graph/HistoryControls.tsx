import { requestManager } from "@/api/client";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { useCallback, useEffect, useState } from "react";
import type { GraphHistory } from "@/lib/schema";

export function HistoryControls() {
  const editor = useEditorOptional();
  const [history, setHistory] = useState<GraphHistory>({ undo: [], redo: [] });

  const refresh = useCallback(() => {
    requestManager.getHistory((state) => setHistory(state));
  }, []);

  useEffect(refresh, [refresh]);

  const walk = useCallback(
    (direction: "undo" | "redo") => {
      const step = direction === "undo" ? requestManager.undo : requestManager.redo;
      step.call(requestManager, (state) => {
        setHistory(state);
        editor?.schemaManager.refreshSchema(direction);
      });
    },
    [editor]
  );

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (!e.ctrlKey && !e.metaKey) {
        return;
      }
      // Typing in a field is the browser's undo, not the graph's.
      const target = e.target as HTMLElement | null;
      if (target?.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(target?.tagName ?? "")) {
        return;
      }

      const key = e.key.toLowerCase();
      if (key === "z" && !e.shiftKey) {
        e.preventDefault();
        walk("undo");
      } else if (key === "y" || (key === "z" && e.shiftKey)) {
        e.preventDefault();
        walk("redo");
      }
    };

    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [walk]);

  // Something else changed the graph, so what is on the stack has moved on.
  useEffect(() => {
    const subscription = editor?.schemaManager.instance$().subscribe(() => refresh());
    return () => subscription?.unsubscribe();
  }, [editor, refresh]);

  const label = (verb: string, steps: Array<string>) =>
    steps.length === 0 ? `Nothing to ${verb}` : `${verb}: ${steps[0]}`;

  return (
    <div id="history-controls-section">
      <div className="sidebar-header">History</div>
      <div className="sidebar-section-content" style={{ flexDirection: "row" }}>
        <button
          type="button"
          className="sidebar-button"
          style={{ flex: 1 }}
          disabled={history.undo.length === 0}
          title={`${label("Undo", history.undo)} (Ctrl+Z)`}
          onClick={() => walk("undo")}
        >
          Undo
        </button>
        <button
          type="button"
          className="sidebar-button"
          style={{ flex: 1 }}
          disabled={history.redo.length === 0}
          title={`${label("Redo", history.redo)} (Ctrl+Y)`}
          onClick={() => walk("redo")}
        >
          Redo
        </button>
      </div>
    </div>
  );
}
