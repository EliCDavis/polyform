import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { EditorProvider } from "@/features/editor/EditorProvider";
import { AppShell } from "@/features/layout/AppShell";
import { useEffect } from "react";
import { useInvalidateSchema } from "@/api/hooks";
import { requestManager } from "@/api/client";

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
    <EditorProvider>
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
}
