import { GraphTabBar } from "./GraphTabBar";
import { NodeFlowPanel } from "@/features/nodeFlow/NodeFlowPanel";
import styles from "./GraphPanel.module.css";

export function GraphPanel() {
  return (
    <div className={styles.panel}>
      <GraphTabBar />
      <div className={styles.body}>
        <NodeFlowPanel />
      </div>
    </div>
  );
}
