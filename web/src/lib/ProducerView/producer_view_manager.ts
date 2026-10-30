import {
  Box3,
  DirectionalLight,
  GridHelper,
  Material,
  MathUtils,
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
  AnimationAction,
  AnimationClip,
  LoopOnce,
  BoxHelper,
  DoubleSide,
  InstancedMesh,
  LessEqualDepth,
  MeshBasicMaterial,
  Object3D,
  SkinnedMesh,
  Sphere,
} from "three";
import { messageActions } from "@/stores/messageStore";
import { modelStatsActions } from "@/stores/modelStatsStore";
import { animationActions } from "@/stores/animationStore";
import { countModelStats } from "./model_stats";
import { buildHierarchy, walkHierarchy } from "./model_hierarchy";
import { ShadingMode, debugMaterialFor } from "./debug_materials";
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
import { requestRender } from "../render_scheduler";
import { SchemaManager } from "../schema_manager";
import { thumbnails } from "../../features/preview/thumbnail_renderer";

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

// Stands the model on the floor. An empty model's box runs from +Infinity to
// -Infinity, and shifting by that would put every gizmo at NaN.
function floorShift(aabb: Box3): number {
  return aabb.isEmpty() ? 0 : -aabb.min.y;
}

// Cells are sized to the model; the grid then runs well past it so the model
// sits on a floor rather than a mat cut to fit.
const GRID_SLACK = 1.15;
const GRID_REACH = 5;
const GRID_MAX_DIVISIONS = 500;
const CAPTURE_PADDING = 1.1;
const CAPTURE_FIT_PASSES = 3;

function boxCorners(box: Box3): Vector3[] {
  const corners: Vector3[] = [];
  for (const x of [box.min.x, box.max.x]) {
    for (const y of [box.min.y, box.max.y]) {
      for (const z of [box.min.z, box.max.z]) {
        corners.push(new Vector3(x, y, z));
      }
    }
  }
  return corners;
}

const HIGHLIGHT_COLOR = 0x4fb3b3;

const textureLoader = new TextureLoader();
const textureEquirec = textureLoader.load(
  "https://i.imgur.com/Ev4X4yY_d.webp?maxwidth=1520&fidelity=grand"
);
textureEquirec.mapping = EquirectangularReflectionMapping;
textureEquirec.colorSpace = SRGBColorSpace;

export class ProducerViewManager {
  loadingCount: number;

  wireframe: boolean;

  shading: ShadingMode;

  grid: Group;

  gridHelper: GridHelper | null;

  producerItemSubscriber: Array<ProducerRefreshCallback>;

  completeRefreshSubscriber: Array<() => void>;

  cachedSchema: GraphInstance;

  private readonly shownVersions = new Map<string, number>();

  private loadFailed = false;

  firstTimeLoadingScene: boolean;

  renderer: WebGLRenderer;

  requestManager: RequestManager;

  camera: PerspectiveCamera;

  dirLight: DirectionalLight;

  ground: Mesh;

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

  private runningTimer: number | null = null;

  private runningClock: number | null = null;

  mixer: AnimationMixer | null;

  private animationClips: AnimationClip[] = [];

  private animationAction: AnimationAction | null = null;

  private animationIndex = 0;

  private animationPlaying = true;

  private animationLoop = true;

  private animationTime = 0;

  private readonly hiddenParts = new Set<string>();

  private highlightedPart: string | null = null;

  private readonly highlightGroup = new Group();

  private readonly highlightMaterial = new MeshBasicMaterial({
    color: HIGHLIGHT_COLOR,
    transparent: true,
    opacity: 0.4,
    depthWrite: false,
    depthFunc: LessEqualDepth,
    side: DoubleSide,
  });

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
    this.ground = app.Ground.Mesh;
    this.ssaoEffect = app.PostProcessing.SSAO;
    this.orbitControls = app.OrbitControls;
    this.viewerContainer = app.ViewerScene;
    this.grid = app.Grid;
    this.gridHelper = null;
    this.scene = app.Scene;
    this.scene.add(this.highlightGroup);

