import { useEffect, useRef } from "react";
import { NodeFlowGraph, Publisher } from "@elicdavis/node-flow";
import { parameterNodeConfigs } from "./parameterNodeConfigs";
import { useFlowGraphBootstrap } from "./FlowGraphBootstrapContext";

export function NodeFlowPanel() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const initialized = useRef(false);
  const { registerFlowGraph } = useFlowGraphBootstrap();

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || initialized.current) return;

    initialized.current = true;

    const publisher = new Publisher({
      name: "Polyform",
      version: "1.0.0",
      nodes: parameterNodeConfigs,
    });
    const nodeFlowGraph = new NodeFlowGraph(canvas, {});
    nodeFlowGraph.addPublisher("polyform", publisher);

    registerFlowGraph({
      NodeFlowGraph: nodeFlowGraph,
      PolyformNodesPublisher: publisher,
    });
  }, [registerFlowGraph]);

  return (
    <div id="light-container">
      <canvas id="light-canvas" ref={canvasRef} />
    </div>
  );
}
