import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import type { NodeFlowGraph, Publisher } from "@elicdavis/node-flow";

export interface FlowGraphInit {
  PolyformNodesPublisher: Publisher;
  NodeFlowGraph: NodeFlowGraph;
}

interface FlowGraphBootstrapContextValue {
  flowGraphInit: FlowGraphInit | null;
  registerFlowGraph: (init: FlowGraphInit) => void;
}

const FlowGraphBootstrapContext = createContext<FlowGraphBootstrapContextValue | null>(null);

export function FlowGraphBootstrapProvider({ children }: { children: ReactNode }) {
  const [flowGraphInit, setFlowGraphInit] = useState<FlowGraphInit | null>(null);

  const registerFlowGraph = useCallback((init: FlowGraphInit) => {
    setFlowGraphInit((current) => current ?? init);
  }, []);

  const value = useMemo(
    () => ({ flowGraphInit, registerFlowGraph }),
    [flowGraphInit, registerFlowGraph]
  );

  return (
    <FlowGraphBootstrapContext.Provider value={value}>{children}</FlowGraphBootstrapContext.Provider>
  );
}

export function useFlowGraphBootstrap() {
  const ctx = useContext(FlowGraphBootstrapContext);
  if (!ctx) {
    throw new Error("useFlowGraphBootstrap must be used within FlowGraphBootstrapProvider");
  }
  return ctx;
}

export function useFlowGraphInit(): FlowGraphInit | null {
  const ctx = useContext(FlowGraphBootstrapContext);
  return ctx?.flowGraphInit ?? null;
}
