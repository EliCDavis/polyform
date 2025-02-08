const panel = new GUI({ width: 310 });

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


// https://threejs.org/examples/?q=Directional#webgl_lights_hemisphere
// https://threejs.org/examples/#webgl_geometry_spline_editor
import * as GaussianSplats3D from '@mkkellogg/gaussian-splats-3d';

const container = document.getElementById('three-viewer-container');

import * as THREE from 'three';
import { NodeManager } from "./node_manager.js";
import { WebSocketManager, WebSocketRepresentationManager } from "./websocket.js";
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import Stats from 'three/addons/libs/stats.module.js';
import { GUI } from 'three/addons/libs/lil-gui.module.min.js';
import { CSS2DRenderer } from 'three/addons/renderers/CSS2DRenderer.js';
// import { RoomEnvironment } from 'three/addons/environments/RoomEnvironment.js';
import { ProgressiveLightMap } from 'three/addons/misc/ProgressiveLightMap.js';

import { EffectComposer } from 'three/addons/postprocessing/EffectComposer.js';
import { RenderPass } from 'three/addons/postprocessing/RenderPass.js';
import { UnrealBloomPass } from 'three/addons/postprocessing/UnrealBloomPass.js';
import { OutputPass } from 'three/addons/postprocessing/OutputPass.js';

import { InitXR } from './xr.js';
import { UpdateManager } from './update-manager.js';
import { getFileExtension } from './utils.js';
import { NoteManager } from './note_manager.js';

const viewportSettings = {
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

const shadowMapRes = 4098, lightMapRes = 4098, lightCount = 8;

const camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);
camera.position.set(0, 2, 3);

representationManager.AddRepresentation(0, camera)

const scene = new THREE.Scene();
scene.background = new THREE.Color(viewportSettings.background);

const textureLoader = new THREE.TextureLoader();
const textureEquirec = textureLoader.load('https://i.imgur.com/Ev4X4yY_d.webp?maxwidth=1520&fidelity=grand');
textureEquirec.mapping = THREE.EquirectangularReflectionMapping;
textureEquirec.colorSpace = THREE.SRGBColorSpace;

// scene.background = textureEquirec;
scene.fog = new THREE.Fog(viewportSettings.fog.color, viewportSettings.fog.near, viewportSettings.fog.far);

const viewerContainer = new THREE.Group();
scene.add(viewerContainer);

const threeCanvas = document.getElementById("three-canvas");
const renderer = new THREE.WebGLRenderer({
    canvas: threeCanvas,
    antialias: RenderingConfiguration.AntiAlias
});
renderer.setPixelRatio(window.devicePixelRatio);
renderer.setSize(threeCanvas.clientWidth, threeCanvas.clientHeight, false);
renderer.shadowMap.enabled = true;
renderer.shadowMap.type = THREE.PCFSoftShadowMap; // default THREE.PCFShadowMap
renderer.toneMapping = THREE.ACESFilmicToneMapping;
renderer.toneMappingExposure = 1;
renderer.xr.enabled = RenderingConfiguration.XrEnabled;
renderer.setAnimationLoop(updateLoop.run.bind(updateLoop))

const renderScene = new RenderPass(scene, camera);

// constructor( resolution, strength, radius, threshold )
const bloomPass = new UnrealBloomPass(new THREE.Vector2(window.innerWidth, window.innerHeight), .3, 0., 1.01);
// bloomPass.threshold = params.threshold;
// bloomPass.strength = params.strength;
// bloomPass.radius = params.radius;

const outputPass = new OutputPass();

const composer = new EffectComposer(renderer);
composer.addPass(renderScene);
composer.addPass(bloomPass);
composer.addPass(outputPass);

// container.appendChild(renderer.domElement);
// progressive lightmap
// const progressiveSurfacemap = new ProgressiveLightMap(renderer, lightMapRes);

const labelRenderer = new CSS2DRenderer();
labelRenderer.setSize(threeCanvas.clientWidth, threeCanvas.clientHeight, false);
labelRenderer.domElement.style.position = 'absolute';
labelRenderer.domElement.style.top = '0px';
labelRenderer.domElement.style.pointerEvents = 'none';
container.appendChild(labelRenderer.domElement);

const stats = new Stats();
container.appendChild(stats.dom);

const hemiLight = new THREE.HemisphereLight(viewportSettings.lighting, 0x8d8d8d, 1);
hemiLight.position.set(0, 20, 0);
scene.add(hemiLight);

