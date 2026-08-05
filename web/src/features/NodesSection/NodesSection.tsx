import { NodeInput, NodeOutput } from "@/lib/schema";
import { useEditorOptional } from "../editor/EditorContext";

function PortSection(
  name: string,
  ports: { [key: string]: NodeOutput | NodeInput },
) {
  return (
    <>
      <div style={{ marginLeft: "20px" }}>
        <b>{name}</b>
        {Object.entries(ports).map(([name, def]) => Port(name, def))}
      </div>
    </>
  );
}

function Port(name: string, portDef: NodeOutput | NodeInput) {
  return (
    <div>
      {name}: {portDef.type}
    </div>
  );
}

export function NodesSection() {
  const editor = useEditorOptional();

  const nodeTypes = editor.registeredTypes.nodeTypes;

  return (
    <>
      <div className="sidebar-header">Nodes</div>
      <div className="sidebar-section-content">
        {nodeTypes.map((def) => (
          <div style={{ marginBottom: "16px" }}>
            <div style={{ fontSize: "14px" }}>{def.displayName}</div>
            {def.info && <div>{def.info}</div>}
            {def.inputs && PortSection("Inputs", def.inputs)}
            {def.outputs && PortSection("Outputs", def.outputs)}
          </div>
        ))}
      </div>
    </>
  );
}
