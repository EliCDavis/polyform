import { useMemo, useRef, useState, type ChangeEvent } from "react";
import { useSchema } from "@/api/hooks";
import { getApiErrorMessage, requestManager } from "@/api/client";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { useGraphTabStore } from "@/stores/graphTabStore";
import { useUiStore } from "@/stores/uiStore";
import type { RuntimeSubGraphDefinition } from "@/lib/schema";
import {
  applyImportedSubGraphs,
  formatImportSummary,
} from "@/lib/importSubGraphs";
import { NewSubGraphModal } from "@/features/popups/NewSubGraphModal";
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

  const groups = useMemo(() => {
    const byGroup = new Map<string, [string, RuntimeSubGraphDefinition][]>();
    for (const entry of subGraphs) {
      const group = entry[1].group ?? "";
      byGroup.set(group, [...(byGroup.get(group) ?? []), entry]);
    }
    return [...byGroup.entries()].sort(([a], [b]) => a.localeCompare(b));
  }, [subGraphs]);

  const instanceCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    const scopes = [graph?.nodes ?? {}, ...subGraphs.map(([, def]) => def.nodes ?? {})];
    for (const nodes of scopes) {
      for (const node of Object.values(nodes)) {
        if (node.subGraphId) {
          counts[node.subGraphId] = (counts[node.subGraphId] ?? 0) + 1;
        }
      }
    }
    return counts;
  }, [graph?.nodes, subGraphs]);

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

      const savedName = (payload as { data?: { name?: string } })?.data?.name;
      requestManager.importSubGraphs(
        payload,
        (result) => {
          applyImportedSubGraphs(editor.nodeManager, result);
          alert(formatImportSummary(result));
        },
        (err) => alert(getApiErrorMessage(err, "Failed to import sub-graphs")),
        savedName || file.name.replace(/\.json$/i, "")
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
        {groups.map(([group, entries]) => {
          const rows = entries.map(([id, def]) => (
            <SubGraphRow
              key={id}
              id={id}
              name={def.name || id}
              description={def.description}
              nodeCount={Object.keys(def.nodes ?? {}).length}
              instanceCount={instanceCounts[id] ?? 0}
            />
          ));
          if (!group) return rows;
          return (
            <details key={group} open>
              <summary>{group}</summary>
              {rows}
            </details>
          );
        })}
      </div>
      <NewSubGraphModal
        open={newOpen}
        onClose={() => setNewOpen(false)}
        nodeManager={editor.nodeManager}
        onCreated={(id, name) => openSubGraphTab(id, name)}
      />
    </>
  );
}
