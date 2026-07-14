import type { GraphInstance, GraphInstanceNodes } from "@/lib/schema";
import { GraphScopeKind, type GraphScope } from "@/lib/portTypes";

export function getScopedNodes(
  schema: GraphInstance,
  scope: GraphScope,
): GraphInstanceNodes {
  if (scope.kind === GraphScopeKind.Root) return schema.nodes ?? {};
  return schema.subGraphs?.[scope.id]?.nodes ?? {};
}

export function getScopedProducers(
  schema: GraphInstance,
  scope: GraphScope,
): GraphInstance["producers"] {
  if (scope.kind === GraphScopeKind.Root) return schema.producers ?? {};
  return {};
}

export function getScopedNotes(
  schema: GraphInstance,
  scope: GraphScope,
): GraphInstance["notes"] {
  if (scope.kind === GraphScopeKind.Root) return schema.notes ?? {};
  return schema.subGraphs?.[scope.id]?.notes ?? {};
}
