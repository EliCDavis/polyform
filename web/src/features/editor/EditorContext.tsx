import { createContext, useContext } from "react";
import type { SchemaManager } from "@/lib/schema_manager";
import type { NodeManager } from "@/lib/node_manager";
import type { ProducerViewManager } from "@/lib/ProducerView/producer_view_manager";
import type { ThreeApp } from "@/lib/three_app";
import type { RequestManager } from "@/lib/requests";
import type { RegisteredTypes } from "@/types/schema";
import type { NodeFlowGraph } from "@elicdavis/node-flow";

export interface EditorContextValue {
  schemaManager: SchemaManager;
  nodeManager: NodeManager;
  producerViewManager: ProducerViewManager;
  requestManager: RequestManager;
  threeApp: ThreeApp;
  nodeFlowGraph: NodeFlowGraph;
  registeredTypes: RegisteredTypes;
  ready: boolean;
}

export const EditorContext = createContext<EditorContextValue | null>(null);

export function useEditor(): EditorContextValue {
  const ctx = useContext(EditorContext);
  if (!ctx) throw new Error("useEditor must be used within EditorProvider");
  return ctx;
}

export function useEditorOptional(): EditorContextValue | null {
  return useContext(EditorContext);
}
