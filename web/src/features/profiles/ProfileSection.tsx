import { useState } from "react";
import { useSchema } from "@/api/hooks";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { useUiStore } from "@/stores/uiStore";
import { ProfileRow } from "./ProfileRow";
import { NewProfileModal } from "@/features/popups/NewProfileModal";

export function ProfileSection() {
  const { data: graph } = useSchema();
  const editor = useEditorOptional();
  const canEdit = useUiStore((s) => s.canEdit);
  const [newOpen, setNewOpen] = useState(false);

  const profiles = graph?.profiles ?? [];

  return (
    <>
      <div className="sidebar-header">Profiles</div>
      <div className="sidebar-section-content">
        {canEdit && (
          <button type="button" onClick={() => setNewOpen(true)}>
            New Profile
          </button>
        )}
        {profiles.map((name) => (
          <ProfileRow key={name} profileName={name} />
        ))}
      </div>
      {editor && (
        <NewProfileModal
          open={newOpen}
          onClose={() => setNewOpen(false)}
          schemaManager={editor.schemaManager}
        />
      )}
    </>
  );
}
