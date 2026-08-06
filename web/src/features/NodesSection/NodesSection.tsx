import { useMemo, useState } from "react";
import { X } from "lucide-react";
import type { NodeDefinition, NodeInput, NodeOutput } from "@/lib/schema";
import { SUBGRAPH_INPUT_TYPE, SUBGRAPH_OUTPUT_TYPE } from "@/lib/portTypes";
import { useEditorOptional } from "../editor/EditorContext";

const PORT_TYPE_PREFIXES = [
  "github.com/EliCDavis/polyform/",
  "github.com/EliCDavis/vector/",
];

function stripKnownPrefix(type: string): string {
  for (const prefix of PORT_TYPE_PREFIXES) {
    if (type.startsWith(prefix)) {
      return type.slice(prefix.length);
    }
  }
  return type;
}

/** Splits the contents of a generic type's brackets on top-level commas, ignoring commas nested inside their own brackets. */
function splitGenericParams(params: string): string[] {
  const parts: string[] = [];
  let depth = 0;
  let start = 0;
  for (let i = 0; i < params.length; i++) {
    const char = params[i];
    if (char === "[") depth++;
    else if (char === "]") depth--;
    else if (char === "," && depth === 0) {
      parts.push(params.slice(start, i));
      start = i + 1;
    }
  }
  parts.push(params.slice(start));
  return parts;
}

function shortenPortType(type: string): string {
  let prefixSymbols = "";
  let rest = type;

  if (rest.startsWith("[]")) {
    prefixSymbols += "[]";
    rest = rest.slice(2);
  }
  if (rest.startsWith("*")) {
    prefixSymbols += "*";
    rest = rest.slice(1);
  }

  const bracketIndex = rest.indexOf("[");
  if (bracketIndex !== -1 && rest.endsWith("]")) {
    const base = rest.slice(0, bracketIndex);
    const params = splitGenericParams(rest.slice(bracketIndex + 1, -1));
    return (
      prefixSymbols +
      stripKnownPrefix(base) +
      "[" +
      params.map(shortenPortType).join(",") +
      "]"
    );
  }

  return prefixSymbols + stripKnownPrefix(rest);
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
        <div className="node-catalog-port-entry" key={name}>
          <div className="node-catalog-port">
            <span>{name}</span>
            <span className="node-catalog-port-type">
              {shortenPortType(def.type)}
            </span>
          </div>
          {def.description && (
            <div className="node-catalog-port-description">
              {def.description}
            </div>
          )}
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

function formatCount(matches: number, total: number, noun: string): string {
  if (matches === total) return pluralize(total, noun);
  return `${matches} of ${pluralize(total, noun)}`;
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
        <div className="node-catalog-search">
          <input
            type="text"
            placeholder="Search nodes..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          {search.length > 0 && (
            <button
              type="button"
              className="icon-button node-catalog-search-clear"
              onClick={() => setSearch("")}
              aria-label="Clear search"
            >
              <X size={14} />
            </button>
          )}
        </div>
        <div className="node-catalog-result-count">
          {formatCount(totalMatchCount, addableNodeTypes.length, "node")}
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
                {formatCount(group.matches.length, group.total, "node")}
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
