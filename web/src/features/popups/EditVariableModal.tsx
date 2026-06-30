import { useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import type { Variable } from "@/types/schema";
import { getApiErrorMessage, requestManager } from "@/api/client";

interface EditVariableModalProps {
  open: boolean;
  variableKey: string;
  variable: Variable;
  onClose: () => void;
}

export function EditVariableModal({
  open,
  variableKey,
  variable,
  onClose,
}: EditVariableModalProps) {
  const [name, setName] = useState(variableKey);
  const [description, setDescription] = useState(variable.description ?? "");

  const save = () => {
    if (name === variableKey && description === variable.description) {
      onClose();
      return;
    }
    requestManager.updateVariable(
      variableKey,
      { name, description },
      () => location.reload(),
      (err) => alert(getApiErrorMessage(err, "Failed to update variable"))
    );
  };

  return (
    <Modal
      title="Edit Variable"
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Cancel</PopupButton>
          <PopupButton variant="primary" onClick={save}>
            Save
          </PopupButton>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 12, width: 400 }}>
        <label>Name</label>
        <input type="text" value={name} onChange={(e) => setName(e.target.value)} />
        <label>Description</label>
        <textarea value={description} onChange={(e) => setDescription(e.target.value)} />
      </div>
    </Modal>
  );
}
