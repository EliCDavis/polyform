import { useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import type { SchemaManager } from "@/lib/schema_manager";
import { requestManager } from "@/api/client";

interface RenameProfileModalProps {
  open: boolean;
  profileName: string;
  schemaManager: SchemaManager;
  onClose: () => void;
}

export function RenameProfileModal({
  open,
  profileName,
  schemaManager,
  onClose,
}: RenameProfileModalProps) {
  const [name, setName] = useState(profileName);

  const rename = () => {
    if (name === profileName) {
      onClose();
      return;
    }
    if (name.trim() === "") {
      alert("Name can not be empty");
      return;
    }
    requestManager.renameProfile(
      profileName,
      name,
      () => {
        schemaManager.refreshSchema("renamed a profile");
        onClose();
      },
      () => alert("unable to rename profile")
    );
  };

  return (
    <Modal
      title="Rename Profile"
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Close</PopupButton>
          <PopupButton variant="primary" onClick={rename}>
            Rename
          </PopupButton>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 12, width: 400 }}>
        <label>Name</label>
        <input type="text" value={name} onChange={(e) => setName(e.target.value)} />
      </div>
    </Modal>
  );
}
