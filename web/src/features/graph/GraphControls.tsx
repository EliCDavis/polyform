import { downloadBlob, requestManager } from "@/api/client";
import { useUiStore } from "@/stores/uiStore";
import { useState } from "react";
import { NewGraphModal } from "@/features/popups/NewGraphModal";

export function GraphControls() {
  const hideFileControls = useUiStore((s) => s.hideFileControls);
  const showNewGraphPopup = useUiStore((s) => s.showNewGraphPopup);
  const setShowNewGraphPopup = useUiStore((s) => s.setShowNewGraphPopup);
  const [newGraphOpen, setNewGraphOpen] = useState(showNewGraphPopup);

  if (hideFileControls) return null;

  const saveGraph = () => {
    requestManager.getGraph((graph) => {
      const bb = new Blob([JSON.stringify(graph)], { type: "application/json" });
      const a = document.createElement("a");
      a.download = "graph.json";
      a.href = URL.createObjectURL(bb);
      a.click();
    });
  };

  const loadGraph = () => {
    const input = document.createElement("input");
    input.type = "file";
    input.onchange = (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;
      const reader = new FileReader();
      reader.onload = (ev) => {
        requestManager.setGraph(JSON.parse(ev.target?.result as string), () =>
          location.reload()
        );
      };
      reader.readAsText(file);
    };
    input.click();
  };

  const saveModel = () => {
    downloadBlob("./zip/", (data) => {
      const a = document.createElement("a");
      a.download = "model.zip";
      a.href = URL.createObjectURL(data);
      a.click();
      URL.revokeObjectURL(a.href);
    });
  };

  const viewMermaid = () => {
    requestManager.fetchText("./mermaid", (data) => {
      const mermaidConfig = {
        code: data,
        mermaid: { securityLevel: "strict" },
      };
      window.open(
        "https://mermaid.live/edit#" + btoa(JSON.stringify(mermaidConfig)),
        "_blank"
      )?.focus();
    });
  };

  const saveSwagger = () => {
    requestManager.getSwagger((swagger) => {
      const bb = new Blob([JSON.stringify(swagger)], { type: "application/json" });
      const a = document.createElement("a");
      a.download = "swagger.json";
      a.href = URL.createObjectURL(bb);
      a.click();
    });
  };

  return (
    <>
      <div id="file-controls-section">
        <div className="sidebar-header">Graph</div>
        <div className="sidebar-section-content" style={{ flexDirection: "row" }}>
          <button
            type="button"
            className="sidebar-button"
            style={{ flex: 1 }}
            onClick={() => setNewGraphOpen(true)}
          >
            New
          </button>
          <button type="button" className="sidebar-button" style={{ flex: 1 }} onClick={saveGraph}>
            Save
          </button>
          <button type="button" className="sidebar-button" style={{ flex: 1 }} onClick={loadGraph}>
            Load
          </button>
        </div>
      </div>
      <div id="export-controls-section">
        <div className="sidebar-header">Export</div>
        <div className="sidebar-section-content" style={{ flexDirection: "row" }}>
          <button type="button" className="sidebar-button" style={{ flex: 1 }} onClick={saveModel}>
            Model
          </button>
          <button type="button" className="sidebar-button" style={{ flex: 1 }} onClick={viewMermaid}>
            Mermaid
          </button>
          <button type="button" className="sidebar-button" style={{ flex: 1 }} onClick={saveSwagger}>
            Swagger
          </button>
        </div>
      </div>
      <NewGraphModal
        open={newGraphOpen}
        onClose={() => {
          setNewGraphOpen(false);
          setShowNewGraphPopup(false);
        }}
      />
    </>
  );
}
