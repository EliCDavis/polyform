import {
  Camera,
  Color,
  DirectionalLight,
  Fog,
  Group,
  HemisphereLight,
  Mesh,
  MeshPhongMaterial,
  NoToneMapping,
  PCFSoftShadowMap,
  PerspectiveCamera,
  PlaneGeometry,
  Scene,
  WebGLRenderer,
} from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import { CSS2DRenderer } from "three/examples/jsm/renderers/CSS2DRenderer.js";
// import { RoomEnvironment } from 'three/examples/jsm/environments/RoomEnvironment.js';

import {
  BloomEffect,
  EffectComposer,
  EffectPass,
  NormalPass,
  RenderPass,
  SSAOEffect,
  ToneMappingEffect,
  ToneMappingMode,
} from "postprocessing";
import { ViewportSettings } from "./viewport_settings";
import { UpdateManager } from "./update_manager";
import { ViewportGizmo } from "three-viewport-gizmo";

// https://threejs.org/examples/?q=Directional#webgl_lights_hemisphere
// https://threejs.org/examples/#webgl_geometry_spline_editor

export interface ThreeAppLighting {
    DirLight: DirectionalLight;
    HemiLight: HemisphereLight;
}

export interface ThreeAppGround {
  Mesh: Mesh;
  Material: MeshPhongMaterial;
}

export interface ThreeAppPostProcessing {
  SSAO: SSAOEffect;
  Bloom: BloomEffect;
  ToneMapping: ToneMappingEffect;
  // All effects share this one EffectPass. Toggle each effect's own
  // blendMode, not this pass's .enabled.
  EffectPass: EffectPass;
}

export interface ThreeApp {
  Camera: PerspectiveCamera;
  Renderer: WebGLRenderer;
  LabelRenderer: CSS2DRenderer;
  OrbitControls: OrbitControls;
  Scene: Scene;
  ViewerScene: Group;
  Ground: ThreeAppGround;
  Lighting: ThreeAppLighting;
  Composer: EffectComposer;
  PostProcessing: ThreeAppPostProcessing;
  Fog: Fog;
  UpdateLoop: UpdateManager;
  ViewportGizmo: ViewportGizmo;
}

