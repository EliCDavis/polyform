import type { FlowNodeConfig, FlowNodeStyle } from "@elicdavis/node-flow";
import { BoundaryType } from "@/lib/schema";
import {
  SUBGRAPH_INPUT_TYPE,
  SUBGRAPH_OUTPUT_TYPE,
} from "@/lib/portTypes";

const InputStyle: FlowNodeStyle = {
  title: { color: "#2d6a4f" },
  idle: { color: "#1a3d2a" },
  mouseOver: { color: "#1f4d33" },
  grabbed: { color: "#245c3d" },
  selected: { color: "#2d6a4f" },
};

const OutputStyle: FlowNodeStyle = {
  title: { color: "#bc6c25" },
  idle: { color: "#4a2c14" },
  mouseOver: { color: "#5c3618" },
  grabbed: { color: "#6e401c" },
  selected: { color: "#804a20" },
};

const SubGraphInputNodeConfig: FlowNodeConfig = {
  title: "Input",
  subTitle: "SubGraph Boundary",
  canEditTitle: true,
  canEditInfo: false,
  inputs: [],
  outputs: [{ name: "Value", type: "any" }],
  style: InputStyle,
  metadata: {
    typeData: {
      type: SUBGRAPH_INPUT_TYPE,
      displayName: "SubGraph Input",
      info: "Sub-graph input boundary port",
      path: "SubGraph",
    },
  },
};

const SubGraphOutputNodeConfig: FlowNodeConfig = {
  title: "Output",
  subTitle: "SubGraph Boundary",
  canEditTitle: true,
  canEditInfo: false,
  inputs: [{ name: "Value", type: "any" }],
  outputs: [],
  style: OutputStyle,
  metadata: {
    typeData: {
      type: SUBGRAPH_OUTPUT_TYPE,
      displayName: "SubGraph Output",
      info: "Sub-graph output boundary port",
      path: "SubGraph",
    },
  },
};

export function buildBoundaryFlowNodeConfig(
  kind: BoundaryType,
  portType: string,
): FlowNodeConfig {
  const base = kind === BoundaryType.Input ? SubGraphInputNodeConfig : SubGraphOutputNodeConfig;
  const resolvedType = portType || "any";

  if (kind === BoundaryType.Input) {
    return {
      ...base,
      outputs: [{ name: "Value", type: resolvedType }],
    };
  }

  return {
    ...base,
    inputs: [{ name: "Value", type: resolvedType }],
  };
}

export const subGraphNodeConfigs: Record<string, FlowNodeConfig> = {
  "SubGraph/Input": SubGraphInputNodeConfig,
  "SubGraph/Output": SubGraphOutputNodeConfig,
};

export const SubGraphRuntimeStyle: FlowNodeStyle = {
  title: { color: "#4a5568" },
  idle: { color: "#2d3748" },
  mouseOver: { color: "#354050" },
  grabbed: { color: "#3d4858" },
  selected: { color: "#4a5568" },
};