const dirLight = new THREE.DirectionalLight(viewportSettings.lighting, 1);
dirLight.position.set(100, 100, 100);
dirLight.castShadow = true;
dirLight.shadow.camera.top = 100;
dirLight.shadow.camera.bottom = -100;
dirLight.shadow.camera.left = - 100;
dirLight.shadow.camera.right = 100;
// dirLight.shadow.camera.far = 40;
dirLight.shadow.camera.near = 0.1;
dirLight.shadow.mapSize.width = shadowMapRes;
dirLight.shadow.mapSize.height = shadowMapRes;
// progressiveSurfacemap.addObjectsToLightMap([dirLight])
scene.add(dirLight);


// ground
const groundMat = new THREE.MeshPhongMaterial({ color: viewportSettings.ground, depthWrite: true });
const groundMesh = new THREE.Mesh(new THREE.PlaneGeometry(1000, 1000), groundMat);
groundMesh.rotation.x = - Math.PI / 2;
groundMesh.receiveShadow = true;
scene.add(groundMesh);
// progressiveSurfacemap.addObjectsToLightMap([groundMesh])

// const environment = new RoomEnvironment(renderer);
// const pmremGenerator = new THREE.PMREMGenerator(renderer);
// scene.environment = pmremGenerator.fromScene( environment ).texture;

const orbitControls = new OrbitControls(camera, renderer.domElement);
// controls.addEventListener('change', render); // use if there is no animation loop
orbitControls.minDistance = 0;
orbitControls.maxDistance = 100;
orbitControls.target.set(0, 0, 0);
orbitControls.update();

camera.position.z = 5;

const requestManager = new RequestManager();

const App = {
    Camera: camera,
    Renderer: renderer,
    // MeshGenFolder: panel.addFolder("Mesh Generation"),
    Scene: scene,
    OrbitControls: orbitControls,
    ViewerScene: viewerContainer,
    NodeFlowGraph: nodeFlowGraph,
    RequestManager: requestManager,
    ServerUpdatingNodeConnections: false,
    SchemaRefreshManager: null,
}

const nodeManager = new NodeManager(App);
const noteManager = new NoteManager(requestManager, nodeFlowGraph)
const schemaManager = new SchemaManager(requestManager, nodeManager, noteManager);

if (RenderingConfiguration.XrEnabled) {
    InitXR(scene, renderer, updateLoop, representationManager, groundMesh);
}

nodeManager.subscribeToParameterChange((param) => {
    schemaManager.setParameter(param.id, param.data, param.binary);
});

let firstTimeLoadingScene = true;

const loader = new GLTFLoader().setPath('producer/value/');
let producerScene = null;

let guassianSplatViewer = null;


class SchemaRefreshManager {
    constructor() {
        this.loadingCount = 0;
        this.cachedSchema = null;
        this.subscribers = [];
    }

    Subscribe(callback) {
        this.subscribers.push(callback);
    }

    AddLoading() {
        this.loadingCount += 1;
    }

    RemoveLoading() {
        if (this.loadingCount === 0) {
            throw new Error("loading count already 0");
        }
        this.loadingCount -= 1;

        if (this.loadingCount === 0 && this.cachedSchema) {
            this.Refresh(this.cachedSchema)
            this.cachedSchema = null;
        }
    }

    CurrentlyLoading() {
        return this.loadingCount > 0;
    }

    NewSchema(schema) {
        if (this.CurrentlyLoading()) {
            this.cachedSchema = schema;
            return;
        }
        this.Refresh(schema);
    }

    loadText(producerURL) {
        this.AddLoading();
        requestManager.fetchText(
            producerURL,
            (data) => {
                InfoManager.ShowInfo(data);
                this.RemoveLoading();
                this.UpdateSubscribers(producerURL, data);
            },
            (error) => {
                this.RemoveLoading();
                console.error("unable to load text", producerURL, error);
                ErrorManager.ShowError(producerURL, JSON.parse(error).error);
            }
        );
    }

    loadImage(producerURL) {
        this.AddLoading();
        requestManager.fetchImage(
            producerURL,
            (data) => {
                this.RemoveLoading();
                this.UpdateSubscribers(producerURL, data);
            },
            (error) => {
                this.RemoveLoading();
                console.error("unable to load image", producerURL, error);
                ErrorManager.ShowError(producerURL, JSON.parse(error).error);
            }
        );
    }

