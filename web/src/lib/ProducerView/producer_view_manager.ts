import {
  Box3,
  DirectionalLight,
  EquirectangularReflectionMapping,
  Group,
  Mesh,
  MeshStandardMaterial,
  PerspectiveCamera,
  Scene,
  SRGBColorSpace,
  TextureLoader,
  Vector3,
  WebGLRenderer,
  AnimationMixer,
} from "three";
import { messageActions } from "@/stores/messageStore";
import { getApiErrorMessage } from "@/api/client";
import { GraphInstance, Manifest, NodeDefinition } from "../schema";
import { getFileExtension } from "../utils";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader.js";
import { OBJLoader } from "three/examples/jsm/loaders/OBJLoader.js";
import { PLYLoader } from "three/examples/jsm/loaders/PLYLoader.js";
import { SSAOEffect } from "postprocessing";
import { RequestManager } from "../requests";
import * as GaussianSplats3D from "@mkkellogg/gaussian-splats-3d";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import { ThreeApp } from "../three_app";
import { SchemaManager } from "../schema_manager";

type ProducerRefreshCallback = (string: string, thing: any) => void;

function loadFailureMessage(error: unknown, fallback: string): string {
  if (typeof error === "string" && error) {
    try {
      return getApiErrorMessage(JSON.parse(error), fallback);
    } catch {
      return error;
    }
  }
  return fallback;
}

const textureLoader = new TextureLoader();
const textureEquirec = textureLoader.load(
  "https://i.imgur.com/Ev4X4yY_d.webp?maxwidth=1520&fidelity=grand"
);
textureEquirec.mapping = EquirectangularReflectionMapping;
textureEquirec.colorSpace = SRGBColorSpace;

export class ProducerViewManager {
  loadingCount: number;

  wireframe: boolean;

  producerItemSubscriber: Array<ProducerRefreshCallback>;

  completeRefreshSubscriber: Array<() => void>;

  cachedSchema: GraphInstance;

  firstTimeLoadingScene: boolean;

  renderer: WebGLRenderer;

  requestManager: RequestManager;

  camera: PerspectiveCamera;

  dirLight: DirectionalLight;

  ssaoEffect: SSAOEffect;

  guassianSplatViewer: GaussianSplats3D.Viewer | null;

  producerScene: Group;

  orbitControls: OrbitControls;

  viewerContainer: Group;

  scene: Scene;

  nodeTypeManifestPorts: Map<string, string>;

  modelVersion: number;

  readonly schemaManager: SchemaManager;

  readonly runningMessage: HTMLElement | undefined;

  mixer: AnimationMixer | null;

  constructor(
    app: ThreeApp,
    requestManager: RequestManager,
    nodeTypes: Array<NodeDefinition>,
    schemaManager: SchemaManager
  ) {
    this.runningMessage = document.getElementById("running-message");
    this.completeRefreshSubscriber = [];
    this.modelVersion = -1;
    this.schemaManager = schemaManager;
    this.nodeTypeManifestPorts = new Map<string, string>();
    app.UpdateLoop.addToUpdate({
      name: "Model Animation",
      loop: (delta) => {
        this.updateLoop(delta);
      },
    });

    for (let i = 0; i < nodeTypes.length; i++) {
      const nodeType = nodeTypes[i];
      if (!nodeType.outputs) {
        continue;
      }
      for (const [outputName, output] of Object.entries(nodeType.outputs)) {
        if (
          output.type ===
          "github.com/EliCDavis/polyform/generator/manifest.Manifest"
        ) {
          this.nodeTypeManifestPorts.set(nodeType.type, outputName);
        }
      }
    }

    this.requestManager = requestManager;
    this.renderer = app.Renderer;
    this.camera = app.Camera;
    this.dirLight = app.Lighting.DirLight;
    this.ssaoEffect = app.PostProcessing.SSAO;
    this.orbitControls = app.OrbitControls;
    this.viewerContainer = app.ViewerScene;
    this.scene = app.Scene;

    this.producerScene = null;
    this.guassianSplatViewer = null;
    this.wireframe = false;
    this.firstTimeLoadingScene = true;
    this.loadingCount = 0;
    this.cachedSchema = null;
    this.producerItemSubscriber = [];
  }

