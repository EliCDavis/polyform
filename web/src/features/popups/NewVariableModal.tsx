import { useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import { VARIABLE_TYPE_OPTIONS, VariableType } from "@/features/variables/variableType";
import type { SchemaManager } from "@/lib/schema_manager";
import type { NodeManager } from "@/lib/node_manager";
import { getApiErrorMessage, requestManager } from "@/api/client";

interface NewVariableModalProps {
  open: boolean;
  onClose: () => void;
  schemaManager: SchemaManager;
  nodeManager: NodeManager;
}

export function NewVariableModal({
  open,
  onClose,
  schemaManager,
  nodeManager,
}: NewVariableModalProps) {
  const [name, setName] = useState("New Variable");
  const [description, setDescription] = useState("");
  const [type, setType] = useState(VariableType.Float);

  const create = () => {
    requestManager.newVariable(
      name,
      { type, description },
      (createResp) => {
        schemaManager.refreshSchema("created a variable");
        nodeManager.registerCustomNodeType(createResp.nodeType);
        onClose();
      },
      (err) => alert(getApiErrorMessage(err, "Failed to create variable"))
    );
  };

  return (
    <Modal
      title="New Variable"
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
        <label>Description</label>
        <textarea value={description} onChange={(e) => setDescription(e.target.value)} />
        <label>Type</label>
        <select value={type} onChange={(e) => setType(e.target.value as VariableType)}>
          {VARIABLE_TYPE_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>
    </Modal>
  );
}
