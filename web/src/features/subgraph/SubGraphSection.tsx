import { useMemo, useState } from "react";
import { useSchema } from "@/api/hooks";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { useGraphTabStore } from "@/stores/graphTabStore";
import type { RuntimeSubGraphDefinition } from "@/lib/schema";
import { NewSubGraphModal } from "@/features/popups/NewSubGraphModal";
import { SubGraphRow } from "./SubGraphRow";

export function SubGraphSection() {
  const { data: graph } = useSchema();
  const editor = useEditorOptional();
  const openSubGraphTab = useGraphTabStore((s) => s.openSubGraphTab);
  const [newOpen, setNewOpen] = useState(false);

  const subGraphs = useMemo((): [string, RuntimeSubGraphDefinition][] => {
    const entries = Object.entries(
      graph?.subGraphs ?? {}
    ) as [string, RuntimeSubGraphDefinition][];
    return entries.sort(([, a], [, b]) =>
      (a.name || "").localeCompare(b.name || "")
    );
  }, [graph?.subGraphs]);

  if (!editor) return null;

  return (
    <>
      <div className="sidebar-header">Sub-Graphs</div>
      <div className="sidebar-section-content">
        <button type="button" onClick={() => setNewOpen(true)}>
          New Sub-Graph
        </button>
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
    </>
  );
}
