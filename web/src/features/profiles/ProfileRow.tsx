import { useState } from "react";
import { DropdownMenu } from "@/components/DropdownMenu";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { requestManager } from "@/api/client";
import { RenameProfileModal } from "@/features/popups/RenameProfileModal";
import { OverwriteProfileModal } from "@/features/popups/OverwriteProfileModal";
import { DeleteProfileModal } from "@/features/popups/DeleteProfileModal";

interface ProfileRowProps {
  profileName: string;
}

export function ProfileRow({ profileName }: ProfileRowProps) {
  const editor = useEditorOptional();
  const [renameOpen, setRenameOpen] = useState(false);
  const [overwriteOpen, setOverwriteOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  const applyProfile = () => {
    if (!editor) return;
    requestManager.applyProfile(
      profileName,
      () => editor.schemaManager.refreshSchema("Applied a profile"),
      () => alert("unable to load profile")
    );
  };

  return (
    <div className="profile-row">
      <div style={{ display: "flex", flexDirection: "row" }}>
        <button type="button" className="variable-name profile-item" onClick={applyProfile}>
          {profileName}
        </button>
        <DropdownMenu
          items={[
            { label: "Rename", onClick: () => setRenameOpen(true) },
            { label: "Overwrite", onClick: () => setOverwriteOpen(true) },
            { label: "Delete", onClick: () => setDeleteOpen(true) },
          ]}
        />
      </div>
      {editor && (
        <>
          <RenameProfileModal
            open={renameOpen}
            profileName={profileName}
            schemaManager={editor.schemaManager}
            onClose={() => setRenameOpen(false)}
          />
          <OverwriteProfileModal
            open={overwriteOpen}
            profileName={profileName}
            schemaManager={editor.schemaManager}
            onClose={() => setOverwriteOpen(false)}
          />
          <DeleteProfileModal
            open={deleteOpen}
            profileName={profileName}
            schemaManager={editor.schemaManager}
            onClose={() => setDeleteOpen(false)}
          />
        </>
      )}
    </div>
  );
}
