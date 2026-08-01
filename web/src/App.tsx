import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { EditorProvider } from "@/features/editor/EditorProvider";
import { AppShell } from "@/features/layout/AppShell";
import { useEffect } from "react";
import { useInvalidateSchema } from "@/api/hooks";
import { requestManager } from "@/api/client";
import { applyImportedSubGraphs } from "@/lib/importSubGraphs";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function AppInner() {
  const invalidate = useInvalidateSchema();
  useEffect(() => {
    requestManager.subscribeToGraphChange(() => invalidate());
  }, [invalidate]);
  return (
    <EditorProvider
      onEditorReady={(nodeManager) => {
        window.importSubGraphs = (content, callback, error) => {
          requestManager.importSubGraphs(
            content,
            (result) => {
              applyImportedSubGraphs(nodeManager, result);
              callback?.(result);
            },
            (err) => error?.(err)
          );
        };
      }}
    >
      <AppShell />
    </EditorProvider>
  );
}

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AppInner />
    </QueryClientProvider>
  );
}

export function setupGlobalHooks() {
  window.loadGraph = (content) => {
    requestManager.setGraph(content, () => location.reload());
  };
  window.getGraph = (cb) => {
    requestManager.getGraph(cb);
  };
  window.graphChangeCallback = (cb) => {
    requestManager.subscribeToGraphChange(cb);
  };
  // Overwritten by EditorProvider once NodeManager is ready so imported
  // runtime types can be registered without a full page reload.
  window.importSubGraphs = (_content, _callback, error) => {
    error?.(new Error("Editor is not ready yet; try again after the graph UI loads"));
  };
}