export function CreateThreeApp(
  container: HTMLElement,
  viewportSettings: ViewportSettings,
  updateLoop: UpdateManager,
  antiAlias: boolean,
  xrEnabled: boolean,
): ThreeApp {
  // progressiveSurfacemap.addObjectsToLightMap([groundMesh])

  // const environment = new RoomEnvironment(renderer);
  // const pmremGenerator = new THREE.PMREMGenerator(renderer);
  // scene.environment = pmremGenerator.fromScene( environment ).texture;

  const shadowMapRes = 4098,
    lightMapRes = 4098,
    lightCount = 8;

  const camera = new PerspectiveCamera(
    50,
    window.innerWidth / window.innerHeight,
    0.1,
    1000,
  );
  camera.position.set(0, 2, 3);

  const scene = new Scene();
  scene.background = new Color(viewportSettings.background);

  // scene.background = textureEquirec;
  const fog = new Fog(
    viewportSettings.fog.color,
    viewportSettings.fog.near,
    viewportSettings.fog.far,
  );
  scene.fog = fog;

  const viewerContainer = new Group();
  scene.add(viewerContainer);

  let threeCanvas = container.querySelector(
    "#three-canvas",
  ) as HTMLCanvasElement | null;
  if (!threeCanvas) {
    threeCanvas = document.getElementById(
      "three-canvas",
    ) as HTMLCanvasElement | null;
  }
  if (!threeCanvas) {
    threeCanvas = document.createElement("canvas");
    threeCanvas.id = "three-canvas";
    container.appendChild(threeCanvas);
  }

  const renderer = new WebGLRenderer({
    canvas: threeCanvas,
    antialias: antiAlias,
  });
  renderer.setPixelRatio(window.devicePixelRatio);
  // renderer.setSize(threeCanvas.clientWidth, threeCanvas.clientHeight, false);
  renderer.shadowMap.enabled = true;
  renderer.shadowMap.type = PCFSoftShadowMap; // default THREE.PCFShadowMap
  // Tone mapping happens in ToneMappingEffect below - leaving this on
  // double-encodes colors.
  renderer.toneMapping = NoToneMapping;
  renderer.xr.enabled = xrEnabled;
  renderer.setAnimationLoop(updateLoop.run.bind(updateLoop));

  // Uses pmndrs/postprocessing instead of three's example
  // SSAOPass/UnrealBloomPass, which are unmaintained.
  const composer = new EffectComposer(renderer, {
    // AA comes from here; the canvas's own antialias flag is a no-op.
    multisampling: antiAlias ? 4 : 0,
  });
  composer.addPass(new RenderPass(scene, camera));

  const normalPass = new NormalPass(scene, camera);
  composer.addPass(normalPass);

  const ssaoEffect = new SSAOEffect(camera, normalPass.texture, {
    samples: 9,
    rings: 7,
    radius: 0.1825,
    intensity: 1.0,
    luminanceInfluence: 0.7,
  });

  // luminanceThreshold gates which pixels bloom. Tuned so lit surfaces
  // cross it but shadowed ones don't.
  const bloomEffect = new BloomEffect({
    intensity: 0.3,
    luminanceThreshold: 0.35,
    luminanceSmoothing: 0.3,
    mipmapBlur: true,
  });
  // mipmapBlurPass.radius is the blur/unblurred mix factor, not a kernel
  // size - 0 disables the blur entirely.

  const toneMappingEffect = new ToneMappingEffect({
    mode: ToneMappingMode.ACES_FILMIC,
  });

  // One shared EffectPass, in order - tone mapping must run last, after
  // SSAO/bloom see linear HDR values.
  const effectPass = new EffectPass(camera, ssaoEffect, bloomEffect, toneMappingEffect);
  composer.addPass(effectPass);

  // progressive lightmap
  // const progressiveSurfacemap = new ProgressiveLightMap(renderer, lightMapRes);

  const labelRenderer = new CSS2DRenderer();
  // labelRenderer.setSize(threeCanvas.clientWidth, threeCanvas.clientHeight, false);
  labelRenderer.domElement.style.position = "absolute";
  labelRenderer.domElement.style.top = "0px";
  labelRenderer.domElement.style.pointerEvents = "none";
  container.appendChild(labelRenderer.domElement);

  const defaultLightIntensity = 1.1;

  const hemiLight = new HemisphereLight(viewportSettings.lighting, 0x8d8d8d, defaultLightIntensity);
  hemiLight.position.set(0, 20, 0);
  scene.add(hemiLight);

  const dirLight = new DirectionalLight(viewportSettings.lighting, defaultLightIntensity);
  dirLight.position.set(100, 100, 100);
  dirLight.castShadow = true;
  dirLight.shadow.camera.top = 100;
  dirLight.shadow.camera.bottom = -100;
  dirLight.shadow.camera.left = -100;
  dirLight.shadow.camera.right = 100;
  // dirLight.shadow.camera.far = 40;
  dirLight.shadow.camera.near = 0.1;
  dirLight.shadow.mapSize.width = shadowMapRes;
  dirLight.shadow.mapSize.height = shadowMapRes;
  dirLight.shadow.bias = -0.0004;
  dirLight.shadow.normalBias = 0.02;
  // progressiveSurfacemap.addObjectsToLightMap([dirLight])
  scene.add(dirLight);
  scene.add(dirLight.target);

  const groundMat = new MeshPhongMaterial({
    color: viewportSettings.ground,
    depthWrite: true,
  });
  const groundMesh = new Mesh(new PlaneGeometry(1000, 1000), groundMat);
  groundMesh.rotation.x = -Math.PI / 2;
  groundMesh.receiveShadow = true;
  scene.add(groundMesh);

  const orbitControls = new OrbitControls(camera, renderer.domElement);
  orbitControls.minDistance = 0;
  orbitControls.maxDistance = 100;
  orbitControls.target.set(0, 0, 0);
  orbitControls.update();

  const viewportGizmo = new ViewportGizmo(camera, renderer, {
    size: Math.round(Math.min(window.innerWidth, window.innerHeight) * 0.1),
  });
  viewportGizmo.attachControls(orbitControls);

  camera.position.z = 5;

  return {
    Camera: camera,
    OrbitControls: orbitControls,
    Renderer: renderer,
    Scene: scene,
    ViewerScene: viewerContainer,
    Ground: {
      Material: groundMat,
      Mesh: groundMesh,
    },
    Lighting: {
      DirLight: dirLight,
      HemiLight: hemiLight,
    },
    Composer: composer,
    PostProcessing: {
      SSAO: ssaoEffect,
      Bloom: bloomEffect,
      ToneMapping: toneMappingEffect,
      EffectPass: effectPass,
    },
    LabelRenderer: labelRenderer,
    Fog: fog,
    UpdateLoop: updateLoop,
    ViewportGizmo: viewportGizmo,
  };
}
