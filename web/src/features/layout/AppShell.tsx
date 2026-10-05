import { AppOverlays } from "@/components/AppOverlays";
import { ResizablePanels } from "./ResizablePanels";
import { Sidebar } from "./Sidebar";
import { ViewportPanel } from "@/features/viewport/ViewportPanel";
import { GraphPanel } from "@/features/graph/GraphPanel";
import { useUiStore } from "@/stores/uiStore";

export function AppShell() {
  const hideGraph = useUiStore((s) => s.hideGraph);

  return (
    <>
      <div id="running-message">Running...</div>
      <AppOverlays />
      <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
        <div id="full-page">
          <ResizablePanels
            direction="horizontal"
            initialFirstPercent={25}
            first={<Sidebar />}
            second={
              <div id="main-content">
                {hideGraph ? (
                  <ViewportPanel />
                ) : (
                  <ResizablePanels
                    direction="vertical"
                    initialFirstPercent={40}
                    first={<ViewportPanel />}
                    second={<GraphPanel />}
                  />
                )}
              </div>
            }
          />
        </div>
      </div>
    </>
  );
}
