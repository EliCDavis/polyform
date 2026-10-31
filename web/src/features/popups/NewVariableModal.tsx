import { useMemo, useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import { VARIABLE_TYPE_OPTIONS, VariableType } from "@/features/variables/variableType";
import type { SchemaManager } from "@/lib/schema_manager";
import type { NodeManager } from "@/lib/node_manager";
import type { EnumDefinition } from "@/types/schema";
import { getApiErrorMessage, requestManager } from "@/api/client";
import { useSchema } from "@/api/hooks";
import { ENUM_TYPE_PREFIX } from "@/lib/portTypes";
import { PortTypeList } from "./PortTypeList";

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

  const { data: graph } = useSchema();
  const { types, labels, values } = useMemo(() => {
    // Built-ins keep their order and labels; the enums follow.
    const enums: Array<[string, EnumDefinition]> = [
      ...Object.entries(nodeManager.goEnums()),
      ...Object.entries(graph?.enums ?? {}).map(([id, def]): [string, EnumDefinition] => [ENUM_TYPE_PREFIX + id, def]),
    ];
    const options = [...VARIABLE_TYPE_OPTIONS, ...enums.map(([type, def]) => ({ value: type, label: def.name }))];
    return {
      types: options.map((opt) => opt.value),
      labels: new Map(options.map((opt) => [opt.value, opt.label])),
      values: new Map(enums.map(([type, def]) => [type, def.members.map((m) => m.name)])),
    };
  }, [nodeManager, graph, open]);
  const typeLabel = (t: string) => labels.get(t) ?? t;
  const selectedValues = values.get(type);

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
        <PortTypeList options={types} selected={type} onSelect={setType} onConfirm={create} label={typeLabel} />
        {selectedValues && <span className="field-hint">One of: {selectedValues.join(", ")}</span>}
      </div>
    </Modal>
  );
}
