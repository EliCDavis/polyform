import type { NodeManager } from "@/lib/node_manager";
import type { ImportSubGraphsResult } from "@/lib/schema";

export function applyImportedSubGraphs(
  nodeManager: NodeManager,
  result: ImportSubGraphsResult
): void {
  for (const entry of result.imported) {
    nodeManager.registerCustomNodeType(entry.nodeType);
  }
  if ((result.enums ?? []).some((entry) => !entry.reused)) {
    nodeManager.refreshRegisteredTypes(() => nodeManager.notifySubGraphDefinitionChanged());
    return;
  }
  if (result.imported.length > 0) {
    nodeManager.notifySubGraphDefinitionChanged();
  }
}

export function formatImportSummary(result: ImportSubGraphsResult): string {
  const enums = result.enums ?? [];
  if (result.imported.length === 0 && enums.length === 0) {
    return "No sub-graph definitions found in the file.";
  }

  const where = result.group ? ` into "${result.group}"` : "";
  const sections = [`Imported ${result.imported.length} sub-graph(s)${where}.`];

  const remapped = result.imported.filter((entry) => entry.originalId);
  if (remapped.length > 0) {
    sections.push(
      "Renamed due to conflicts:\n" +
        remapped.map((entry) => `${entry.originalId} → ${entry.id}`).join("\n")
    );
  }

  const inputs = result.imported.flatMap((entry) =>
    (entry.inputs ?? []).map(
      (input) =>
        `${entry.name}: variable "${input.variable}" → input "${input.port}"` +
        (input.hasDefault ? "" : " (no default value)")
    )
  );
  if (inputs.length > 0) {
    sections.push("Variables turned into inputs:\n" + inputs.join("\n"));
  }

  if (enums.length > 0) {
    sections.push(
      "Enums:\n" +
        enums
          .map((entry) => {
            if (entry.reused) return `${entry.name} reused`;
            const renamed = entry.originalId && entry.originalId !== entry.id ? ` as ${entry.id}` : "";
            return `${entry.name} imported${renamed}`;
          })
          .join("\n")
    );
  }

  return sections.join("\n\n");
}
