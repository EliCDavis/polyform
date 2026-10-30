import { useState } from "react";
import { useSchema } from "@/api/hooks";
import { requestManager } from "@/api/client";
import { DropdownMenu } from "@/components/DropdownMenu";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { NewProfileModal } from "@/features/popups/NewProfileModal";
import { RenameProfileModal } from "@/features/popups/RenameProfileModal";
import { DeleteProfileModal } from "@/features/popups/DeleteProfileModal";
import { useProfileStore } from "@/stores/profileStore";
import { useUiStore } from "@/stores/uiStore";

export function ProfileBar() {
  const { data: graph } = useSchema();
  const editor = useEditorOptional();
  const canEdit = useUiStore((s) => s.canEdit);
  const active = useProfileStore((s) => s.activeProfile);
  const setActive = useProfileStore((s) => s.setActiveProfile);
  const [newOpen, setNewOpen] = useState(false);
  const [renameOpen, setRenameOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  const profiles = graph?.profiles ?? [];
  const current = active !== null && profiles.includes(active) ? active : profiles[0] ?? "";

  const switchTo = (name: string) => {
    requestManager.applyProfile(
      name,
      () => {
        setActive(name);
        editor?.schemaManager.refreshSchema("switched profile");
      },
      () => alert("unable to switch profile")
    );
  };

  return (
    <div className="profile-bar">
      <select
        className="profile-bar-select"
        title="Open profile"
        value={current}
        disabled={!canEdit || profiles.length === 0}
        onChange={(e) => e.target.value !== current && switchTo(e.target.value)}
      >
        {profiles.map((name) => (
          <option key={name} value={name}>
            {name}
          </option>
        ))}
      </select>
      {canEdit && (
        <DropdownMenu
          items={[
            { label: "New Profile", onClick: () => setNewOpen(true) },
            { label: "Rename Profile", onClick: () => setRenameOpen(true) },
            ...(profiles.length > 1 ? [{ label: "Delete Profile", onClick: () => setDeleteOpen(true) }] : []),
          ]}
        />
      )}
      {editor && (
        <>
          <NewProfileModal
            open={newOpen}
            onClose={() => setNewOpen(false)}
            schemaManager={editor.schemaManager}
            onCreated={setActive}
          />
          {current && (
            <>
              <RenameProfileModal
                key={current}
                open={renameOpen}
                profileName={current}
                schemaManager={editor.schemaManager}
                onRenamed={setActive}
                onClose={() => setRenameOpen(false)}
              />
              <DeleteProfileModal
                open={deleteOpen}
                profileName={current}
                schemaManager={editor.schemaManager}
                onDeleted={() => {
                  const next = profiles.find((name) => name !== current);
                  if (next === undefined) setActive(null);
                  else switchTo(next);
                }}
                onClose={() => setDeleteOpen(false)}
              />
            </>
          )}
        </>
      )}
    </div>
  );
}
