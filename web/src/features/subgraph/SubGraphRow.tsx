import { useState } from "react";
import { DropdownMenu } from "@/components/DropdownMenu";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { RenameSubGraphModal } from "@/features/popups/RenameSubGraphModal";
import { DeleteSubGraphModal } from "@/features/popups/DeleteSubGraphModal";
import { useGraphTabStore } from "@/stores/graphTabStore";

export interface SubGraphPlacement {
  tabId: string;
  label: string;
  nodeId: string;
}

interface SubGraphRowProps {
  id: string;
  name: string;
  description?: string;
  group?: string;
  groups: string[];
  nodeCount: number;
  placements: SubGraphPlacement[];
}

const plural = (count: number, word: string) => `${count} ${word}${count === 1 ? "" : "s"}`;

function UsedInChip({ placements }: { placements: SubGraphPlacement[] }) {
  const focusNodeInTab = useGraphTabStore((s) => s.focusNodeInTab);
  const [next, setNext] = useState(0);
  const { tabId, label } = placements[0];
  return (
    <button
      type="button"
      className="used-in-chip"
      title={placements.length === 1 ? `Go to the placement in ${label}` : `Cycle through the ${placements.length} placements in ${label}`}
      onClick={() => {
        focusNodeInTab(tabId, label, placements[next % placements.length].nodeId);
        setNext(next + 1);
      }}
    >
      {label}
      {placements.length > 1 && <span className="used-in-count">×{placements.length}</span>}
    </button>
  );
}

export function SubGraphRow({ id, name, description, group, groups, nodeCount, placements }: SubGraphRowProps) {
  const editor = useEditorOptional();
  const openSubGraphTab = useGraphTabStore((s) => s.openSubGraphTab);
  const renameSubGraphTab = useGraphTabStore((s) => s.renameSubGraphTab);
  const closeTab = useGraphTabStore((s) => s.closeTab);
  const activeTabId = useGraphTabStore((s) => s.activeTabId);
  const isActive = activeTabId === id;
  const [renameOpen, setRenameOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const byScope = new Map<string, SubGraphPlacement[]>();
  for (const p of placements) byScope.set(p.tabId, [...(byScope.get(p.tabId) ?? []), p]);

  return (
    <div className={`subgraph-row ${isActive ? "active" : ""}`}>
      <div className="subgraph-row-header">
        <button
          type="button"
          className="subgraph-row-name"
          title={description || undefined}
          onClick={() => openSubGraphTab(id, name)}
        >
          {name}
        </button>
        <DropdownMenu
          items={[
            { label: "Edit", onClick: () => setRenameOpen(true) },
            { label: "Delete", onClick: () => setDeleteOpen(true) },
          ]}
        />
      </div>
      <div className="subgraph-row-body">
        <span className="subgraph-row-count">{plural(nodeCount, "node")}</span>
        <span className="subgraph-row-divider" />
        {byScope.size === 0 ? (
          <span className="used-in-none">not placed</span>
        ) : (
          [...byScope.values()].map((group) => <UsedInChip key={group[0].tabId} placements={group} />)
        )}
      </div>
      {editor && (
        <>
          <RenameSubGraphModal
            open={renameOpen}
            subGraphId={id}
            name={name}
            description={description}
            group={group}
            groups={groups}
            schemaManager={editor.schemaManager}
            nodeManager={editor.nodeManager}
            onRenamed={(nextName) => renameSubGraphTab(id, nextName)}
            onClose={() => setRenameOpen(false)}
          />
          <DeleteSubGraphModal
            open={deleteOpen}
            subGraphId={id}
            name={name}
            schemaManager={editor.schemaManager}
            nodeManager={editor.nodeManager}
            onDeleted={() => closeTab(id)}
            onClose={() => setDeleteOpen(false)}
          />
        </>
      )}
    </div>
  );
}
