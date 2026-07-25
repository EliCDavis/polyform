import { useEffect, useRef } from "react";
import {
  ContextMenuItemState,
  NodeFlowGraph,
  Publisher,
} from "@elicdavis/node-flow";
import { parameterNodeConfigs } from "./parameterNodeConfigs";
import { useFlowGraphBootstrap } from "./FlowGraphBootstrapContext";
import { InstanceIDProperty } from "@/lib/nodes/node";
import { useConvertSubGraphStore } from "@/stores/convertSubGraphStore";
import { useGraphTabStore } from "@/stores/graphTabStore";
import styles from "./NodeFlowPanel.module.css";

export function NodeFlowPanel() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const initialized = useRef(false);
  const { registerFlowGraph } = useFlowGraphBootstrap();

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || initialized.current) return;

    initialized.current = true;

    // SubGraph Input/Output are registered by NodeManager only while a
    // sub-graph tab is active (they require a port type at create time).
    const publisher = new Publisher({
      name: "Polyform",
      version: "1.0.0",
      nodes: { ...parameterNodeConfigs },
    });

    const nodeFlowGraph = new NodeFlowGraph(canvas, {
      contextMenu: {
        items: [
          {
            name: "Convert to Subgraph",
            enabled: () =>
              nodeFlowGraph.getSelectedNodes().length > 0
                ? ContextMenuItemState.Enabled
                : ContextMenuItemState.Hidden,
            callback: () => {
              const nodeIds = nodeFlowGraph
                .getSelectedNodes()
                .map((node) => node.getProperty(InstanceIDProperty) as string)
                .filter((id) => typeof id === "string" && id.length > 0);
              if (nodeIds.length === 0) return;

              const activeTabId = useGraphTabStore.getState().activeTabId;
              const scope =
                activeTabId === "root" ? "root" : `subgraph/${activeTabId}`;
              useConvertSubGraphStore.getState().openConvert(nodeIds, scope);
            },
          },
        ],
      },
    });
    nodeFlowGraph.addPublisher("polyform", publisher);

    registerFlowGraph({
      NodeFlowGraph: nodeFlowGraph,
      PolyformNodesPublisher: publisher,
    });
  }, [registerFlowGraph]);

  return (
    <div className={styles.container}>
      <canvas className={styles.canvas} ref={canvasRef} />
    </div>
  );
}
