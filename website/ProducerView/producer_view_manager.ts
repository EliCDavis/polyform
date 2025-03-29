import { Box3, EquirectangularReflectionMapping, Group, Mesh, PerspectiveCamera, Scene, SRGBColorSpace, TextureLoader, WebGLRenderer } from "three";
import { ErrorManager } from "../error_manager";
import { InfoManager } from "../info_manager";
import { GraphInstance } from "../schema";
import { getFileExtension } from "../utils";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader.js";
import { RequestManager } from "../requests";
import * as GaussianSplats3D from '@mkkellogg/gaussian-splats-3d';
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import { ThreeApp } from "../three_app";

const loader = new GLTFLoader().setPath('producer/value/');

type ProducerRefreshCallback = (string: string, thing: any) => void;


const textureLoader = new TextureLoader();
const textureEquirec = textureLoader.load('https://i.imgur.com/Ev4X4yY_d.webp?maxwidth=1520&fidelity=grand');
textureEquirec.mapping = EquirectangularReflectionMapping;
textureEquirec.colorSpace = SRGBColorSpace;


export class ProducerViewManager {

    loadingCount: number;

    wireframe: boolean;

    subscribers: Array<ProducerRefreshCallback>;

    cachedSchema: GraphInstance;

    firstTimeLoadingScene: boolean;

    renderer: WebGLRenderer;

    requestManager: RequestManager;

    camera: PerspectiveCamera;

    guassianSplatViewer: GaussianSplats3D.Viewer;

    producerScene: Group;

    orbitControls: OrbitControls;

    viewerContainer: Group;

    scene: Scene;

    constructor(
        app: ThreeApp,
        requestManager: RequestManager,
    ) {
        this.requestManager = requestManager;
        this.renderer = app.Renderer;
        this.camera = app.Camera;
        this.orbitControls = app.OrbitControls;
        this.viewerContainer = app.ViewerScene;
        this.scene = app.Scene;

        this.producerScene = null;
        this.wireframe = false;
        this.firstTimeLoadingScene = true;
        this.loadingCount = 0;
        this.cachedSchema = null;
        this.subscribers = [];
    }

    Subscribe(callback: ProducerRefreshCallback): void {
        this.subscribers.push(callback);
    }

    AddLoading(): void {
        this.loadingCount += 1;
    }

    RemoveLoading(): void {
        if (this.loadingCount === 0) {
            throw new Error("loading count already 0");
        }
        this.loadingCount -= 1;

        if (this.loadingCount === 0 && this.cachedSchema) {
            this.Refresh(this.cachedSchema)
            this.cachedSchema = null;
        }
    }

    CurrentlyLoading(): boolean {
        return this.loadingCount > 0;
    }

    NewSchema(schema: GraphInstance): void {
        if (this.CurrentlyLoading()) {
            this.cachedSchema = schema;
            return;
        }
        this.Refresh(schema);
    }

    Render(): void {
        if (this.guassianSplatViewer) {
            this.guassianSplatViewer.update();
            this.guassianSplatViewer.render();
        }
    }

