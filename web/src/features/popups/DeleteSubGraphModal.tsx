import { Modal, PopupButton } from "@/components/Modal";
import type { SchemaManager } from "@/lib/schema_manager";
import type { NodeManager } from "@/lib/node_manager";
import { getApiErrorMessage, requestManager } from "@/api/client";

interface DeleteSubGraphModalProps {
  open: boolean;
  subGraphId: string;
  name: string;
  schemaManager: SchemaManager;
  nodeManager: NodeManager;
  onDeleted: () => void;
  onClose: () => void;
}

export function DeleteSubGraphModal({
  open,
  subGraphId,
  name,
  schemaManager,
  nodeManager,
  onDeleted,
  onClose,
}: DeleteSubGraphModalProps) {
  const deleteSubGraph = () => {
    requestManager.deleteSubGraph(
      subGraphId,
      () => {
        nodeManager.unregisterRuntimeSubGraphType(subGraphId);
        schemaManager.refreshSchema("deleted sub-graph");
        onDeleted();
        onClose();
      },
      (err) => alert(getApiErrorMessage(err, "Failed to delete sub-graph"))
    );
  };

  return (
    <Modal
      title="Delete Sub-Graph"
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Close</PopupButton>
          <PopupButton variant="destructive" onClick={deleteSubGraph}>
            Delete
          </PopupButton>
        </>
      }
    >
      <p>
        Are you sure you want to delete <strong>{name}</strong>?
      </p>
      <p>This cannot be undone. Sub-graph nodes on the main graph must be removed first.</p>
    </Modal>
  );
}
