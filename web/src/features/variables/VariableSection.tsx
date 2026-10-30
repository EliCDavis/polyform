import { useState } from "react";
import { useSchema } from "@/api/hooks";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { useUiStore } from "@/stores/uiStore";
import type { ThreeApp } from "@/lib/three_app";
import type { VariableGroup } from "@/lib/schema";
import { VariableRow } from "./VariableRow";
import { NewVariableModal } from "@/features/popups/NewVariableModal";
import { ProfileBar } from "./ProfileBar";

const sorted = <T,>(entries: Record<string, T> | undefined): [string, T][] =>
  Object.entries(entries ?? {}).sort(([a], [b]) => a.localeCompare(b));

// A variable's key is its full path, "Group/Sub/Name"; the schema nests it by group.
function VariableGroupRows({ group, path, threeApp }: { group: VariableGroup; path: string; threeApp?: ThreeApp }) {
  return (
    <>
      {sorted(group.variables).map(([name, variable]) => (
        <VariableRow key={name} variableKey={path + name} label={name} variable={variable} threeApp={threeApp} />
      ))}
      {sorted(group.subgroups).map(([name, subgroup]) => (
        <details key={name} className="sidebar-group" open>
          <summary>{name}</summary>
          <div className="sidebar-group-body">
            <VariableGroupRows group={subgroup} path={`${path}${name}/`} threeApp={threeApp} />
          </div>
        </details>
      ))}
    </>
  );
}

function isEmpty(group: VariableGroup | undefined): boolean {
  if (!group) return true;
  return Object.keys(group.variables ?? {}).length === 0 && Object.values(group.subgroups ?? {}).every(isEmpty);
}

export function VariableSection() {
  const { data: graph } = useSchema();
  const editor = useEditorOptional();
  const canEdit = useUiStore((s) => s.canEdit);
  const [newOpen, setNewOpen] = useState(false);

  return (
    <>
      <div className="sidebar-header sidebar-header-row">
        <span>Variables</span>
        <ProfileBar />
      </div>
      <div className="sidebar-section-content">
        {canEdit && (
          <button type="button" onClick={() => setNewOpen(true)}>
            New Variable
          </button>
        )}
        {graph?.variables && <VariableGroupRows group={graph.variables} path="" threeApp={editor?.threeApp} />}
        {isEmpty(graph?.variables) && <div>No Variables Defined</div>}
      </div>
      {editor && (
        <NewVariableModal
          open={newOpen}
          onClose={() => setNewOpen(false)}
          schemaManager={editor.schemaManager}
          nodeManager={editor.nodeManager}
        />
      )}
    </>
  );
}
