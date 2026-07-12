export enum GraphScopeKind {
  Root = "root",
  SubGraph = "subgraph",
}

export type GraphScope =
  | { kind: GraphScopeKind.Root }
  | { kind: GraphScopeKind.SubGraph; id: string };

export const ROOT_SCOPE: GraphScope = { kind: GraphScopeKind.Root };

export function subgraphScope(id: string): GraphScope {
  return { kind: GraphScopeKind.SubGraph, id };
}

export function scopeToApiPath(scope: GraphScope): string | null {
  if (scope.kind === GraphScopeKind.Root) return null;
  return `graph/subgraph/${scope.id}`;
}

export function formatPortTypeLabel(type: string): string {
  return type.split("github.com/EliCDavis/").join("");
}

export function isSubGraphRuntimeType(type: string): boolean {
  return type.startsWith("subgraph/");
}

export function subGraphRuntimeType(id: string): string {
  return `subgraph/${id}`;
}

export const SUBGRAPH_INPUT_TYPE =
  "github.com/EliCDavis/polyform/generator/subgraph.InputNode";
export const SUBGRAPH_OUTPUT_TYPE =
  "github.com/EliCDavis/polyform/generator/subgraph.OutputNode";
