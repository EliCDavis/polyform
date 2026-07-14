import { useEffect, useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import type { SchemaManager } from "@/lib/schema_manager";
import type { NodeManager } from "@/lib/node_manager";
import { getApiErrorMessage, requestManager } from "@/api/client";

interface RenameSubGraphModalProps {
  open: boolean;
  subGraphId: string;
  name: string;
  description?: string;
  schemaManager: SchemaManager;
  nodeManager: NodeManager;
  onRenamed: (name: string) => void;
  onClose: () => void;
}

export function RenameSubGraphModal({
  open,
  subGraphId,
  name,
  description = "",
  schemaManager,
  nodeManager,
  onRenamed,
  onClose,
}: RenameSubGraphModalProps) {
  const [nextName, setNextName] = useState(name);
  const [nextDescription, setNextDescription] = useState(description);

  useEffect(() => {
    if (!open) return;
    setNextName(name);
    setNextDescription(description);
  }, [open, name, description]);

  const rename = () => {
    const trimmedName = nextName.trim();
    if (!trimmedName) {
      alert("Name can not be empty");
      return;
    }

    if (trimmedName === name && nextDescription.trim() === description.trim()) {
      onClose();
      return;
    }

    requestManager.updateSubGraphInfo(
      subGraphId,
      { name: trimmedName, description: nextDescription.trim() },
      () => {
        nodeManager.refreshRuntimeSubGraphType(subGraphId, () => {
          schemaManager.refreshSchema("renamed sub-graph");
          onRenamed(trimmedName);
          onClose();
        });
      },
      (err) => alert(getApiErrorMessage(err, "Failed to update sub-graph"))
    );
  };

  return (
    <Modal
      title="Rename Sub-Graph"
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Close</PopupButton>
          <PopupButton variant="primary" onClick={rename}>
            Save
          </PopupButton>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 12, width: 400 }}>
        <label>Name</label>
        <input type="text" value={nextName} onChange={(e) => setNextName(e.target.value)} />
        <label>Description</label>
        <textarea
          value={nextDescription}
          onChange={(e) => setNextDescription(e.target.value)}
          placeholder="Optional description"
        />
      </div>
    </Modal>
  );
}
