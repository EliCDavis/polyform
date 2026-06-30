import { useState } from "react";
import { useSchema } from "@/api/hooks";
import { getVariablesFromGraph } from "@/api/hooks";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { useUiStore } from "@/stores/uiStore";
import { VariableRow } from "./VariableRow";
import { NewVariableModal } from "@/features/popups/NewVariableModal";

export function VariableSection() {
  const { data: graph } = useSchema();
  const editor = useEditorOptional();
  const canEdit = useUiStore((s) => s.canEdit);
  const [newOpen, setNewOpen] = useState(false);

  const variables = getVariablesFromGraph(graph);

  return (
    <>
      <div className="sidebar-header">Variables</div>
      <div className="sidebar-section-content">
        {canEdit && (
          <button type="button" onClick={() => setNewOpen(true)}>
            New Variable
          </button>
        )}
        {variables.map(({ key, variable }) => (
          <VariableRow
            key={key}
            variableKey={key}
            variable={variable}
            threeApp={editor?.threeApp}
          />
        ))}
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
