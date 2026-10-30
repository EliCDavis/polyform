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
  group?: string;
  groups?: string[];
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
  group = "",
  groups = [],
  schemaManager,
  nodeManager,
  onRenamed,
  onClose,
}: RenameSubGraphModalProps) {
  const [nextName, setNextName] = useState(name);
  const [nextDescription, setNextDescription] = useState(description);
  const [nextGroup, setNextGroup] = useState(group);

  useEffect(() => {
    if (!open) return;
    setNextName(name);
    setNextDescription(description);
    setNextGroup(group);
  }, [open, name, description, group]);

  const rename = () => {
    const trimmedName = nextName.trim();
    if (!trimmedName) {
      alert("Name can not be empty");
      return;
    }

    if (trimmedName === name && nextDescription.trim() === description.trim() && nextGroup.trim() === group) {
      onClose();
      return;
    }

    requestManager.updateSubGraphInfo(
      subGraphId,
      { name: trimmedName, description: nextDescription.trim(), group: nextGroup.trim() },
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
      title="Edit Sub-Graph"
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
        <label>Group</label>
        <input
          type="text"
          list="subgraph-groups"
          value={nextGroup}
          onChange={(e) => setNextGroup(e.target.value)}
          placeholder="Optional heading it is listed under"
        />
        <datalist id="subgraph-groups">
          {groups.map((g) => (
            <option key={g} value={g} />
          ))}
        </datalist>
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
