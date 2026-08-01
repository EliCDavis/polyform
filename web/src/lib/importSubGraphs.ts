import type { NodeManager } from "@/lib/node_manager";
import type { ImportSubGraphsResult } from "@/lib/schema";

export function applyImportedSubGraphs(
  nodeManager: NodeManager,
  result: ImportSubGraphsResult
): void {
  for (const entry of result.imported) {
    nodeManager.registerCustomNodeType(entry.nodeType);
  }
  if (result.imported.length > 0) {
    nodeManager.notifySubGraphDefinitionChanged();
  }
}

export function formatImportRemapSummary(result: ImportSubGraphsResult): string | null {
  const remapped = result.imported.filter((entry) => entry.originalId);
  if (remapped.length === 0) {
    return null;
  }
  return remapped
    .map((entry) => `${entry.originalId} → ${entry.id}`)
    .join("\n");
}
