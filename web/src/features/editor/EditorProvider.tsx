import { useEffect, useRef, useState, type ReactNode } from "react";
import Stats from "three/examples/jsm/libs/stats.module.js";
import { MeshBasicMaterial, MeshPhongMaterial, SphereGeometry } from "three";
import { NodeManager } from "@/lib/node_manager";
import { WebSocketManager, WebSocketRepresentationManager } from "@/lib/websocket";
import { XRManager } from "@/lib/xr";
import { UpdateManager } from "@/lib/update_manager";
import { NoteManager } from "@/lib/note_manager";
import { ViewportManager } from "@/lib/viewport_manager";
import { SchemaManager } from "@/lib/schema_manager";
import { ProducerViewManager } from "@/lib/ProducerView/producer_view_manager";
import { CreateThreeApp, type ThreeApp } from "@/lib/three_app";
import type { ViewportSettings } from "@/lib/viewport_settings";
import { requestManager } from "@/api/client";
import { getRenderingConfiguration } from "@/config";
import { useNodeTypes, useStartedPolling } from "@/api/hooks";
import { useUiStore } from "@/stores/uiStore";
import { EditorContext, type EditorContextValue, useEditorOptional } from "./EditorContext";
import type { RegisteredTypes } from "@/types/schema";
import {
  FlowGraphBootstrapProvider,
  useFlowGraphInit,
} from "@/features/nodeFlow/FlowGraphBootstrapContext";
import { useGraphTabStore, activeGraphScope } from "@/stores/graphTabStore";

const viewportSettings: ViewportSettings = {
  renderWireframe: false,
  fog: { color: "0xa0a0a0", near: 100, far: 150 },
  background: "0xa0a0a0",
  lighting: "0xffffff",
  ground: "0xcbcbcb",
};

function EditorModelVersionPoller() {
  const editor = useEditorOptional();
  useStartedPolling((modelVersion) => {
    editor?.producerViewManager.setModelVersion(modelVersion);
  });
  return null;
}

function GraphTabScopeSync({
  nodeManager,
  noteManager,
  schemaManager,
}: {
  nodeManager: NodeManager;
  noteManager: NoteManager;
  schemaManager: SchemaManager;
}) {
  const activeTabId = useGraphTabStore((s) => s.activeTabId);

  useEffect(() => {
    if (!schemaManager.currentGraph) return;
    const scope = activeGraphScope(activeTabId);
    nodeManager.switchGraphScope(scope, schemaManager.currentGraph);
    noteManager.switchGraphScope(scope, schemaManager.currentGraph);
  }, [activeTabId, nodeManager, noteManager, schemaManager]);

  return null;
}

