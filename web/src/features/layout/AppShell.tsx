import { AppOverlays } from "@/components/AppOverlays";
import { ResizablePanels } from "./ResizablePanels";
import { Sidebar } from "./Sidebar";
import { ViewportPanel } from "@/features/viewport/ViewportPanel";
import { GraphPanel } from "@/features/graph/GraphPanel";
import { useUiStore } from "@/stores/uiStore";

export function AppShell() {
  const hideGraph = useUiStore((s) => s.hideGraph);
  const sidebarCollapsed = useUiStore((s) => s.sidebarCollapsed);
  const compactLayout = useUiStore((s) => s.compactLayout);
  const setSidebarCollapsed = useUiStore((s) => s.setSidebarCollapsed);

  return (
    <>
      <div id="running-message">Running...</div>
      <AppOverlays />
      {compactLayout && !sidebarCollapsed && (
        <div id="sidebar-backdrop" onClick={() => setSidebarCollapsed(true)} />
      )}
      <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
        <div id="full-page">
          <ResizablePanels
            direction="horizontal"
            initialFirstPercent={25}
            collapsed={sidebarCollapsed || compactLayout}
            firstCollapsedSize="63px"
            first={<Sidebar />}
            second={
              <div id="main-content">
                {hideGraph ? (
                  <ViewportPanel />
                ) : (
                  <ResizablePanels
                    direction="vertical"
                    initialFirstPercent={compactLayout ? 50 : 40}
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