    loadText(producerURL: string) {
        this.AddLoading();
        this.requestManager.fetchText(
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

    loadImage(producerURL: string) {
        this.AddLoading();
        this.requestManager.fetchImage(
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

    viewAABB(aabb: Box3): void {

        const aabbDepth = (aabb.max.z - aabb.min.z)
        const aabbWidth = (aabb.max.x - aabb.min.x)
        const aabbHeight = (aabb.max.y - aabb.min.y)
        const aabbHalfHeight = aabbHeight / 2
        const mid = (aabb.max.y + aabb.min.y) / 2

        if (this.firstTimeLoadingScene && isFinite(aabbWidth) && isFinite(aabbDepth) && isFinite(aabbHeight)) {
            // console.log("Camera position intialized", aabbWidth, aabbDepth, aabbHeight);
            this.firstTimeLoadingScene = false;

            this.camera.position.y = (- mid + aabbHalfHeight) * (3 / 2);
            this.camera.position.z = Math.sqrt(
                (aabbWidth * aabbWidth) +
                (aabbDepth * aabbDepth) +
                (aabbHeight * aabbHeight)
            ) / 2;

            this.orbitControls.target.set(
                (aabb.max.x + aabb.min.x) / 2,
                - mid + aabbHalfHeight,
                (aabb.max.z + aabb.min.z) / 2
            );
            this.orbitControls.update();
        }
    }

    loadGltf(key: string, producerURL: string) {
        this.AddLoading();
        loader.load(
            producerURL,
            ((gltf) => {

                const aabb = new Box3();
                aabb.setFromObject(gltf.scene);
                const aabbHeight = (aabb.max.y - aabb.min.y)
                const aabbHalfHeight = aabbHeight / 2
                const mid = (aabb.max.y + aabb.min.y) / 2

                this.producerScene.add(gltf.scene);

                // We have to do this weird thing because the pivot of the scene
                // Isn't always the center of the AABB
                this.viewerContainer.position.set(0, - mid + aabbHalfHeight, 0)

                const objects = [];

                gltf.scene.traverse((object) => {
                    if (object.isMesh) {
                        object.castShadow = true;
                        object.receiveShadow = true;
                        object.material.wireframe = this.wireframe;
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

                if (typeof error === 'object' && "response" in error) {
                    var resp = error.response as any;
                    resp.json().then(x => {
                        ErrorManager.ShowError(key, x.error);
                    })
                } else {
                    console.error("Unkown error type from gltf loading", error)
                }

            });
    }

    loadSplat(key: string, producerURL: string): void {
        this.AddLoading();
        if (this.guassianSplatViewer) {
            this.guassianSplatViewer.dispose();
        }

        this.renderer.setPixelRatio(1);

        const wasm = true;
        // https://github.com/mkkellogg/GaussianSplats3D/blob/main/src/Viewer.js
        const splatViewerOptions = {
            selfDrivenMode: false,
            // 'cameraUp': [0, -1, 0],
            sphericalHarmonicsDegree: 2,
            useBuiltInControls: false,
            rootElement: this.renderer.domElement.parentElement,
            renderer: this.renderer,
            threeScene: this.scene,
            camera: this.camera,
            // 'sceneRevealMode': GaussianSplats3D.SceneRevealMode.Instant,

            gpuAcceleratedSort: !wasm,
            sharedMemoryForWorkers: !wasm
        }

        this.guassianSplatViewer = new GaussianSplats3D.Viewer(splatViewerOptions);

        // getSplatCenter
        this.guassianSplatViewer.addSplatScene(producerURL, {
            // rotation: [1, 0, 0, 0],
            // scale: [-1, -1, 1, 0],
            // streamView: false
            // showLoadingUI: false,
            // 'scale': [0.25, 0.25, 0.25],
        }).then((() => {

            this.guassianSplatViewer.splatMesh.onSplatTreeReady((splatTree) => {
                const tree = splatTree.subTrees[0]
                const aabb = new Box3();
                aabb.setFromPoints([tree.sceneMin, tree.sceneMax]);
                const aabbHeight = (aabb.max.y - aabb.min.y)
                const aabbHalfHeight = aabbHeight / 2
                const mid = (aabb.max.y + aabb.min.y) / 2

                const shiftY = - mid + aabbHalfHeight
                this.guassianSplatViewer.splatMesh.position.set(0, shiftY, 0)
                this.viewerContainer.position.set(0, shiftY, 0);

                this.viewAABB(aabb);
            });

            this.RemoveLoading();
            this.UpdateSubscribers(producerURL, this.guassianSplatViewer.splatMesh);

        }).bind(this)).catch(x => {
            console.error(x)
            this.RemoveLoading();
            ErrorManager.ShowError(key, x.error);
        })
    }

    SetWireframe(wireframe: boolean): void {
        this.wireframe = wireframe;
        this.producerScene.traverse((object) => {
            // https://discourse.threejs.org/t/gltf-scene-traverse-property-ismesh-does-not-exist-on-type-object3d/27212
            if (object instanceof Mesh) {
                object.material.wireframe = wireframe;
            }
        });
    }

    Refresh(schema: GraphInstance) {
        InfoManager.ClearInfo();

        if (this.producerScene != null) {
            this.viewerContainer.remove(this.producerScene)
        }

        this.producerScene = new Group();
        this.viewerContainer.add(this.producerScene);

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

    UpdateSubscribers(url: string, thing: any) {
        this.subscribers
            .forEach(sub => {
                if (!sub) {
                    return;
                }
                sub(url, thing);
            })
    }
}

