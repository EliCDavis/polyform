import { Modal, PopupButton } from "@/components/Modal";
import type { SchemaManager } from "@/lib/schema_manager";
import type { NodeManager } from "@/lib/node_manager";
import { getApiErrorMessage, requestManager } from "@/api/client";
import type { SubGraphPlacement } from "@/features/subgraph/SubGraphRow";

interface DeleteSubGraphModalProps {
  open: boolean;
  subGraphId: string;
  name: string;
  placements: SubGraphPlacement[];
  schemaManager: SchemaManager;
  nodeManager: NodeManager;
  onDeleted: () => void;
  onClose: () => void;
}

export function DeleteSubGraphModal({
  open,
  subGraphId,
  name,
  placements,
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
      {placements.length > 0 && (
        <p>
          This also removes its {placements.length === 1 ? "placement" : `${placements.length} placements`} in{" "}
          {[...new Set(placements.map((p) => p.label))].join(", ")}.
        </p>
      )}
    </Modal>
  );
}
