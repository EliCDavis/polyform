import { useState } from "react";
import { portTypeColor } from "@elicdavis/node-flow";
import { DropdownMenu } from "@/components/DropdownMenu";
import { formatPortTypeLabel } from "@/lib/portTypes";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { RenameSubGraphModal } from "@/features/popups/RenameSubGraphModal";
import { DeleteSubGraphModal } from "@/features/popups/DeleteSubGraphModal";
import { useGraphTabStore } from "@/stores/graphTabStore";

export interface SubGraphPlacement {
  tabId: string;
  label: string;
  nodeId: string;
}

export interface SubGraphPort {
  name: string;
  type: string;
}

interface SubGraphRowProps {
  id: string;
  name: string;
  description?: string;
  group?: string;
  groups: string[];
  nodeCount: number;
  placements: SubGraphPlacement[];
  inputs: SubGraphPort[];
  outputs: SubGraphPort[];
}

const plural = (count: number, word: string) => `${count} ${word}${count === 1 ? "" : "s"}`;

function UsedInLink({ placements }: { placements: SubGraphPlacement[] }) {
  const focusNodeInTab = useGraphTabStore((s) => s.focusNodeInTab);
  const [next, setNext] = useState(0);
  const { tabId, label } = placements[0];
  return (
    <button
      type="button"
      className="used-in-link"
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

function PortList({ ports, side }: { ports: SubGraphPort[]; side: "input" | "output" }) {
  return (
    <div className={`subgraph-port-list ${side}`}>
      {ports.map((port) => (
        <span key={port.name} className="subgraph-port" title={formatPortTypeLabel(port.type)}>
          <span className="subgraph-port-dot" style={{ backgroundColor: portTypeColor(port.type) }} />
          <span className="subgraph-port-name">{port.name}</span>
        </span>
      ))}
    </div>
  );
}

export function SubGraphRow({ id, name, description, group, groups, nodeCount, placements, inputs, outputs }: SubGraphRowProps) {
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
          title={description || name}
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
      <div className="subgraph-row-meta">
        {plural(nodeCount, "node")} ·{" "}
        {byScope.size === 0 ? (
          <span className="used-in-none">not placed</span>
        ) : (
          <>
            in{" "}
            {[...byScope.values()].map((group, i) => (
              <span key={group[0].tabId}>
                {i > 0 && ", "}
                <UsedInLink placements={group} />
              </span>
            ))}
          </>
        )}
      </div>
      {(inputs.length > 0 || outputs.length > 0) && (
        <div className="subgraph-row-ports">
          <PortList ports={inputs} side="input" />
          <PortList ports={outputs} side="output" />
        </div>
      )}
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
            placements={placements}
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
