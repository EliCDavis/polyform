import { NodeManager } from "./node_manager.js";
import { WebSocketManager, WebSocketRepresentationManager } from "./websocket.js";
import Stats from 'three/examples/jsm/libs/stats.module.js';
import { GUI } from 'three/examples/jsm/libs/lil-gui.module.min.js';

import { XRManager } from './xr.js';
import { UpdateManager } from './update_manager.js';
import { NoteManager } from './note_manager.js';
import { MeshBasicMaterial, MeshPhongMaterial, SphereGeometry } from 'three';
import { ViewportManager } from './viewport_manager.js';
import { SchemaManager } from './schema_manager.js';
import { ProducerViewManager } from './ProducerView/producer_view_manager.js';
import { downloadBlob, RequestManager } from './requests.js';
import { CreateNodeFlowGraph } from "./flow_graph.js";
import { CreateThreeApp, ThreeApp } from "./three_app.js";
import { ViewportSettings } from "./viewport_settings.js";
import { NewGraphPopup } from './popups/new_graph.js';
import { BuildFogSettings, BuildRenderingSetting } from "./gui_settings/fog.js";
import { ArrayBufferToBase64, Compress, CopyToClipboard } from "./utils.js";
import { VariableManager } from "./variables/variable_manager.js";


const graphPopup = new NewGraphPopup(globalThis.ExampleGraphs);

const RenderingConfiguration = {
    AntiAlias: globalThis.RenderingConfiguration.AntiAlias,
    XrEnabled: globalThis.RenderingConfiguration.XrEnabled
}

const viewportSettings: ViewportSettings = {
    renderWireframe: false,
    fog: {
        color: "0xa0a0a0",
        near: 10,
        far: 50,
    },
    background: "0xa0a0a0",
    lighting: "0xffffff",
    ground: "0xcbcbcb"
}

const updateLoop = new UpdateManager();
const container = document.getElementById('three-viewer-container');
const threeApp: ThreeApp = CreateThreeApp(
    container,
    viewportSettings,
    updateLoop,
    RenderingConfiguration.AntiAlias,
    RenderingConfiguration.XrEnabled
);

const stats = new Stats();
stats.dom.style.left = "unset";
stats.dom.style.right = "0";
container.appendChild(stats.dom);

const flowGraphStuff = CreateNodeFlowGraph();
const requestManager = new RequestManager();

