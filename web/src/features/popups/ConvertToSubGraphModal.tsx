import { useEffect, useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import type { NodeManager } from "@/lib/node_manager";
import type { SchemaManager } from "@/lib/schema_manager";
import { getApiErrorMessage, requestManager } from "@/api/client";
import { useConvertSubGraphStore } from "@/stores/convertSubGraphStore";
import { useGraphTabStore } from "@/stores/graphTabStore";

interface ConvertToSubGraphModalProps {
  nodeManager: NodeManager;
  schemaManager: SchemaManager;
}

export function ConvertToSubGraphModal({
  nodeManager,
  schemaManager,
}: ConvertToSubGraphModalProps) {
  const open = useConvertSubGraphStore((s) => s.open);
  const nodeIds = useConvertSubGraphStore((s) => s.nodeIds);
  const scope = useConvertSubGraphStore((s) => s.scope);
  const close = useConvertSubGraphStore((s) => s.close);
  const openSubGraphTab = useGraphTabStore((s) => s.openSubGraphTab);

  const [name, setName] = useState("New Sub-Graph");
  const [description, setDescription] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (open) {
      setName("New Sub-Graph");
      setDescription("");
      setSubmitting(false);
    }
  }, [open]);

  const convert = () => {
    const trimmedName = name.trim();
    if (!trimmedName || nodeIds.length === 0 || submitting) return;

    setSubmitting(true);
    requestManager.convertSelectionToSubGraph(
      {
        scope,
        nodeIds,
        name: trimmedName,
        description: description.trim(),
      },
      (resp) => {
        if (resp.nodeType) {
          nodeManager.registerCustomNodeType(resp.nodeType);
        }
        // Load the post-convert schema, then open the new tab so scope sync
        // paints the definition before we center the camera.
        requestManager.getSchema((graph) => {
          schemaManager.setGraph(graph);
          useConvertSubGraphStore.getState().requestCenterOnGraph();
          openSubGraphTab(resp.subGraphId, resp.name);
          close();
        });
      },
      (err) => {
        setSubmitting(false);
        alert(getApiErrorMessage(err, "Failed to convert selection to sub-graph"));
      }
    );
  };

  return (
    <Modal
      title="Convert to Sub-Graph"
      open={open}
      onClose={close}
      actions={
        <>
          <PopupButton onClick={close}>Close</PopupButton>
          <PopupButton variant="primary" onClick={convert}>
            Convert
          </PopupButton>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 12, width: 400 }}>
        <p style={{ margin: 0, opacity: 0.8 }}>
          Convert {nodeIds.length} selected node{nodeIds.length === 1 ? "" : "s"} into a
          new sub-graph.
        </p>
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
