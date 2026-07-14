import { useState } from "react";
import { DropdownMenu } from "@/components/DropdownMenu";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { RenameSubGraphModal } from "@/features/popups/RenameSubGraphModal";
import { DeleteSubGraphModal } from "@/features/popups/DeleteSubGraphModal";
import { useGraphTabStore } from "@/stores/graphTabStore";

interface SubGraphRowProps {
  id: string;
  name: string;
  description?: string;
}

export function SubGraphRow({ id, name, description }: SubGraphRowProps) {
  const editor = useEditorOptional();
  const openSubGraphTab = useGraphTabStore((s) => s.openSubGraphTab);
  const renameSubGraphTab = useGraphTabStore((s) => s.renameSubGraphTab);
  const closeTab = useGraphTabStore((s) => s.closeTab);
  const activeTabId = useGraphTabStore((s) => s.activeTabId);
  const isActive = activeTabId === id;
  const [renameOpen, setRenameOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  return (
    <div className="profile-row">
      <div style={{ display: "flex", flexDirection: "row" }}>
        <button
          type="button"
          className="variable-name profile-item"
          onClick={() => openSubGraphTab(id, name)}
          style={isActive ? { backgroundColor: "rgb(25 110 108)" } : undefined}
        >
          {name}
        </button>
        <DropdownMenu
          items={[
            { label: "Rename", onClick: () => setRenameOpen(true) },
            { label: "Delete", onClick: () => setDeleteOpen(true) },
          ]}
        />
      </div>
      {editor && (
        <>
          <RenameSubGraphModal
            open={renameOpen}
            subGraphId={id}
            name={name}
            description={description}
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
