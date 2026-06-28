import { useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import type { NodeManager } from "@/lib/node_manager";
import { getApiErrorMessage, requestManager } from "@/api/client";

interface NewSubGraphModalProps {
  open: boolean;
  onClose: () => void;
  nodeManager: NodeManager;
  onCreated: (id: string, name: string) => void;
}

export function NewSubGraphModal({
  open,
  onClose,
  nodeManager,
  onCreated,
}: NewSubGraphModalProps) {
  const [name, setName] = useState("New Sub-Graph");
  const [description, setDescription] = useState("");

  const create = () => {
    const trimmedName = name.trim();
    const id = trimmedName.replace(/\s+/g, "_");
    if (!id) return;

    requestManager.createSubGraph(
      id,
      { name: trimmedName, description: description.trim() },
      (resp) => {
        nodeManager.notifySubGraphDefinitionChanged(resp.nodeType);
        onCreated(id, trimmedName);
        onClose();
      },
      (err) => alert(getApiErrorMessage(err, "Failed to create sub-graph"))
    );
  };

  return (
    <Modal
      title="New Sub-Graph"
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
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Sub-graph name"
        />
        <label>Description</label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Optional description"
        />
      </div>
    </Modal>
  );
}