  private updateLoop(delta: number): void {
    if (this.mixer) {
      this.mixer.update(delta);
    }
  }

  setModelVersion(newModelVersion: number): void {
    if (newModelVersion === this.modelVersion) {
      return;
    }
    const previousVersion = this.modelVersion;
    this.modelVersion = newModelVersion;
    // Initial poll only records version; schema is already loading from bootstrap.
    if (previousVersion === -1) {
      return;
    }
    this.schemaManager.refreshSchema("Model version change");
  }

  SubscribeToProducerRefresh(callback: ProducerRefreshCallback): void {
    this.producerItemSubscriber.push(callback);
  }

  // Called whenever
  SubscribeToCompleteRefresh(subsriber: () => void): void {
    this.completeRefreshSubscriber.push(subsriber);
  }

  private showRunningMessage() {
    if (this.runningMessage) {
      this.runningMessage.style.display = "block";
    }
  }

  private hideRunningMessage() {
    if (this.runningMessage) {
      this.runningMessage.style.display = "none";
    }
  }

  AddLoading(): void {
    this.loadingCount += 1;
  }

  RemoveLoading(): void {
    if (this.loadingCount === 0) {
      throw new Error("loading count already 0");
    }
    this.loadingCount -= 1;

    if (this.loadingCount === 0) {
      if (this.cachedSchema) {
        this.Refresh(this.cachedSchema);
        this.cachedSchema = null;
      } else {
        // We're all done loading!!!
        this.hideRunningMessage();
        for (let i = 0; i < this.completeRefreshSubscriber.length; i++) {
          this.completeRefreshSubscriber[i]();
        }
      }
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
        messageActions.showInfo(data);
        this.RemoveLoading();
        this.UpdateSubscribers(producerURL, data);
      },
      (error) => {
        this.RemoveLoading();
        console.error("unable to load text", producerURL, error);
        messageActions.showError(producerURL, loadFailureMessage(error, "Failed to load text"));
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
        messageActions.showError(producerURL, loadFailureMessage(error, "Failed to load image"));
      }
    );
  }

  viewAABB(aabb: Box3): void {
    const aabbDepth = aabb.max.z - aabb.min.z;
    const aabbWidth = aabb.max.x - aabb.min.x;
    const aabbHeight = aabb.max.y - aabb.min.y;
    const aabbHalfHeight = aabbHeight / 2;
    const mid = (aabb.max.y + aabb.min.y) / 2;

    if (
      this.firstTimeLoadingScene &&
      isFinite(aabbWidth) &&
      isFinite(aabbDepth) &&
      isFinite(aabbHeight)
    ) {
      // console.log("Camera position intialized", aabbWidth, aabbDepth, aabbHeight);
      this.firstTimeLoadingScene = false;

      this.camera.position.y = (-mid + aabbHalfHeight) * (3 / 2);
      this.camera.position.z =
        Math.sqrt(
          aabbWidth * aabbWidth +
            aabbDepth * aabbDepth +
            aabbHeight * aabbHeight
        ) / 2;

      this.orbitControls.target.set(
        (aabb.max.x + aabb.min.x) / 2,
        -mid + aabbHalfHeight,
        (aabb.max.z + aabb.min.z) / 2
      );
      this.orbitControls.update();
    }
  }

  // Attempt to size the shadows to fit exactly the model to maximize the map 
  // resolution
  private fitShadowToViewerContainer(): void {
    const box = new Box3().setFromObject(this.viewerContainer);
    if (box.isEmpty() || !isFinite(box.min.x) || !isFinite(box.max.x)) {
      return;
    }

    const center = new Vector3();
    box.getCenter(center);
    const size = new Vector3();
    box.getSize(size);
    const radius = Math.max(size.length() / 2, 0.01);

    this.dirLight.target.position.copy(center);
    this.dirLight.target.updateMatrixWorld();

    // Orient the shadow camera exactly as WebGLShadowMap will at render
    // time (same light position, same target, same up vector) so the
    // frustum we compute below matches what's actually used.
    const shadowCamera = this.dirLight.shadow.camera;
    shadowCamera.position.copy(this.dirLight.position);
    shadowCamera.lookAt(center);
    shadowCamera.updateMatrixWorld(true);

    // Add a bunch of padding cause this math sucks
    const groundPadding = radius * 3;
    const groundBox = new Box3(
      new Vector3(center.x - groundPadding, box.min.y - 0.01, center.z - groundPadding),
      new Vector3(center.x + groundPadding, box.min.y + 0.01, center.z + groundPadding),
    );
    const combined = box.clone().union(groundBox);

    const corners = [
      new Vector3(combined.min.x, combined.min.y, combined.min.z),
      new Vector3(combined.min.x, combined.min.y, combined.max.z),
      new Vector3(combined.min.x, combined.max.y, combined.min.z),
      new Vector3(combined.min.x, combined.max.y, combined.max.z),
      new Vector3(combined.max.x, combined.min.y, combined.min.z),
      new Vector3(combined.max.x, combined.min.y, combined.max.z),
      new Vector3(combined.max.x, combined.max.y, combined.min.z),
      new Vector3(combined.max.x, combined.max.y, combined.max.z),
    ];

    let minX = Infinity, maxX = -Infinity;
    let minY = Infinity, maxY = -Infinity;
    let minZ = Infinity, maxZ = -Infinity;
    for (const corner of corners) {
      corner.applyMatrix4(shadowCamera.matrixWorldInverse);
      minX = Math.min(minX, corner.x);
      maxX = Math.max(maxX, corner.x);
      minY = Math.min(minY, corner.y);
      maxY = Math.max(maxY, corner.y);
      minZ = Math.min(minZ, corner.z);
      maxZ = Math.max(maxZ, corner.z);
    }

    const margin = radius * 0.05;
    shadowCamera.left = minX - margin;
    shadowCamera.right = maxX + margin;
    shadowCamera.bottom = minY - margin;
    shadowCamera.top = maxY + margin;
    // The camera looks down its local -Z, so more-negative Z is farther away.
    shadowCamera.near = Math.max(-maxZ - margin, 0.01);
    shadowCamera.far = -minZ + margin;
    shadowCamera.updateProjectionMatrix();

    this.dirLight.shadow.normalBias = radius * 0.01;

    // SSAOEffect's `radius` is resolution-relative (a fraction of screen
    // size), not world-space, so unlike SSAOPass's kernelRadius it doesn't
    // need rescaling per model. worldProximityThreshold/Falloff are still
    // absolute world-space distances though, so a tiny model (or a huge
    // one) still needs them scaled or contact creases either get filtered
    // out as noise or the occlusion washes out across the whole surface.
    const ssaoMaterial = this.ssaoEffect.ssaoMaterial;
    ssaoMaterial.worldProximityThreshold = radius * 0.15;
    ssaoMaterial.worldProximityFalloff = radius * 0.05;
  }

  loadObj(objLoader: OBJLoader, key: string, producerURL: string): void {
    this.AddLoading();

    objLoader.load(
      producerURL,
      (obj) => {
        this.RemoveLoading();
        this.cleanProducerScene();

        const aabb = new Box3();
        aabb.setFromObject(obj);
        const aabbHeight = aabb.max.y - aabb.min.y;
        const aabbHalfHeight = aabbHeight / 2;
        const mid = (aabb.max.y + aabb.min.y) / 2;

        this.producerScene.add(obj);

        // We have to do this weird thing because the pivot of the scene
        // Isn't always the center of the AABB
        this.viewerContainer.position.set(0, -mid + aabbHalfHeight, 0);

        this.viewAABB(aabb);
        this.fitShadowToViewerContainer();
      },
      undefined,
      (err) => {
        this.cleanProducerScene();
        console.error(err);
        this.RemoveLoading();
      }
    );
  }

  loadGltf(gltfLoader: GLTFLoader, key: string, producerURL: string) {
    this.AddLoading();
    gltfLoader.load(
      producerURL,
      ((gltf) => {
        this.cleanProducerScene();

        const aabb = new Box3();
        aabb.setFromObject(gltf.scene);
        const aabbHeight = aabb.max.y - aabb.min.y;
        const aabbHalfHeight = aabbHeight / 2;
        const mid = (aabb.max.y + aabb.min.y) / 2;

        this.producerScene.add(gltf.scene);

        // We have to do this weird thing because the pivot of the scene
        // Isn't always the center of the AABB
        this.viewerContainer.position.set(0, -mid + aabbHalfHeight + 0.001, 0);

        const objects = [];

        if (gltf.animations && gltf.animations.length > 0) {
          this.mixer = new AnimationMixer(gltf.scene);
          this.mixer.clipAction(gltf.animations[0]).play();
        }

        gltf.scene.traverse((object) => {
          if (object.isMesh) {
            object.castShadow = true;
            object.receiveShadow = true;
            object.material.wireframe = this.wireframe;
            object.material.envMap = textureEquirec;
            object.material.needsUpdate = true;
            // object.material.transparent = true;

            objects.push(object);
          } else if (object.isPoints) {
            object.material.size = 2;
          }
        });

        // progressiveSurfacemap.addObjectsToLightMap(objects);

        this.viewAABB(aabb);
        this.fitShadowToViewerContainer();

        this.UpdateSubscribers(producerURL, gltf);

        this.RemoveLoading();
      }).bind(this),
      undefined,
      (error) => {
        this.RemoveLoading();
        this.cleanProducerScene();

        if (typeof error === "object" && "response" in error) {
          var resp = error.response as any;
          resp.json().then((x) => {
            messageActions.showError(key, x.error);
          });
        } else {
          console.error("Unkown error type from gltf loading", error);
        }
      }
    );
  }

  loadPly(plyLoader: PLYLoader, key: string, producerURL: string): void {
    this.AddLoading();

    plyLoader.load(
      producerURL,
      (geometry) => {
        this.cleanProducerScene();

        this.RemoveLoading();
        geometry.computeVertexNormals();

        const material = new MeshStandardMaterial({});
        const mesh = new Mesh(geometry, material);
        mesh.castShadow = true;
        mesh.receiveShadow = true;

        const aabb = new Box3();
        aabb.setFromObject(mesh);
        const aabbHeight = aabb.max.y - aabb.min.y;
        const aabbHalfHeight = aabbHeight / 2;
        const mid = (aabb.max.y + aabb.min.y) / 2;

        this.producerScene.add(mesh);

        // We have to do this weird thing because the pivot of the scene
        // Isn't always the center of the AABB
        this.viewerContainer.position.set(0, -mid + aabbHalfHeight, 0);

        this.viewAABB(aabb);
        this.fitShadowToViewerContainer();
      },
      undefined,
      (err) => {
        console.error(err);
        this.cleanProducerScene();
        this.RemoveLoading();
      }
    );
  }

  loadSplat(key: string, producerURL: string): void {
    this.cleanProducerScene();

    this.AddLoading();
    if (this.guassianSplatViewer) {
      this.guassianSplatViewer.dispose();
      this.guassianSplatViewer = null;
    }

    this.renderer.setPixelRatio(1);

    const wasm = true;
    const splatViewerOptions = {
      selfDrivenMode: false,
      sphericalHarmonicsDegree: 2,
      useBuiltInControls: false,
      rootElement: this.renderer.domElement.parentElement,
      renderer: this.renderer,
      threeScene: this.scene,
      camera: this.camera,
      gpuAcceleratedSort: !wasm,
    };

    this.guassianSplatViewer = new GaussianSplats3D.Viewer(splatViewerOptions);

    this.guassianSplatViewer
      .addSplatScene(producerURL, {})
      .then(
        (() => {
          this.guassianSplatViewer.splatMesh.onSplatTreeReady((splatTree) => {
            const tree = splatTree.subTrees[0];
            const aabb = new Box3();
            aabb.setFromPoints([tree.sceneMin, tree.sceneMax]);
            const aabbHeight = aabb.max.y - aabb.min.y;
            const aabbHalfHeight = aabbHeight / 2;
            const mid = (aabb.max.y + aabb.min.y) / 2;

            const shiftY = -mid + aabbHalfHeight;
            this.guassianSplatViewer.splatMesh.position.set(0, shiftY, 0);
            this.viewerContainer.position.set(0, shiftY, 0);

            this.viewAABB(aabb);
          });

          this.RemoveLoading();
          this.UpdateSubscribers(
            producerURL,
            this.guassianSplatViewer.splatMesh
          );
        }).bind(this)
      )
      .catch((x) => {
        console.error(x);
        this.RemoveLoading();
        messageActions.showError(
          key,
          loadFailureMessage(x, "Failed to load splat")
        );
      });
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

  ManifestLoaded(nodeId: string, portName: string, manifest: Manifest): void {
    const manifestUrl: string = `./manifest/${nodeId}/${portName}/`;
    const fileToLoad = manifest.main;
    const fileToLoadMetadata = manifest.entries[fileToLoad].metadata;

    messageActions.clearError(manifest.main);
    const fileExt = getFileExtension(manifest.main);

    switch (fileExt) {
      case "txt":
        this.loadText(manifestUrl + fileToLoad);
        break;

      case "gltf":
      case "glb":
        const gltfLoader = new GLTFLoader().setPath(manifestUrl);
        this.loadGltf(gltfLoader, fileToLoad, fileToLoad);
        break;

      case "obj":
        const objLoader = new OBJLoader().setPath(manifestUrl);
        this.loadObj(objLoader, fileToLoad, fileToLoad);
        break;

      case "splat":
        this.loadSplat(fileToLoad, manifestUrl + fileToLoad);
        break;

      case "ply":
        if (
          fileToLoadMetadata &&
          fileToLoadMetadata["gaussianSplat"] === true
        ) {
          this.loadSplat(fileToLoad, manifestUrl + fileToLoad);
        } else {
          const plyLoader = new PLYLoader().setPath(manifestUrl);
          this.loadPly(plyLoader, fileToLoad, fileToLoad);
        }
        break;

      // case "png":
      //     this.loadImage(manifestUrl + fileToLoad);
      //     break;
    }
  }

  private cleanProducerScene() {
    if (this.producerScene != null) {
      this.viewerContainer.remove(this.producerScene);
    }

    this.producerScene = new Group();
    this.viewerContainer.add(this.producerScene);
  }

  Refresh(schema: GraphInstance) {
    // this.cleanProducerScene();
    messageActions.clearInfo();

    this.showRunningMessage();

    let hasManifests = false;

    for (const [nodeID, nodeInstance] of Object.entries(schema.nodes)) {
      // This node doesn't have a manifest output, continue
      if (!this.nodeTypeManifestPorts.has(nodeInstance.type)) {
        continue;
      }

      hasManifests = true;
      const portName = this.nodeTypeManifestPorts.get(nodeInstance.type);
      this.requestManager.getManifest(nodeID, portName, (manifest) => {
        this.ManifestLoaded(nodeID, portName, manifest);
      });
    }
    if (hasManifests) {
      this.showRunningMessage();
    }
  }

  UpdateSubscribers(url: string, thing: any) {
    this.producerItemSubscriber.forEach((sub) => {
      if (!sub) {
        return;
      }
      sub(url, thing);
    });
  }
}
