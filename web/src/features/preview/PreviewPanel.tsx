import { useEffect, useRef } from "react";
import { useSchema } from "@/api/hooks";
import { PreviewViewer } from "./preview_viewer";
import styles from "./PreviewPanel.module.css";

export function nodeOutputUrl(nodeId: string, port: string, scope: string | null): string {
  const base = `./node/output/${nodeId}/${port}`;
  return scope ? `${base}?scope=${encodeURIComponent(scope)}` : base;
}

export function PreviewPanel({ nodeId, port, scope }: { nodeId: string; port: string; scope: string | null }) {
  const { data: graph } = useSchema();
  const nodes = scope ? graph?.subGraphs?.[scope.replace(/^subgraph\//, "")]?.nodes : graph?.nodes;
  const version = nodes?.[nodeId]?.output?.[port]?.version ?? 0;
  const container = useRef<HTMLDivElement>(null);
  const viewer = useRef<PreviewViewer | null>(null);

  useEffect(() => {
    const created = new PreviewViewer(container.current!);
    viewer.current = created;
    return () => {
      created.dispose();
      viewer.current = null;
    };
  }, []);

  useEffect(() => {
    const url = nodeOutputUrl(nodeId, port, scope);
    viewer.current?.load(`${url}${url.includes("?") ? "&" : "?"}v=${version}`);
  }, [nodeId, port, scope, version]);

  return <div ref={container} className={styles.viewer} />;
}