    viewAABB(aabb) {

        const aabbDepth = (aabb.max.z - aabb.min.z)
        const aabbWidth = (aabb.max.x - aabb.min.x)
        const aabbHeight = (aabb.max.y - aabb.min.y)
        const aabbHalfHeight = aabbHeight / 2
        const mid = (aabb.max.y + aabb.min.y) / 2

        if (firstTimeLoadingScene && isFinite(aabbWidth) && isFinite(aabbDepth) && isFinite(aabbHeight)) {
            // console.log("Camera position intialized", aabbWidth, aabbDepth, aabbHeight);
            firstTimeLoadingScene = false;

            camera.position.y = (- mid + aabbHalfHeight) * (3 / 2);
            camera.position.z = Math.sqrt(
                (aabbWidth * aabbWidth) +
                (aabbDepth * aabbDepth) +
                (aabbHeight * aabbHeight)
            ) / 2;

            orbitControls.target.set(
                (aabb.max.x + aabb.min.x) / 2, 
                - mid + aabbHalfHeight, 
                (aabb.max.z + aabb.min.z) / 2
            );
            orbitControls.update();
        }
    }

    loadGltf(key, producerURL) {
        this.AddLoading();
        loader.load(producerURL, ((gltf) => {

            const aabb = new THREE.Box3();
            aabb.setFromObject(gltf.scene);
            const aabbHeight = (aabb.max.y - aabb.min.y)
            const aabbHalfHeight = aabbHeight / 2
            const mid = (aabb.max.y + aabb.min.y) / 2

            producerScene.add(gltf.scene);

            // We have to do this weird thing because the pivot of the scene
            // Isn't always the center of the AABB
            viewerContainer.position.set(0, - mid + aabbHalfHeight, 0)

            const objects = [];

            gltf.scene.traverse((object) => {
                if (object.isMesh) {
                    object.castShadow = true;
                    object.receiveShadow = true;
                    object.material.wireframe = viewportSettings.renderWireframe;
                    object.material.envMap = textureEquirec;
                    object.material.needsUpdate = true;
                    // object.material.transparent = true;

                    objects.push(object)
                } else if (object.isPoints) {
                    object.material.size = 2;
                }
            });

            // progressiveSurfacemap.addObjectsToLightMap(objects);

            this.viewAABB(aabb);

            this.UpdateSubscribers(producerURL, gltf);

            this.RemoveLoading();
        }).bind(this),
            undefined,
            (error) => {
                this.RemoveLoading();
                error.response.json().then(x => {
                    ErrorManager.ShowError(key, x.error);
                })
            });
    }

    loadSplat(key, producerURL) {
        this.AddLoading();
        if (guassianSplatViewer) {
            guassianSplatViewer.dispose();
        }

        renderer.setPixelRatio(1);

        // https://github.com/mkkellogg/GaussianSplats3D/blob/main/src/Viewer.js
        const splatViewerOptions = {
            selfDrivenMode: false,
            // 'cameraUp': [0, -1, 0],
            sphericalHarmonicsDegree: 2,
            useBuiltInControls: false,
            rootElement: renderer.domElement.parentElement,
            renderer: renderer,
            threeScene: scene,
            camera: camera,
            gpuAcceleratedSort: true,
            // 'sceneRevealMode': GaussianSplats3D.SceneRevealMode.Instant,
            sharedMemoryForWebWorkers: true
        }

        guassianSplatViewer = new GaussianSplats3D.Viewer(splatViewerOptions);

        // getSplatCenter
        guassianSplatViewer.addSplatScene(producerURL, {
            // rotation: [1, 0, 0, 0],
            // scale: [-1, -1, 1, 0],
            // streamView: false
            // 'scale': [0.25, 0.25, 0.25],
        }).then((() => {

            guassianSplatViewer.splatMesh.onSplatTreeReady((splatTree) => {
                const tree = splatTree.subTrees[0]
                const aabb = new THREE.Box3();
                aabb.setFromPoints([tree.sceneMin, tree.sceneMax]);
                const aabbHeight = (aabb.max.y - aabb.min.y)
                const aabbHalfHeight = aabbHeight / 2
                const mid = (aabb.max.y + aabb.min.y) / 2

                const shiftY = - mid + aabbHalfHeight
                guassianSplatViewer.splatMesh.position.set(0, shiftY, 0)
                viewerContainer.position.set(0, shiftY, 0);

                this.viewAABB(aabb);
            });

            this.RemoveLoading();
            this.UpdateSubscribers(producerURL, guassianSplatViewer.splatMesh);

        }).bind(this)).catch(x => {
            console.error(x)
            this.RemoveLoading();
            ErrorManager.ShowError(key, x.error);
        })
    }

