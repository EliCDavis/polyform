import { useState } from "react";
import { useSchema } from "@/api/hooks";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { useUiStore } from "@/stores/uiStore";
import type { EnumDefinition } from "@/lib/schema";
import { DropdownMenu } from "@/components/DropdownMenu";
import { EnumModal } from "@/features/popups/EnumModal";
import { DeleteEnumModal } from "@/features/popups/DeleteEnumModal";

function EnumRow({ id, definition }: { id: string; definition: EnumDefinition }) {
  const editor = useEditorOptional();
  const canEdit = useUiStore((s) => s.canEdit);
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  return (
    <div className="variable-row">
      <div className="variable-header">
        <span className="variable-name" title={id}>{definition.name}</span>
        {canEdit && (
          <DropdownMenu
            items={[
              { label: "Edit", onClick: () => setEditOpen(true) },
              { label: "Delete", onClick: () => setDeleteOpen(true) },
            ]}
          />
        )}
      </div>
      {definition.description && <div className="variable-description">{definition.description}</div>}
      <div className="variable-description">{definition.members.map((member) => member.name).join(", ")}</div>
      {editor && editOpen && (
        <EnumModal
          open
          enumId={id}
          initial={definition}
          onClose={() => setEditOpen(false)}
          schemaManager={editor.schemaManager}
          nodeManager={editor.nodeManager}
        />
      )}
      {editor && (
        <DeleteEnumModal
          open={deleteOpen}
          enumId={id}
          name={definition.name}
          schemaManager={editor.schemaManager}
          nodeManager={editor.nodeManager}
          onClose={() => setDeleteOpen(false)}
        />
      )}
    </div>
  );
}

export function EnumSection() {
  const { data: graph } = useSchema();
  const editor = useEditorOptional();
  const canEdit = useUiStore((s) => s.canEdit);
  const [newOpen, setNewOpen] = useState(false);

  const enums = Object.entries(graph?.enums ?? {}).sort(([, a], [, b]) => a.name.localeCompare(b.name));

  return (
    <>
      <div className="sidebar-header">Enums</div>
      <div className="sidebar-section-content">
        {canEdit && (
          <button type="button" onClick={() => setNewOpen(true)}>
            New Enum
          </button>
        )}
        {enums.map(([id, definition]) => (
          <EnumRow key={id} id={id} definition={definition} />
        ))}
        {enums.length === 0 && <div>No Enums Defined</div>}
      </div>
      {editor && newOpen && (
        <EnumModal
          open
          onClose={() => setNewOpen(false)}
          schemaManager={editor.schemaManager}
          nodeManager={editor.nodeManager}
        />
      )}
    </>
  );
}