    this.producerScene = null;
    this.guassianSplatViewer = null;
    this.wireframe = false;
    this.shading = ShadingMode.Shaded;
    this.firstTimeLoadingScene = true;
    this.loadingCount = 0;
    this.cachedSchema = null;
    this.producerItemSubscriber = [];
  }

  private updateLoop(delta: number): void {
    if (this.animationAction && this.animationPlaying) {
      const duration = this.animationAction.getClip().duration;
      let time = this.animationTime + delta;
      if (time >= duration) {
        if (this.animationLoop && duration > 0) {
          time %= duration;
        } else {
          time = duration;
          this.animationPlaying = false;
        }
      }
      this.applyAnimationTime(time);
      requestRender();
    }

    if (this.guassianSplatViewer) {
      requestRender();
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
    if (!this.runningMessage || this.runningTimer !== null) {
      return;
    }
    const started = performance.now();
    this.runningTimer = window.setTimeout(() => {
      this.runningMessage!.style.display = "flex";
      this.runningClock = window.setInterval(() => {
        const seconds = (performance.now() - started) / 1000;
        this.runningMessage!.querySelector(".running-elapsed")!.textContent = `${seconds.toFixed(seconds < 10 ? 1 : 0)}s`;
      }, 100);
    }, 500);
  }

  private hideRunningMessage() {
    if (this.runningTimer !== null) {
      window.clearTimeout(this.runningTimer);
      this.runningTimer = null;
    }
    if (this.runningClock !== null) {
      window.clearInterval(this.runningClock);
      this.runningClock = null;
    }
    if (this.runningMessage) {
      this.runningMessage.style.display = "none";
    }
  }

  AddLoading(): void {
    this.loadingCount += 1;
  }

  RemoveLoading(failed = false): void {
    if (this.loadingCount === 0) {
      throw new Error("loading count already 0");
    }
    this.loadingCount -= 1;
    this.loadFailed ||= failed;

    if (this.loadingCount !== 0) {
      return;
    }

    if (this.loadFailed) {
      this.shownVersions.clear();
    }
    const cached = this.cachedSchema;
    this.cachedSchema = null;
    if (cached && this.Refresh(cached)) {
      return;
    }

    // We're all done loading!!!
    this.hideRunningMessage();
    thumbnails.release();
    this.applyHiddenParts();
    this.fitGridToModel();
    this.applyShading();
    this.applyWireframe();
    this.publishModelStats();
    this.rebuildHighlight();
    for (let i = 0; i < this.completeRefreshSubscriber.length; i++) {
      this.completeRefreshSubscriber[i]();
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

  // Fits the visible model inside the centre square of the frame; call the result to undo.
  FrameForCapture(): () => void {
    const box = this.visibleModelBounds();
    if (box.isEmpty()) {
      return () => {};
    }

    const camera = this.camera;
    const target = this.orbitControls.target;
    const savedPosition = camera.position.clone();
    const savedNear = camera.near;
    const savedFar = camera.far;
    const savedFog = this.scene.fog;

    const hidden: Object3D[] = [];
    for (const child of this.scene.children) {
      if (child.visible && child !== this.viewerContainer && !(child as { isLight?: boolean }).isLight) {
        child.visible = false;
        hidden.push(child);
      }
    }
    this.scene.fog = null;

    const sphere = box.getBoundingSphere(new Sphere());
    const halfFov = MathUtils.degToRad(camera.fov / 2);
    const fitHalfAngle = camera.aspect >= 1 ? halfFov : Math.atan(Math.tan(halfFov) * camera.aspect);
    let distance = (sphere.radius * CAPTURE_PADDING) / Math.sin(fitHalfAngle);

    const direction = savedPosition.clone().sub(target);
    if (direction.lengthSq() === 0) {
      direction.set(0, 0.5, 1);
    }
    direction.normalize();

    const cropX = camera.aspect >= 1 ? 1 / camera.aspect : 1;
    const cropY = camera.aspect >= 1 ? 1 : camera.aspect;
    const corners = boxCorners(box);
    const projected = new Vector3();
    for (let i = 0; i < CAPTURE_FIT_PASSES; i++) {
      camera.position.copy(sphere.center).addScaledVector(direction, distance);
      camera.lookAt(sphere.center);
      this.setClipPlanes(distance / 100, distance + sphere.radius * 4);
      camera.updateMatrixWorld();
      let extent = 0;
      for (const corner of corners) {
        projected.copy(corner).project(camera);
        extent = Math.max(extent, Math.abs(projected.x) / cropX, Math.abs(projected.y) / cropY);
      }
      distance = Math.max(distance * extent * CAPTURE_PADDING, sphere.radius * 1.2);
    }
    camera.position.copy(sphere.center).addScaledVector(direction, distance);
    camera.lookAt(sphere.center);
    this.setClipPlanes(distance / 100, distance + sphere.radius * 4);

    return () => {
      for (const child of hidden) {
        child.visible = true;
      }
      this.scene.fog = savedFog;
      camera.position.copy(savedPosition);
      camera.lookAt(target);
      this.setClipPlanes(savedNear, savedFar);
    };
  }

  private setClipPlanes(near: number, far: number): void {
    this.camera.near = near;
    this.camera.far = far;
    this.camera.updateProjectionMatrix();
    const ssao = this.ssaoEffect.ssaoMaterial.uniforms;
    ssao.cameraNearFar.value.set(near, far);
    ssao.projectionMatrix.value.copy(this.camera.projectionMatrix);
    ssao.inverseProjectionMatrix.value.copy(this.camera.projectionMatrixInverse);
  }

  private visibleModelBounds(): Box3 {
    const box = new Box3();
    if (!this.producerScene) {
      return box;
    }
    const part = new Box3();
    this.producerScene.updateWorldMatrix(true, true);
    this.producerScene.traverseVisible((obj) => {
      if (obj instanceof InstancedMesh) {
        obj.computeBoundingBox();
        part.copy(obj.boundingBox!);
      } else {
        const geometry = (obj as Mesh).geometry;
        if (!geometry) {
          return;
        }
        geometry.computeBoundingBox();
        part.copy(geometry.boundingBox!);
      }
      if (!part.isEmpty()) {
        box.union(part.applyMatrix4(obj.matrixWorld));
      }
    });
    return box;
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
        this.RemoveLoading(true);
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
        this.RemoveLoading(true);
        console.error("unable to load image", producerURL, error);
        messageActions.showError(producerURL, loadFailureMessage(error, "Failed to load image"));
      }
    );
  }

  viewAABB(aabb: Box3): void {
    const diagonal = aabb.getSize(new Vector3()).length();
    if (aabb.isEmpty() || !isFinite(diagonal) || diagonal <= 0) {
      return;
    }

    this.orbitControls.maxDistance = diagonal * 20;

    if (this.firstTimeLoadingScene) {
      this.firstTimeLoadingScene = false;
      this.frameModel(aabb);
    }
  }

  // aabb is in the model's own space; the model is drawn lifted by floorShift.
  private frameModel(aabb: Box3): void {
    const center = aabb.getCenter(new Vector3());
    const half = aabb.getSize(new Vector3()).multiplyScalar(0.5);
    center.y += floorShift(aabb);

    const azimuth = MathUtils.degToRad(35);
    const elevation = MathUtils.degToRad(25);
    const back = new Vector3(
      Math.sin(azimuth) * Math.cos(elevation),
      Math.sin(elevation),
      Math.cos(azimuth) * Math.cos(elevation),
    );
    const right = new Vector3(0, 1, 0).cross(back).normalize();
    const up = back.clone().cross(right);

    const tanV = Math.tan(MathUtils.degToRad(this.camera.fov) / 2);
    const tanH = tanV * this.camera.aspect;
    let distance = 0;
    const corner = new Vector3();
    for (const sx of [-1, 1]) {
      for (const sy of [-1, 1]) {
        for (const sz of [-1, 1]) {
          corner.set(sx * half.x, sy * half.y, sz * half.z);
          const depth = corner.dot(back);
          distance = Math.max(
            distance,
            Math.abs(corner.dot(right)) / tanH + depth,
            Math.abs(corner.dot(up)) / tanV + depth,
          );
        }
      }
    }

    this.orbitControls.target.copy(center);
    this.camera.position.copy(center).addScaledVector(back, distance * 1.1);
    this.orbitControls.update();
    requestRender();
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

    // Shadow coords are interpolated per vertex, so ground corners far outside
    // the shadow frustum lose enough precision to make the shadow jump.
    this.ground.position.set(center.x, 0, center.z);
    this.ground.scale.setScalar(radius / 2);

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

    // radius is resolution-relative and needs no rescaling.
    // worldProximityThreshold/Falloff are world-space, so they do.
    // const ssaoMaterial = this.ssaoEffect.ssaoMaterial;
    // ssaoMaterial.worldProximityThreshold = radius * 0.15;
    // ssaoMaterial.worldProximityFalloff = radius * 0.05;
  }

  loadObj(objLoader: OBJLoader, key: string, producerURL: string): void {
    this.AddLoading();

    objLoader.load(
      producerURL,
      (obj) => {
        this.cleanProducerScene();

        const aabb = new Box3();
        aabb.setFromObject(obj);

        this.producerScene.add(obj);
        requestRender();

        this.viewerContainer.position.set(0, floorShift(aabb), 0);

        this.viewAABB(aabb);
        this.fitShadowToViewerContainer();

        this.RemoveLoading();
      },
      undefined,
      (err) => {
        this.cleanProducerScene();
        console.error(err);
        this.RemoveLoading(true);
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

        this.producerScene.add(gltf.scene);
        requestRender();

        this.viewerContainer.position.set(0, floorShift(aabb) + 0.001, 0);

        const objects = [];

        if (gltf.animations && gltf.animations.length > 0) {
          this.mixer = new AnimationMixer(gltf.scene);
          this.animationClips = gltf.animations;
          this.PlayAnimationClip(Math.min(this.animationIndex, gltf.animations.length - 1));
        }

        gltf.scene.traverse((object) => {
          if (object.isMesh) {
            object.castShadow = true;
            object.receiveShadow = true;
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
        this.RemoveLoading(true);
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

        geometry.computeVertexNormals();

        const material = new MeshStandardMaterial({});
        const mesh = new Mesh(geometry, material);
        mesh.castShadow = true;
        mesh.receiveShadow = true;

        const aabb = new Box3();
        aabb.setFromObject(mesh);

        this.producerScene.add(mesh);
        requestRender();

        this.viewerContainer.position.set(0, floorShift(aabb), 0);

        this.viewAABB(aabb);
        this.fitShadowToViewerContainer();

        this.RemoveLoading();
      },
      undefined,
      (err) => {
        console.error(err);
        this.cleanProducerScene();
        this.RemoveLoading(true);
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
      sharedMemoryForWorkers: self.crossOriginIsolated,
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
            const shiftY = floorShift(aabb);
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
        this.RemoveLoading(true);
        messageActions.showError(
          key,
          loadFailureMessage(x, "Failed to load splat")
        );
      });
  }

  SetShading(shading: ShadingMode): void {
    this.shading = shading;
    this.applyShading();
    this.applyWireframe();
  }

  private applyShading(): void {
    const override = debugMaterialFor(this.shading);
    this.producerScene?.traverse((object) => {
      if (!(object instanceof Mesh)) {
        return;
      }
      if (object.userData.polyformMaterial === undefined) {
        object.userData.polyformMaterial = object.material;
      }
      object.material = override ?? object.userData.polyformMaterial;
    });
  }

  /** Re-aims the shadow camera after the light has been moved. */
  RefitShadowCamera(): void {
    this.fitShadowToViewerContainer();
  }

  /**
   * The grid marks the model's own origin, not the shift applied to stand the
   * model on the floor, so it answers "where is 0,0,0" rather than "where is
   * the bottom of this mesh". Cells land on a power of ten so it reads as a
   * ruler, and it spans several times the model's reach from origin.
   */
  private fitGridToModel(): void {
    const box = new Box3().setFromObject(this.producerScene);
    if (box.isEmpty() || !isFinite(box.min.x) || !isFinite(box.max.x)) {
      return;
    }

    // Measured from origin rather than across the model, so a mesh authored
    // far off-centre still has the origin and itself on the same grid.
    const reach = Math.max(
      Math.abs(box.min.x),
      Math.abs(box.max.x),
      Math.abs(box.min.z),
      Math.abs(box.max.z)
    );
    if (!isFinite(reach) || reach <= 0) {
      return;
    }

    const extent = reach * GRID_SLACK;
    const cell = Math.pow(10, Math.round(Math.log10(extent / 10)));
    const divisions = Math.min(
      Math.max(Math.ceil((extent * GRID_REACH) / cell) * 2, 4),
      GRID_MAX_DIVISIONS
    );

    this.buildGrid(divisions * cell, divisions);
    this.grid.position.y = this.viewerContainer.position.y + cell * 0.001;
  }

  private buildGrid(size: number, divisions: number): void {
    if (this.gridHelper) {
      this.gridHelper.removeFromParent();
      this.gridHelper.geometry.dispose();
      (this.gridHelper.material as Material).dispose();
    }
    this.gridHelper = new GridHelper(size, divisions, 0x6b6b6b, 0xb4b4b4);
    (this.gridHelper.material as Material).depthWrite = false;
    this.grid.add(this.gridHelper);
  }

  SetWireframe(wireframe: boolean): void {
    this.wireframe = wireframe;
    this.applyWireframe();
  }

  private applyWireframe(): void {
    this.producerScene?.traverse((object) => {
      // https://discourse.threejs.org/t/gltf-scene-traverse-property-ismesh-does-not-exist-on-type-object3d/27212
      if (!(object instanceof Mesh)) {
        return;
      }
      const materials = Array.isArray(object.material)
        ? object.material
        : [object.material];
      for (const material of materials) {
        if (material && "wireframe" in material) {
          material.wireframe = this.wireframe;
        }
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

  private publishModelStats(): void {
    const stats = countModelStats(this.producerScene, false);
    if (stats.vertices === 0) {
      modelStatsActions.set(null, []);
      return;
    }
    modelStatsActions.set(countModelStats(this.producerScene), buildHierarchy(this.producerScene));
  }

  private findPart(path: string): Object3D | null {
    let found: Object3D | null = null;
    walkHierarchy(this.producerScene, (p, object) => {
      if (p === path) {
        found = object;
      }
    });
    return found;
  }

  private applyHiddenParts(): void {
    walkHierarchy(this.producerScene, (path, object) => {
      if (this.hiddenParts.has(path)) {
        object.visible = false;
      }
    });
  }

  SetPartVisible(path: string, visible: boolean): void {
    if (visible) {
      this.hiddenParts.delete(path);
    } else {
      this.hiddenParts.add(path);
    }
    const part = this.findPart(path);
    if (part) {
      part.visible = visible;
    }
    this.publishModelStats();
    this.rebuildHighlight();
    requestRender();
  }

  HighlightPart(path: string | null): void {
    if (path === this.highlightedPart) {
      return;
    }
    this.highlightedPart = path;
    this.rebuildHighlight();
    requestRender();
  }

  private clearHighlight(): void {
    for (const child of [...this.highlightGroup.children]) {
      if (child instanceof BoxHelper) {
        child.geometry.dispose();
        (child.material as Material).dispose();
      }
      this.highlightGroup.remove(child);
    }
  }

  private rebuildHighlight(): void {
    this.clearHighlight();
    const target = this.highlightedPart ? this.findPart(this.highlightedPart) : null;
    if (!target) {
      return;
    }

    const box = new BoxHelper(target, HIGHLIGHT_COLOR);
    box.frustumCulled = false;
    box.onBeforeRender = () => box.update();
    this.highlightGroup.add(box);

    for (let o: Object3D | null = target; o; o = o.parent) {
      if (!o.visible) {
        return;
      }
    }

    this.highlightMaterial.wireframe = this.wireframe;
    target.traverseVisible((object) => {
      if (object instanceof Mesh) {
        this.highlightGroup.add(this.highlightOverlay(object));
      }
    });
  }

  // Must stay outside producerScene: shading, wireframe and stat passes traverse it.
  private highlightOverlay(source: Mesh): Mesh {
    let overlay: Mesh;
    if (source instanceof SkinnedMesh) {
      const skinned = new SkinnedMesh(source.geometry, this.highlightMaterial);
      skinned.bind(source.skeleton, source.bindMatrix);
      skinned.bindMode = source.bindMode;
      overlay = skinned;
    } else if (source instanceof InstancedMesh) {
      const instanced = new InstancedMesh(source.geometry, this.highlightMaterial, source.count);
      instanced.instanceMatrix = source.instanceMatrix;
      overlay = instanced;
    } else {
      overlay = new Mesh(source.geometry, this.highlightMaterial);
    }
    overlay.morphTargetInfluences = source.morphTargetInfluences;
    overlay.morphTargetDictionary = source.morphTargetDictionary;
    overlay.matrixAutoUpdate = false;
    overlay.matrixWorldAutoUpdate = false;
    overlay.frustumCulled = false;
    overlay.renderOrder = 1;
    overlay.onBeforeRender = () => {
      overlay.matrixWorld.copy(source.matrixWorld);
      if (overlay instanceof SkinnedMesh && source instanceof SkinnedMesh) {
        overlay.bindMatrixInverse.copy(source.bindMatrixInverse);
      }
    };
    return overlay;
  }

  PlayAnimationClip(index: number): void {
    const clip = this.animationClips[index];
    if (!this.mixer || !clip) {
      return;
    }
    this.mixer.stopAllAction();
    if (index !== this.animationIndex) {
      this.animationTime = 0;
    }
    this.animationIndex = index;
    this.animationAction = this.mixer.clipAction(clip);
    this.animationAction.setLoop(LoopOnce, 1);
    this.animationAction.clampWhenFinished = true;
    this.animationAction.play();
    this.animationAction.paused = true;
    animationActions.set({
      clips: this.animationClips.map((c, i) => ({ name: c.name || `Animation ${i + 1}`, duration: c.duration })),
      active: index,
      playing: this.animationPlaying,
      loop: this.animationLoop,
    });
    this.applyAnimationTime(Math.min(this.animationTime, clip.duration));
    requestRender();
  }

  SetAnimationPlaying(playing: boolean): void {
    if (playing && this.animationAction && this.animationTime >= this.animationAction.getClip().duration) {
      this.applyAnimationTime(0);
    }
    this.animationPlaying = playing;
    animationActions.set({ playing });
  }

  SetAnimationLoop(loop: boolean): void {
    this.animationLoop = loop;
    animationActions.set({ loop });
  }

  SeekAnimation(time: number): void {
    this.applyAnimationTime(time);
    requestRender();
  }

  private applyAnimationTime(time: number): void {
    this.animationTime = time;
    if (this.animationAction && this.mixer) {
      this.animationAction.time = time;
      this.mixer.update(0);
    }
    animationActions.set({ time, playing: this.animationPlaying });
  }

  private clearAnimation(): void {
    this.mixer?.stopAllAction();
    this.mixer = null;
    this.animationAction = null;
    this.animationClips = [];
    animationActions.set({ clips: [] });
  }

  private cleanProducerScene() {
    this.clearHighlight();
    this.clearAnimation();
    if (this.producerScene != null) {
      this.viewerContainer.remove(this.producerScene);
    }

    this.producerScene = new Group();
    this.viewerContainer.add(this.producerScene);
    requestRender();
  }

  Refresh(schema: GraphInstance): boolean {
    // this.cleanProducerScene();
    messageActions.clearInfo();

    const producers: Array<[string, string, number]> = [];
    for (const [nodeID, nodeInstance] of Object.entries(schema.nodes)) {
      const portName = this.nodeTypeManifestPorts.get(nodeInstance.type);
      if (portName === undefined) {
        continue;
      }
      producers.push([nodeID, portName, nodeInstance.output?.[portName]?.version ?? 0]);
    }

    const shown =
      producers.length > 0 &&
      producers.length === this.shownVersions.size &&
      producers.every(([nodeID, portName, version]) => this.shownVersions.get(`${nodeID}/${portName}`) === version);
    if (shown) {
      return false;
    }
    this.shownVersions.clear();
    this.loadFailed = false;

    for (const [nodeID, portName] of producers) {
      this.AddLoading();
      void this.requestManager.getManifest(nodeID, portName).then((loaded) => {
        if (loaded !== undefined) {
          this.shownVersions.set(`${nodeID}/${portName}`, loaded.version);
          this.ManifestLoaded(nodeID, portName, loaded.manifest);
        }
        this.RemoveLoading(loaded === undefined);
      });
    }
    // Nothing to load means nothing will ever call the loading done.
    if (producers.length > 0) {
      thumbnails.hold();
      this.showRunningMessage();
    } else {
      this.hideRunningMessage();
      thumbnails.release();
    }
    return producers.length > 0;
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
