import { useMemo, useState } from "react";
import type { NodeDefinition, NodeInput, NodeOutput } from "@/lib/schema";
import { SUBGRAPH_INPUT_TYPE, SUBGRAPH_OUTPUT_TYPE } from "@/lib/portTypes";
import { useEditorOptional } from "../editor/EditorContext";

const PORT_TYPE_PREFIXES = [
  "github.com/EliCDavis/polyform/",
  "github.com/EliCDavis/vector/",
];

function shortenPortType(type: string): string {
  let arrAppend = "";
  let cleanedType = type;
  if (type.startsWith("[]")) {
    cleanedType = type.slice(2);
    arrAppend = "[]";
  }

  let pointerAppend = "";
  if (type.startsWith("*")) {
    cleanedType = type.slice(1);
    pointerAppend = "*";
  }

  for (const prefix of PORT_TYPE_PREFIXES) {
    if (cleanedType.startsWith(prefix)) {
      return arrAppend + pointerAppend + cleanedType.slice(prefix.length);
    }
  }
  return type;
}

function buildSearchableText(def: NodeDefinition): string {
  const parts: string[] = [def.displayName, def.path, def.info ?? ""];

//   const addPorts = (ports?: { [key: string]: NodeOutput | NodeInput }) => {
//     if (!ports) return;
//     for (const [name, portDef] of Object.entries(ports)) {
//       parts.push(name);
//       parts.push(shortenPortType(portDef.type));
//     }
//   };
//   addPorts(def.inputs);
//   addPorts(def.outputs);

  return parts.join(" ").toLowerCase();
}

function PortList({
  title,
  ports,
}: {
  title: string;
  ports: { [key: string]: NodeOutput | NodeInput };
}) {
  const entries = Object.entries(ports);
  if (entries.length === 0) return null;

  return (
    <div className="node-catalog-ports">
      <span className="node-catalog-ports-title">{title}</span>
      {entries.map(([name, def]) => (
        <div className="node-catalog-port" key={name}>
          <span>{name}</span>
          <span className="node-catalog-port-type">
            {shortenPortType(def.type)}
          </span>
        </div>
      ))}
    </div>
  );
}

interface NodeCatalogEntryProps {
  def: NodeDefinition;
  onAdd: () => void;
}

function NodeCatalogEntry({ def, onAdd }: NodeCatalogEntryProps) {
  return (
    <div className="node-catalog-entry">
      <div className="node-catalog-entry-header">
        <div className="node-catalog-entry-name">{def.displayName}</div>
        <button type="button" onClick={onAdd}>
          Add
        </button>
      </div>
      {def.info && <div className="variable-description">{def.info}</div>}
      {def.inputs && <PortList title="Inputs" ports={def.inputs} />}
      {def.outputs && <PortList title="Outputs" ports={def.outputs} />}
    </div>
  );
}

function pluralize(count: number, noun: string): string {
  return `${count} ${noun}${count === 1 ? "" : "s"}`;
}

interface SearchableNode {
  def: NodeDefinition;
  text: string;
}

interface NodeGroup {
  path: string;
  nodes: SearchableNode[];
}

export function NodesSection() {
  const editor = useEditorOptional();
  const [search, setSearch] = useState("");

  const addableNodeTypes = useMemo(() => {
    const nodeTypes = editor?.registeredTypes.nodeTypes ?? [];
    return nodeTypes.filter(
      (def) =>
        !def.parameter &&
        def.type !== SUBGRAPH_INPUT_TYPE &&
        def.type !== SUBGRAPH_OUTPUT_TYPE,
    );
  }, [editor?.registeredTypes.nodeTypes]);

  const nodeGroups = useMemo((): NodeGroup[] => {
    const groups = new Map<string, SearchableNode[]>();
    for (const def of addableNodeTypes) {
      const path = def.path || "(uncategorized)";
      const nodes = groups.get(path) ?? [];
      nodes.push({ def, text: buildSearchableText(def) });
      groups.set(path, nodes);
    }
    return Array.from(groups.entries())
      .map(([path, nodes]) => ({ path, nodes }))
      .sort((a, b) => a.path.localeCompare(b.path));
  }, [addableNodeTypes]);

  const filteredGroups = useMemo(() => {
    const terms = search.trim().toLowerCase().split(/\s+/).filter(Boolean);
    return nodeGroups
      .map((group) => ({
        path: group.path,
        total: group.nodes.length,
        matches:
          terms.length === 0
            ? group.nodes
            : group.nodes.filter(({ text }) =>
                terms.every((term) => text.includes(term)),
              ),
      }))
      .filter((group) => group.matches.length > 0);
  }, [nodeGroups, search]);

  const totalMatchCount = useMemo(
    () => filteredGroups.reduce((sum, group) => sum + group.matches.length, 0),
    [filteredGroups],
  );

  if (!editor) return null;

  return (
    <>
      <div className="sidebar-header">Nodes</div>
      <div className="sidebar-section-content">
        <input
          type="text"
          placeholder="Search nodes..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className="node-catalog-result-count">
          {totalMatchCount} of {pluralize(addableNodeTypes.length, "node")}
        </div>
        {filteredGroups.length === 0 && (
          <div className="variable-description">
            No nodes match your search.
          </div>
        )}
        {filteredGroups.map((group) => (
          <div className="node-catalog-group" key={group.path}>
            <div className="node-catalog-group-header">
              <span className="node-catalog-group-name">{group.path}</span>
              <span className="node-catalog-group-count">
                {group.matches.length} of {pluralize(group.total, "node")}
              </span>
            </div>
            {group.matches.map(({ def }) => (
              <NodeCatalogEntry
                key={def.type}
                def={def}
                onAdd={() => editor.nodeManager.createNodeFromType(def.type)}
              />
            ))}
          </div>
        ))}
      </div>
    </>
  );
}
