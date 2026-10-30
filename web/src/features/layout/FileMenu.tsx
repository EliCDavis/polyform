import { useState } from "react";
import { downloadBlob, requestManager } from "@/api/client";
import { DropdownMenu, type MenuItem } from "@/components/DropdownMenu";
import { NewGraphModal } from "@/features/popups/NewGraphModal";
import { EditGraphDetailsModal } from "@/features/popups/EditGraphDetailsModal";
import { useUiStore } from "@/stores/uiStore";

function download(name: string, blob: Blob) {
  const a = document.createElement("a");
  a.download = name;
  a.href = URL.createObjectURL(blob);
  a.click();
  URL.revokeObjectURL(a.href);
}

export function FileMenu() {
  const hideExportControls = useUiStore((s) => s.hideExportControls);
  const canEdit = useUiStore((s) => s.canEdit);
  const showNewGraphPopup = useUiStore((s) => s.showNewGraphPopup);
  const setShowNewGraphPopup = useUiStore((s) => s.setShowNewGraphPopup);
  const [newGraphOpen, setNewGraphOpen] = useState(showNewGraphPopup);
  const [detailsOpen, setDetailsOpen] = useState(false);

  const saveGraph = () => {
    requestManager.getGraph((graph) => download("graph.json", new Blob([JSON.stringify(graph)], { type: "application/json" })));
  };

  const loadGraph = () => {
    const input = document.createElement("input");
    input.type = "file";
    input.onchange = (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;
      const reader = new FileReader();
      reader.onload = (ev) => {
        requestManager.setGraph(JSON.parse(ev.target?.result as string), () => location.reload());
      };
      reader.readAsText(file);
    };
    input.click();
  };

  const saveModel = () => downloadBlob("./zip/", (data) => download("model.zip", data));

  const viewMermaid = () => {
    requestManager.fetchText("./mermaid", (data) => {
      const mermaidConfig = { code: data, mermaid: { securityLevel: "strict" } };
      window.open("https://mermaid.live/edit#" + btoa(JSON.stringify(mermaidConfig)), "_blank")?.focus();
    });
  };

  const saveSwagger = () => {
    requestManager.getSwagger((swagger) => download("swagger.json", new Blob([JSON.stringify(swagger)], { type: "application/json" })));
  };

  const graph: MenuItem = {
    label: "Graph",
    items: [
      { label: "New", onClick: () => setNewGraphOpen(true) },
      ...(canEdit ? [{ label: "Details", onClick: () => setDetailsOpen(true) }] : []),
      { label: "Open from disk", onClick: loadGraph },
      { label: "Save to disk", onClick: saveGraph },
    ],
  };
  const exports: MenuItem = {
    label: "Export",
    items: [
      { label: "Model", onClick: saveModel },
      { label: "Mermaid diagram", onClick: viewMermaid },
      { label: "Swagger definition", onClick: saveSwagger },
    ],
  };

  return (
    <>
      <DropdownMenu items={hideExportControls ? [graph] : [graph, exports]} />
      <NewGraphModal
        open={newGraphOpen}
        onClose={() => {
          setNewGraphOpen(false);
          setShowNewGraphPopup(false);
        }}
      />
      <EditGraphDetailsModal open={detailsOpen} onClose={() => setDetailsOpen(false)} />
    </>
  );
}
