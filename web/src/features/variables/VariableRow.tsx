import { useState } from "react";
import type { Variable } from "@/types/schema";
import type { ThreeApp } from "@/lib/three_app";
import { DropdownMenu } from "@/components/DropdownMenu";
import { VariableValueEditor } from "./VariableValueEditor";
import { EditVariableModal } from "@/features/popups/EditVariableModal";
import { DeleteVariableModal } from "@/features/popups/DeleteVariableModal";

interface VariableRowProps {
  variableKey: string;
  variable: Variable;
  threeApp?: ThreeApp;
}

export function VariableRow({ variableKey, variable, threeApp }: VariableRowProps) {
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  return (
    <div className="variable-row">
      <div className="variable-header">
        <span className="variable-name">{variableKey}</span>
        <DropdownMenu
          items={[
            { label: "Edit", onClick: () => setEditOpen(true) },
            { label: "Delete", onClick: () => setDeleteOpen(true) },
          ]}
        />
      </div>
      {variable.description && (
        <div className="variable-description">{variable.description}</div>
      )}
      <VariableValueEditor variableKey={variableKey} variable={variable} threeApp={threeApp} />
      <EditVariableModal
        open={editOpen}
        variableKey={variableKey}
        variable={variable}
        onClose={() => setEditOpen(false)}
      />
      <DeleteVariableModal
        open={deleteOpen}
        variableKey={variableKey}
        onClose={() => setDeleteOpen(false)}
      />
    </div>
  );
}
