import type { ImportSubGraphsResult } from "@/lib/schema";

export interface RenderingConfiguration {
  AntiAlias: boolean;
  XrEnabled: boolean;
  ShowNewGraphPopup: boolean;
  IsWasm?: boolean;
}

declare global {
  interface Window {
    RenderingConfiguration: RenderingConfiguration;
    ExampleGraphs: string[];
    loadGraph: (content: unknown) => void;
    getGraph: (cb: (graph: unknown) => void) => void;
    graphChangeCallback: (cb: (e: string) => void) => void;
    importSubGraphs: (
      content: unknown,
      callback?: (result: ImportSubGraphsResult) => void,
      error?: (err: unknown) => void
    ) => void;
  }
}

export function getRenderingConfiguration(): RenderingConfiguration {
  return globalThis.RenderingConfiguration ?? {
    AntiAlias: true,
    XrEnabled: false,
    ShowNewGraphPopup: false,
  };
}

export function getExampleGraphs(): string[] {
  return globalThis.ExampleGraphs ?? [];
}

export function isWasmDeployment(): boolean {
  const config = getRenderingConfiguration();
  if (config.IsWasm !== undefined) return config.IsWasm;
  return location.pathname.endsWith("app.html");
}