    Refresh(schema) {
        InfoManager.ClearInfo();

        if (producerScene != null) {
            viewerContainer.remove(producerScene)
        }

        producerScene = new THREE.Group();
        viewerContainer.add(producerScene);

        for (const [producer, producerData] of Object.entries(schema.producers)) {
            ErrorManager.ClearError(producer);
            const fileExt = getFileExtension(producer);

            switch (fileExt) {
                case "txt":
                    this.loadText('producer/value/' + producer);
                    break;

                case "gltf":
                case "glb":
                    this.loadGltf(producer, producer);
                    break;

                case "splat":
                    this.loadSplat(producer, 'producer/value/' + producer)
                    break;

                case "ply":
                    this.loadSplat(producer, 'producer/value/' + producer)
                    break;

                case "png":
                    this.loadImage('producer/value/' + producer);
                    break;
            }
        }
    }

    UpdateSubscribers(url, thing) {
        this.subscribers.forEach(sub => {
            if (sub == null) {
                return;
            }
            sub(url, thing);
        })
    }
}

const schemaRefreshManager = new SchemaRefreshManager();
App.SchemaRefreshManager = schemaRefreshManager;
schemaManager.subscribe(schemaRefreshManager.NewSchema.bind(schemaRefreshManager));



const fileControls = {
    newGraph: () => {
        document.getElementById("new-graph-popup").style.display = "flex"; 
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
    loadProfile: () => {
        const input = document.createElement('input');
        input.type = 'file';

        input.onchange = e => {

            // getting a hold of the file reference
            const file = e.target.files[0];

            // setting up the reader
            const reader = new FileReader();
            reader.readAsText(file, 'UTF-8');

            // here we tell the reader what to do when it's done reading...
            reader.onload = readerEvent => {
                const content = readerEvent.target.result; // this is the content!
                requestManager.setGraph(JSON.parse(content), (_) => {
                    location.reload();
                })
            }

        }

        input.click();
    },
    saveModel: () => {
        download("/zip", (data) => {
            const a = document.createElement('a');
            a.download = 'model.zip';
            const url = window.URL.createObjectURL(data);
            a.href = url;
            a.click();
            window.URL.revokeObjectURL(url);
        })
    },
    viewProgram: () => {
        requestManager.fetchText("/mermaid", (data) => {
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

const fileSettingsFolder = panel.addFolder("Graph");
fileSettingsFolder.add(fileControls, "newGraph").name("New")
fileSettingsFolder.add(fileControls, "saveGraph").name("Save")
fileSettingsFolder.add(fileControls, "loadProfile").name("Load")

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
            if (producerScene == null) {
                return;
            }
            producerScene.traverse((object) => {
                if (object.isMesh) {
                    object.material.wireframe = viewportSettings.renderWireframe;
                }
            });
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
            scene.background = new THREE.Color(viewportSettings.background);
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
            dirLight.color = new THREE.Color(viewportSettings.lighting);
            hemiLight.color = new THREE.Color(viewportSettings.lighting);
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
            groundMat.color = new THREE.Color(viewportSettings.ground);
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
            scene.fog.color = new THREE.Color(viewportSettings.fog.color);
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
            scene.fog.near = viewportSettings.fog.near;
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
            scene.fog.far = viewportSettings.fog.far;
        }
    )
);


function resize(force) {
    const w = renderer.domElement.clientWidth;
    const h = renderer.domElement.clientHeight

    if (renderer.domElement.width !== w || renderer.domElement.height !== h || force) {
        renderer.setSize(w, h, false);
        composer.setSize(w, h);
        camera.aspect = w / h;
        camera.updateProjectionMatrix();
        // nodeCanvas.resize(nodeCanvas.clientWidth, nodeCanvas.clientHeight, false)
        labelRenderer.setSize(w, h);
    }
}

const websocketManager = new WebSocketManager(
    representationManager,
    scene,
    {
        playerGeometry: new THREE.SphereGeometry(1, 32, 16),
        playerMaterial: new THREE.MeshPhongMaterial({ color: 0xffff00 }),
        playerEyeMaterial: new THREE.MeshBasicMaterial({ color: 0x000000 }),
    },
    viewportManager,
    schemaManager
);
if (websocketManager.canConnect()) {
    websocketManager.connect();
    updateLoop.addToUpdate(websocketManager.update.bind(websocketManager));
} else {
    console.error("web browser does not support web sockets")
}

updateLoop.addToUpdate(() => {
    resize(false);

    composer.render(scene, camera);

    if (guassianSplatViewer) {
        guassianSplatViewer.update();
        guassianSplatViewer.render();
    }

    labelRenderer.render(scene, camera);
    stats.update();
});