function EditorBootstrap({
  registeredTypes,
  onReady,
}: {
  registeredTypes: RegisteredTypes;
  onReady: (ctx: EditorContextValue) => void;
}) {
  const flowGraphInit = useFlowGraphInit();
  const hideStats = useUiStore((s) => s.hideStats);
  const initialized = useRef(false);

  useEffect(() => {
    if (!flowGraphInit || initialized.current) return;

    let cancelled = false;

    const tryInit = () => {
      if (cancelled || initialized.current) return;

      const container = document.getElementById("three-viewer-container");
      const threeCanvas = document.getElementById("three-canvas");
      if (!container || !threeCanvas) {
        requestAnimationFrame(tryInit);
        return;
      }

      initialized.current = true;
      const renderingConfig = getRenderingConfiguration();
      const updateLoop = new UpdateManager();

      const threeApp: ThreeApp = CreateThreeApp(
        container,
        viewportSettings,
        updateLoop,
        renderingConfig.AntiAlias,
        renderingConfig.XrEnabled
      );

      if (!hideStats) {
        const stats = new Stats();
        stats.dom.style.left = "unset";
        stats.dom.style.right = "0";
        container.appendChild(stats.dom);
        updateLoop.addToUpdate({
          name: "Stats",
          loop: () => stats.update(),
        });
      }

      const schemaManager = new SchemaManager(requestManager);
      const producerViewManager = new ProducerViewManager(
        threeApp,
        requestManager,
        registeredTypes.nodeTypes,
        schemaManager
      );
      const noteManager = new NoteManager(requestManager, flowGraphInit.NodeFlowGraph);
      const nodeManager = new NodeManager(
        flowGraphInit.NodeFlowGraph,
        requestManager,
        flowGraphInit.PolyformNodesPublisher,
        threeApp,
        producerViewManager,
        registeredTypes
      );

      producerViewManager.SubscribeToCompleteRefresh(() => {
        nodeManager.refreshExecutionReport();
      });

      nodeManager.subscribeToParameterChange((param) => {
        schemaManager.setParameter(param.id, param.data, param.binary);
      });

      nodeManager.setOnSchemaRefreshNeeded(() => {
        schemaManager.refreshSchema("sub-graph definition changed");
      });

      schemaManager.subscribe((g) => {
        producerViewManager.NewSchema(g);
        nodeManager.updateNodes(g);
        noteManager.schemaUpdate(g);
      });

      schemaManager.refreshSchema("initial load");

      const viewportManager = new ViewportManager(viewportSettings);
      const representationManager = new WebSocketRepresentationManager();
      representationManager.AddRepresentation(0, threeApp.Camera);

      if (renderingConfig.XrEnabled) {
        new XRManager(threeApp, representationManager, updateLoop);
      }

      const websocketManager = new WebSocketManager(
        representationManager,
        threeApp.Scene,
        {
          playerGeometry: new SphereGeometry(1, 32, 16),
          playerMaterial: new MeshPhongMaterial({ color: 0xffff00 }),
          playerEyeMaterial: new MeshBasicMaterial({ color: 0x000000 }),
        },
        viewportManager,
        producerViewManager
      );

      if (websocketManager.canConnect()) {
        websocketManager.connect();
        updateLoop.addToUpdate({
          name: "Websocket",
          loop: websocketManager.update.bind(websocketManager),
        });
      }

      function resize() {
        const renderer = threeApp.Renderer;
        const rect = renderer.domElement.getBoundingClientRect();
        const w = Math.floor(rect.width);
        const h = Math.floor(rect.height);
        if (renderer.domElement.width !== w || renderer.domElement.height !== h) {
          threeApp.Camera.aspect = w / h;
          threeApp.Camera.updateProjectionMatrix();
          renderer.setSize(w, h, false);
          threeApp.Composer.setSize(w, h);
          threeApp.LabelRenderer.setSize(w, h);
        }
      }

      updateLoop.addToUpdate({
        name: "Rendering",
        loop: (delta) => {
          resize();
          threeApp.OrbitControls.update();
          threeApp.Composer.render(delta);
          producerViewManager.Render();
          threeApp.LabelRenderer.render(threeApp.Scene, threeApp.Camera);
        },
      });

      const value: EditorContextValue = {
        schemaManager,
        nodeManager,
        noteManager,
        producerViewManager,
        requestManager,
        threeApp,
        nodeFlowGraph: flowGraphInit.NodeFlowGraph,
        registeredTypes,
        ready: true,
      };
      onReady(value);
    };

    tryInit();

    return () => {
      cancelled = true;
    };
  }, [registeredTypes, hideStats, onReady, flowGraphInit]);

  return null;
}

export function EditorProvider({ children }: { children: ReactNode }) {
  const { data: registeredTypes } = useNodeTypes();
  const [ctx, setCtx] = useState<EditorContextValue | null>(null);

  return (
    <FlowGraphBootstrapProvider>
      <EditorContext.Provider value={ctx}>
        {children}
        {ctx && (
          <>
            <GraphTabScopeSync
              nodeManager={ctx.nodeManager}
              noteManager={ctx.noteManager}
              schemaManager={ctx.schemaManager}
            />
            <EditorModelVersionPoller />
          </>
        )}
        {registeredTypes && !ctx && (
          <EditorBootstrap registeredTypes={registeredTypes} onReady={setCtx} />
        )}
      </EditorContext.Provider>
    </FlowGraphBootstrapProvider>
  );
}
