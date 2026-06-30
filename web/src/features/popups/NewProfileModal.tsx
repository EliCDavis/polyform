import { useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import type { SchemaManager } from "@/lib/schema_manager";
import { requestManager } from "@/api/client";

interface NewProfileModalProps {
  open: boolean;
  onClose: () => void;
  schemaManager: SchemaManager;
}

export function NewProfileModal({ open, onClose, schemaManager }: NewProfileModalProps) {
  const [name, setName] = useState("New Profile");

  const create = () => {
    requestManager.newProfile(
      name || "New Profile",
      () => {
        schemaManager.refreshSchema("created a profile");
        onClose();
      },
      () => alert("unable to create profile")
    );
  };

  return (
    <Modal
      title="New Profile"
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Close</PopupButton>
          <PopupButton variant="primary" onClick={create}>
            Create
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
