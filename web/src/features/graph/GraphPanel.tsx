import { GraphTabBar } from "./GraphTabBar";
import { NodeFlowPanel } from "@/features/nodeFlow/NodeFlowPanel";
import { PreviewPanel } from "@/features/preview/PreviewPanel";
import { GraphTabKind, useGraphTabStore } from "@/stores/graphTabStore";
import styles from "./GraphPanel.module.css";

export function GraphPanel() {
  const active = useGraphTabStore((s) => s.tabs.find((t) => t.id === s.activeTabId));
  const preview = active?.kind === GraphTabKind.Preview ? active : null;

  return (
    <div className={styles.panel}>
      <GraphTabBar />
      <div className={styles.body}>
        <div className={styles.canvasHost} style={preview ? { display: "none" } : undefined}>
          <NodeFlowPanel />
        </div>
        {preview && <PreviewPanel key={preview.id} nodeId={preview.nodeId!} port={preview.port!} scope={preview.scope ?? null} />}
      </div>
    </div>
  );
}
