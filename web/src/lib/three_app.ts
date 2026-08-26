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
  // All effects are merged into this single EffectPass (the idiomatic
  // pmndrs pattern — one shader instead of one per effect). Per-effect
  // on/off must go through each effect's own blendMode, NOT this pass's
  // `.enabled`: with only one pass in the chain, disabling the pass
  // disables the whole chain, leaving nothing rendered to the screen at
  // all rather than falling back to an earlier stage.
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
  // Tone mapping is done by ToneMappingEffect below, not here. Measured
  // live: with renderer.toneMapping set to anything but NoToneMapping,
  // materials gamma-encode their output during RenderPass's scene render,
  // and that already-encoded result then lands in the composer's
  // intermediate render target — which is itself SRGBColorSpace, so the
  // GPU encodes it AGAIN on write. That double-encoding pushed every
  // tonemapping mode toward identical saturated output, which is exactly
  // why switching the "color grading" dropdown looked like it did nothing.
  renderer.toneMapping = NoToneMapping;
  renderer.xr.enabled = xrEnabled;
  renderer.setAnimationLoop(updateLoop.run.bind(updateLoop));

  // pmndrs/postprocessing instead of three's own examples/jsm postprocessing
  // stack: three's SSAOPass/UnrealBloomPass are unmaintained example code
  // (not part of three's stable API) and produced the AO artifacts we kept
  // fighting. This library also merges effects into fewer shader passes and
  // handles output color space automatically, so no OutputPass equivalent
  // is needed here.
  const composer = new EffectComposer(renderer, {
    // The canvas's own antialias flag is a no-op once postprocessing reads
    // from an offscreen render target, so AA has to come from the composer.
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

  // luminanceThreshold gates the bloom pass BEFORE blur/intensity ever run
  // (true regardless of mipmapBlur). RenderPass now renders with
  // NoToneMapping, so bloom sees true linear HDR scene values, not
  // something pre-compressed into [0, 1) — which is the correct order
  // (bloom is supposed to react to real overexposure, not a tonemapped
  // approximation of it). Tuned against our actual lighting (dir + hemi
  // ~1.1 intensity): low enough that ordinary lit surfaces still cross it,
  // high enough that shadowed areas don't.
  const bloomEffect = new BloomEffect({
    intensity: 0.3,
    luminanceThreshold: 0.35,
    luminanceSmoothing: 0.3,
    mipmapBlur: true,
  });
  // NOTE: mipmapBlurPass.radius is NOT the same knob as UnrealBloomPass's
  // old "radius" arg (which was a mild kernel-shape tweak with blur always
  // on) — here it's the mix factor between the blurred and unblurred
  // image, so 0 fully disables the blur and kills bloom entirely. Leave it
  // at the library default (0.85) so bloom actually spreads.

  const toneMappingEffect = new ToneMappingEffect({
    mode: ToneMappingMode.ACES_FILMIC,
  });

  // Merged into ONE EffectPass (not one each) — this is both the
  // idiomatic pmndrs pattern (fewer shader passes) and required for
  // correctness here: only the last pass added to the composer renders to
  // the screen, so if these were separate passes and the last one got
  // disabled, nothing would render at all instead of falling back to
  // showing the stage before it. Order matters: tone mapping must run
  // LAST, after SSAO/bloom have had a chance to operate on true linear
  // HDR values.
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
