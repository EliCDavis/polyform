import { Modal, PopupButton } from "@/components/Modal";
import type { SchemaManager } from "@/lib/schema_manager";
import type { NodeManager } from "@/lib/node_manager";
import { ENUM_TYPE_PREFIX } from "@/lib/portTypes";
import { getApiErrorMessage, requestManager } from "@/api/client";

interface DeleteEnumModalProps {
  open: boolean;
  enumId: string;
  name: string;
  schemaManager: SchemaManager;
  nodeManager: NodeManager;
  onClose: () => void;
}

export function DeleteEnumModal({ open, enumId, name, schemaManager, nodeManager, onClose }: DeleteEnumModalProps) {
  const deleteEnum = () => {
    requestManager.deleteEnum(
      enumId,
      () => {
        nodeManager.unregisterEnumParameterType(ENUM_TYPE_PREFIX + enumId);
        nodeManager.refreshRegisteredTypes(() => schemaManager.refreshSchema("deleted an enum"));
        onClose();
      },
      (err) => alert(getApiErrorMessage(err, "Failed to delete enum"))
    );
  };

  return (
    <Modal
      title="Delete Enum"
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Cancel</PopupButton>
          <PopupButton variant="destructive" onClick={deleteEnum}>
            Delete
          </PopupButton>
        </>
      }
    >
      <p>
        Are you sure you want to delete <strong>{name}</strong>?
      </p>
      <p>Variables, parameters and ports still using it must be removed first.</p>
    </Modal>
  );
}
