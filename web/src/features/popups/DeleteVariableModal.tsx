import { Modal, PopupButton } from "@/components/Modal";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { GeneratorVariablePublisherPath } from "@/lib/node_manager";
import { requestManager } from "@/api/client";

interface DeleteVariableModalProps {
  open: boolean;
  variableKey: string;
  onClose: () => void;
}

export function DeleteVariableModal({ open, variableKey, onClose }: DeleteVariableModalProps) {
  const editor = useEditorOptional();

  const deleteVariable = () => {
    if (!editor) return;
    requestManager.deleteVariable(
      variableKey,
      () => {
        editor.schemaManager.refreshSchema("Deleted a variable");
        editor.nodeManager.unregisterNodeType(GeneratorVariablePublisherPath + variableKey);
        onClose();
      },
      () => alert("Error deleting variable")
    );
  };

  return (
    <Modal
      title="Delete Variable"
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Cancel</PopupButton>
          <PopupButton variant="destructive" onClick={deleteVariable}>
            Delete
          </PopupButton>
        </>
      }
    >
      <p>Are you sure you want to delete {variableKey}?</p>
    </Modal>
  );
}
