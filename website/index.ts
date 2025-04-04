import { NodeManager } from "./node_manager.js";
import { WebSocketManager, WebSocketRepresentationManager } from "./websocket.js";
import Stats from 'three/examples/jsm/libs/stats.module.js';
import { GUI } from 'three/examples/jsm/libs/lil-gui.module.min.js';

import { XRManager } from './xr.js';
import { UpdateManager } from './update_manager.js';
import { NoteManager } from './note_manager.js';
import { Color, MeshBasicMaterial, MeshPhongMaterial, SphereGeometry } from 'three';
import { ViewportManager, ViewportSetting } from './viewport_manager.js';
import { SchemaManager } from './schema_manager.js';
import { ProducerViewManager } from './ProducerView/producer_view_manager.js';
import { downloadBlob, RequestManager } from './requests.js';
import { CreateNodeFlowGraph } from "./flow_graph.js";
import { CreateThreeApp, ThreeApp } from "./three_app.js";
import { ViewportSettings } from "./viewport_settings.js";
import { NewGraphPopup } from './popups/new_graph.js';

const graphPopup = new NewGraphPopup(globalThis.ExampleGraphs);

const RenderingConfiguration = {
    AntiAlias: globalThis.RenderingConfiguration.AntiAlias,
    XrEnabled: globalThis.RenderingConfiguration.XrEnabled
}

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

const representationManager = new WebSocketRepresentationManager();
const viewportManager = new ViewportManager(viewportSettings);
const updateLoop = new UpdateManager();

const requestManager = new RequestManager();

const flowGraphStuff = CreateNodeFlowGraph();

const container = document.getElementById('three-viewer-container');
const threeApp: ThreeApp = CreateThreeApp(
    container,
    viewportSettings,
    updateLoop,
    RenderingConfiguration.AntiAlias,
    RenderingConfiguration.XrEnabled
);
representationManager.AddRepresentation(0, threeApp.Camera)

const stats = new Stats();
container.appendChild(stats.dom);

const producerViewManager = new ProducerViewManager(
    threeApp,
    requestManager,
);

const nodeManager = new NodeManager(
    flowGraphStuff.NodeFlowGraph,
    requestManager,
    flowGraphStuff.PolyformNodesPublisher,
    threeApp,
    producerViewManager
);
const noteManager = new NoteManager(requestManager, flowGraphStuff.NodeFlowGraph)
const schemaManager = new SchemaManager(requestManager, nodeManager, noteManager, graphPopup);

if (RenderingConfiguration.XrEnabled) {
    new XRManager(threeApp, representationManager, updateLoop);
}

nodeManager.subscribeToParameterChange((param) => {
    schemaManager.setParameter(param.id, param.data, param.binary);
});

schemaManager.subscribe(producerViewManager.NewSchema.bind(producerViewManager));

const Compress = async (str: string, encoding = 'gzip' as CompressionFormat): Promise<ArrayBuffer> => {
    const byteArray = new TextEncoder().encode(str)
    const cs = new CompressionStream(encoding)
    const writer = cs.writable.getWriter()
    writer.write(byteArray)
    writer.close()
    return new Response(cs.readable).arrayBuffer()
}

const Decompress = async (byteArray: BufferSource, encoding = 'gzip' as CompressionFormat): Promise<string> => {
    const cs = new DecompressionStream(encoding)
    const writer = cs.writable.getWriter()
    writer.write(byteArray)
    writer.close()
    const arrayBuffer = await new Response(cs.readable).arrayBuffer()
    return new TextDecoder().decode(arrayBuffer)
}

async function CopyToClipboard(text: string) {
    try {
        await navigator.clipboard.writeText(text);
        console.log('Text copied to clipboard');
    } catch (err) {
        console.error('Failed to copy text: ', err);
    }
}

function ArrayBufferToBase64(buffer: ArrayBuffer) {
    return new Promise((resolve, reject) => {
        let blob = new Blob([buffer]);
        let reader = new FileReader();
        reader.onloadend = () => resolve(reader.result);
        reader.onerror = reject;
        reader.readAsDataURL(blob);
    });
}


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
    loadProfile: () => {
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
        downloadBlob("./zip", (data) => {
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

const panel = new GUI({ width: 310 });

const fileSettingsFolder = panel.addFolder("Graph");
fileSettingsFolder.add(fileControls, "newGraph").name("New")
fileSettingsFolder.add(fileControls, "saveGraph").name("Save")
fileSettingsFolder.add(fileControls, "loadProfile").name("Load");

// Graphs when compressed still make for a giant URL
// fileSettingsFolder.add(fileControls, "link").name("Get Link");

const exportSettingsFolder = panel.addFolder("Export");
exportSettingsFolder.add(fileControls, "saveModel").name("Model")
exportSettingsFolder.add(fileControls, "viewProgram").name("Mermaid")
exportSettingsFolder.add(fileControls, "saveSwagger").name("Swagger 2.0")
exportSettingsFolder.close();

const viewportSettingsFolder = panel.addFolder("Rendering");
viewportSettingsFolder.close();

viewportManager.AddSetting(
    "renderWireframe",
    new ViewportSetting(
        "renderWireframe",
        viewportSettings,
        viewportSettingsFolder
            .add(viewportSettings, "renderWireframe")
            .name("Render Wireframe"),
        () => {
            producerViewManager.SetWireframe(viewportSettings.renderWireframe);
        }
    )
)

viewportManager.AddSetting(
    "background",
    new ViewportSetting(
        "background",
        viewportSettings,
        viewportSettingsFolder
            .addColor(viewportSettings, "background")
            .name("Background"),
        () => {
            threeApp.Scene.background = new Color(viewportSettings.background);
        }
    )
);

viewportManager.AddSetting(
    "lighting",
    new ViewportSetting(
        "lighting",
        viewportSettings,
        viewportSettingsFolder
            .addColor(viewportSettings, "lighting")
            .name("Lighting"),
        () => {
            threeApp.Lighting.DirLight.color = new Color(viewportSettings.lighting);
            threeApp.Lighting.HemiLight.color = new Color(viewportSettings.lighting);
        },
    )
);

viewportManager.AddSetting(
    "ground",
    new ViewportSetting(
        "ground",
        viewportSettings,
        viewportSettingsFolder
            .addColor(viewportSettings, "ground")
            .name("Ground"),
        () => {
            threeApp.Ground.Material.color = new Color(viewportSettings.ground);
        }
    )
);

const fogSettingsFolder = viewportSettingsFolder.addFolder("Fog");
fogSettingsFolder.close();

viewportManager.AddSetting(
    "fog/color",
    new ViewportSetting(
        "color",
        viewportSettings.fog,
        fogSettingsFolder.addColor(viewportSettings.fog, "color"),
        () => {
            threeApp.Fog.color = new Color(viewportSettings.fog.color);
        }
    )
);

viewportManager.AddSetting(
    "fog/near",
    new ViewportSetting(
        "near",
        viewportSettings.fog,
        fogSettingsFolder.add(viewportSettings.fog, "near"),
        () => {
            threeApp.Fog.near = viewportSettings.fog.near;
        }
    )
);

viewportManager.AddSetting(
    "fog/far",
    new ViewportSetting(
        "far",
        viewportSettings.fog,
        fogSettingsFolder.add(viewportSettings.fog, "far"),
        () => {
            threeApp.Fog.far = viewportSettings.fog.far;
        }
    )
);


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