requestManager.getNodeTypes((nodeTypes) => {
    const producerViewManager = new ProducerViewManager(threeApp, requestManager, nodeTypes);

    let initID = null
    setInterval(() => {
        requestManager.getStartedTime((payload) => {
            if (initID === null) {
                initID = payload.time;
            }

            if (initID !== payload.time) {
                location.reload();
            }
            schemaManager.setModelVersion(payload.modelVersion)
        })
    }, 1000);

    console.log(nodeTypes);
    
    const noteManager = new NoteManager(requestManager, flowGraphStuff.NodeFlowGraph)

    const nodeManager = new NodeManager(
        flowGraphStuff.NodeFlowGraph,
        requestManager,
        flowGraphStuff.PolyformNodesPublisher,
        threeApp,
        producerViewManager,
        nodeTypes
    );
    const schemaManager = new SchemaManager(requestManager, nodeManager, noteManager, graphPopup);
    new VariableManager(document.getElementById("sidebar-content"), schemaManager, nodeManager, flowGraphStuff.PolyformNodesPublisher, threeApp);

    nodeManager.subscribeToParameterChange((param) => {
        schemaManager.setParameter(param.id, param.data, param.binary);
    });

    schemaManager.subscribe(producerViewManager.NewSchema.bind(producerViewManager));

    const fileControls = {
        newGraph: () => {
            graphPopup.show();
        },
        saveSwagger: () => {
            requestManager.getSwagger((swagger) => {
                const fileContent = JSON.stringify(swagger);
                const bb = new Blob([fileContent], { type: 'application/json' });
                const a = document.createElement('a');
                a.download = 'swagger.json';
                a.href = window.URL.createObjectURL(bb);
                a.click();
            })
        },
        saveGraph: () => {
            requestManager.getGraph((graph) => {
                const fileContent = JSON.stringify(graph);
                const bb = new Blob([fileContent], { type: 'application/json' });
                const a = document.createElement('a');
                a.download = 'graph.json';
                a.href = window.URL.createObjectURL(bb);
                a.click();
            })
        },
        link: () => {
            requestManager.getGraph(async (graph) => {
                const fileContent = JSON.stringify(graph);
                const compressedGraph = await Compress(fileContent);
                const compressedString = await ArrayBufferToBase64(compressedGraph)
                const url = window.location.href + "q?graph=" + compressedString
                CopyToClipboard(url);
            })
        },
        loadGraph: () => {
            const input = document.createElement('input');
            input.type = 'file';

            input.onchange = e => {

                // getting a hold of the file reference
                const file = (e.target as HTMLInputElement).files[0];

                // setting up the reader
                const reader = new FileReader();
                reader.readAsText(file, 'UTF-8');

                // here we tell the reader what to do when it's done reading...
                reader.onload = readerEvent => {
                    const content = readerEvent.target.result as string; // this is the content!
                    requestManager.setGraph(JSON.parse(content), (_) => {
                        location.reload();
                    })
                }
            }

            input.click();
        },
        saveModel: () => {
            downloadBlob("./zip/", (data) => {
                const a = document.createElement('a');
                a.download = 'model.zip';
                const url = window.URL.createObjectURL(data);
                a.href = url;
                a.click();
                window.URL.revokeObjectURL(url);
            })
        },
        viewProgram: () => {
            requestManager.fetchText("./mermaid", (data) => {
                const mermaidConfig = {
                    "code": data,
                    "mermaid": {
                        "securityLevel": "strict"
                    }
                }

                const mermaidURL = "https://mermaid.live/edit#" + btoa(JSON.stringify(mermaidConfig));
                window.open(mermaidURL, '_blank').focus();
            })
        }
    }

    document.getElementById("new-graph-button").onclick = fileControls.newGraph;
    document.getElementById("save-graph-button").onclick = fileControls.saveGraph;
    document.getElementById("load-graph-button").onclick = fileControls.loadGraph;
    document.getElementById("export-model-button").onclick = fileControls.saveModel;
    document.getElementById("export-mermaid-button").onclick = fileControls.viewProgram;
    document.getElementById("export-swagger-button").onclick = fileControls.saveSwagger;

    // const panel = new GUI({ width: 310 });

    // const fileSettingsFolder = panel.addFolder("Graph");
    // fileSettingsFolder.add(fileControls, "newGraph").name("New")
    // fileSettingsFolder.add(fileControls, "saveGraph").name("Save")
    // fileSettingsFolder.add(fileControls, "loadGraph").name("Load");

    // // Graphs when compressed still make for a giant URL
    // // fileSettingsFolder.add(fileControls, "link").name("Get Link");

    // const exportSettingsFolder = panel.addFolder("Export");
    // exportSettingsFolder.add(fileControls, "saveModel").name("Model")
    // exportSettingsFolder.add(fileControls, "viewProgram").name("Mermaid")
    // exportSettingsFolder.add(fileControls, "saveSwagger").name("Swagger 2.0")
    // exportSettingsFolder.close();

    // const viewportSettingsFolder = panel.addFolder("Rendering");
    // viewportSettingsFolder.close();

    const viewportManager = new ViewportManager(viewportSettings);

    // BuildRenderingSetting(
    //     viewportSettingsFolder,
    //     viewportManager,
    //     viewportSettings,
    //     threeApp,
    //     producerViewManager
    // )

    // BuildFogSettings(
    //     viewportSettingsFolder,
    //     viewportManager,
    //     viewportSettings,
    //     threeApp
    // )

    const doWebsocketStuff = true;
    if (doWebsocketStuff) {
        const representationManager = new WebSocketRepresentationManager();
        representationManager.AddRepresentation(0, threeApp.Camera)

        if (RenderingConfiguration.XrEnabled) {
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
            schemaManager
        );

        if (websocketManager.canConnect()) {
            websocketManager.connect();
            updateLoop.addToUpdate({
                name: "Websocket",
                loop: websocketManager.update.bind(websocketManager)
            });
        } else {
            console.error("web browser does not support web sockets")
        }
    }


    function resize(force: boolean) {
        const renderer = threeApp.Renderer;
        const w = renderer.domElement.clientWidth;
        const h = renderer.domElement.clientHeight

        if (renderer.domElement.width !== w || renderer.domElement.height !== h || force) {
            renderer.setSize(w, h, false);
            threeApp.Composer.setSize(w, h);
            threeApp.Camera.aspect = w / h;
            threeApp.Camera.updateProjectionMatrix();
            // nodeCanvas.resize(nodeCanvas.clientWidth, nodeCanvas.clientHeight, false)
            threeApp.LabelRenderer.setSize(w, h);
        }
    }

    updateLoop.addToUpdate({
        name: "Rendering",
        loop: (delta) => {
            resize(false);

            threeApp.Composer.render(delta);
            producerViewManager.Render();
            threeApp.LabelRenderer.render(threeApp.Scene, threeApp.Camera);

            stats.update();
        }
    });
})
