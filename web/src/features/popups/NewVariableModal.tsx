import { useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import { VARIABLE_TYPE_OPTIONS, VariableType } from "@/features/variables/variableType";
import type { SchemaManager } from "@/lib/schema_manager";
import type { NodeManager } from "@/lib/node_manager";
import { getApiErrorMessage, requestManager } from "@/api/client";
import { PortTypeList } from "./PortTypeList";

const TYPES: string[] = VARIABLE_TYPE_OPTIONS.map((opt) => opt.value);
const LABELS = new Map<string, string>(VARIABLE_TYPE_OPTIONS.map((opt) => [opt.value, opt.label]));
const typeLabel = (type: string) => LABELS.get(type) ?? type;

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
  const [type, setType] = useState<string>(VariableType.Float);

  const create = () => {
    if (!type) return;
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
          <PopupButton variant="primary" onClick={create} disabled={!type}>
            Create
          </PopupButton>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 12, width: 400 }}>
        <label>Name</label>
        <input type="text" value={name} onChange={(e) => setName(e.target.value)} />
        <span className="field-hint">A slash puts it in a group: Lighting/Sun</span>
        <label>Description</label>
        <textarea value={description} onChange={(e) => setDescription(e.target.value)} />
        <label>Type</label>
        <PortTypeList options={TYPES} selected={type} onSelect={setType} onConfirm={create} label={typeLabel} />
      </div>
    </Modal>
  );
}
