import { useMemo, useRef, useState, type ChangeEvent } from "react";
import { useSchema } from "@/api/hooks";
import { getApiErrorMessage, requestManager } from "@/api/client";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { useGraphTabStore } from "@/stores/graphTabStore";
import { useUiStore } from "@/stores/uiStore";
import type { RuntimeSubGraphDefinition } from "@/lib/schema";
import {
  applyImportedSubGraphs,
  formatImportRemapSummary,
} from "@/lib/importSubGraphs";
import { NewSubGraphModal } from "@/features/popups/NewSubGraphModal";
import { ConvertToSubGraphModal } from "@/features/popups/ConvertToSubGraphModal";
import { SubGraphRow } from "./SubGraphRow";

export function SubGraphSection() {
  const { data: graph } = useSchema();
  const editor = useEditorOptional();
  const openSubGraphTab = useGraphTabStore((s) => s.openSubGraphTab);
  const hideImportSubGraphs = useUiStore((s) => s.hideImportSubGraphs);
  const [newOpen, setNewOpen] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const subGraphs = useMemo((): [string, RuntimeSubGraphDefinition][] => {
    const entries = Object.entries(
      graph?.subGraphs ?? {}
    ) as [string, RuntimeSubGraphDefinition][];
    return entries.sort(([, a], [, b]) =>
      (a.name || "").localeCompare(b.name || "")
    );
  }, [graph?.subGraphs]);

  if (!editor) return null;

  const importFromFile = () => {
    fileInputRef.current?.click();
  };

  const onFileSelected = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (ev) => {
      let payload: unknown;
      try {
        payload = JSON.parse(ev.target?.result as string);
      } catch {
        alert("Selected file is not valid JSON");
        return;
      }

      requestManager.importSubGraphs(
        payload,
        (result) => {
          applyImportedSubGraphs(editor.nodeManager, result);
          const remaps = formatImportRemapSummary(result);
          if (remaps) {
            alert(`Imported ${result.imported.length} sub-graph(s).\n\nRenamed due to conflicts:\n${remaps}`);
          } else if (result.imported.length === 0) {
            alert("No sub-graph definitions found in the file.");
          }
        },
        (err) => alert(getApiErrorMessage(err, "Failed to import sub-graphs"))
      );
    };
    reader.readAsText(file);
  };

  return (
    <>
      <div className="sidebar-header">Sub-Graphs</div>
      <div className="sidebar-section-content">
        <div style={{ display: "flex", flexDirection: "row", gap: 8 }}>
          <button type="button" style={{ flex: 1 }} onClick={() => setNewOpen(true)}>
            New Sub-Graph
          </button>
          {!hideImportSubGraphs && (
            <button type="button" style={{ flex: 1 }} onClick={importFromFile}>
              Import…
            </button>
          )}
        </div>
        {!hideImportSubGraphs && (
          <input
            ref={fileInputRef}
            type="file"
            accept="application/json,.json"
            style={{ display: "none" }}
            onChange={onFileSelected}
          />
        )}
        {subGraphs.map(([id, def]) => (
          <SubGraphRow
            key={id}
            id={id}
            name={def.name || id}
            description={def.description}
          />
        ))}
      </div>
      <NewSubGraphModal
        open={newOpen}
        onClose={() => setNewOpen(false)}
        nodeManager={editor.nodeManager}
        onCreated={(id, name) => openSubGraphTab(id, name)}
      />
      <ConvertToSubGraphModal
        nodeManager={editor.nodeManager}
        schemaManager={editor.schemaManager}
      />
    </>
  );
}
