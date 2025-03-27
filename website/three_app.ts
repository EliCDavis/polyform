import { ACESFilmicToneMapping, Camera, Color, DirectionalLight, Fog, Group, HemisphereLight, Mesh, MeshPhongMaterial, PCFSoftShadowMap, PerspectiveCamera, PlaneGeometry, Scene, Vector2, WebGLRenderer } from "three";
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { CSS2DRenderer } from 'three/examples/jsm/renderers/CSS2DRenderer.js';
// import { RoomEnvironment } from 'three/examples/jsm/environments/RoomEnvironment.js';

import { EffectComposer } from 'three/examples/jsm/postprocessing/EffectComposer.js';
import { RenderPass } from 'three/examples/jsm/postprocessing/RenderPass.js';
import { UnrealBloomPass } from 'three/examples/jsm/postprocessing/UnrealBloomPass.js';
import { OutputPass } from 'three/examples/jsm/postprocessing/OutputPass.js';
import { ViewportSettings } from "./ViewportSettings";
import { UpdateManager } from "./update-manager";

// https://threejs.org/examples/?q=Directional#webgl_lights_hemisphere
// https://threejs.org/examples/#webgl_geometry_spline_editor

export interface ThreeAppLighting {
    DirLight: DirectionalLight;
    HemiLight: HemisphereLight
}

export interface ThreeAppGround {
    Mesh: Mesh;
    Material: MeshPhongMaterial;
}

export interface ThreeApp {
    Camera: PerspectiveCamera
    Renderer: WebGLRenderer
    LabelRenderer: CSS2DRenderer,
    OrbitControls: OrbitControls
    Scene: Scene
    ViewerScene: Group,
    Ground: ThreeAppGround,
    Lighting: ThreeAppLighting,
    Composer: EffectComposer,
    Fog: Fog,
}

export function CreateThreeApp(
    container: HTMLElement, 
    viewportSettings: ViewportSettings,
    updateLoop: UpdateManager,
    antiAlias: boolean,
    xrEnabled: boolean
): ThreeApp {
    // progressiveSurfacemap.addObjectsToLightMap([groundMesh])

    // const environment = new RoomEnvironment(renderer);
    // const pmremGenerator = new THREE.PMREMGenerator(renderer);
    // scene.environment = pmremGenerator.fromScene( environment ).texture;

    const shadowMapRes = 4098, lightMapRes = 4098, lightCount = 8;

    const camera = new PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);
    camera.position.set(0, 2, 3);

    const scene = new Scene();
    scene.background = new Color(viewportSettings.background);

    // scene.background = textureEquirec;
    const fog = new Fog(viewportSettings.fog.color, viewportSettings.fog.near, viewportSettings.fog.far);
    scene.fog = fog;

    const viewerContainer = new Group();
    scene.add(viewerContainer);

    const threeCanvas = document.getElementById("three-canvas");
    const renderer = new WebGLRenderer({
        canvas: threeCanvas,
        antialias: antiAlias
    });
    renderer.setPixelRatio(window.devicePixelRatio);
    // renderer.setSize(threeCanvas.clientWidth, threeCanvas.clientHeight, false);
    renderer.shadowMap.enabled = true;
    renderer.shadowMap.type = PCFSoftShadowMap; // default THREE.PCFShadowMap
    renderer.toneMapping = ACESFilmicToneMapping;
    renderer.toneMappingExposure = 1;
    renderer.xr.enabled = xrEnabled;
    renderer.setAnimationLoop(updateLoop.run.bind(updateLoop))

    const renderScene = new RenderPass(scene, camera);
    const bloomPass = new UnrealBloomPass(new Vector2(window.innerWidth, window.innerHeight), .3, 0., 1.01);
    const outputPass = new OutputPass();

    const composer = new EffectComposer(renderer);
    composer.addPass(renderScene);
    composer.addPass(bloomPass);
    composer.addPass(outputPass);

    // progressive lightmap
    // const progressiveSurfacemap = new ProgressiveLightMap(renderer, lightMapRes);

    const labelRenderer = new CSS2DRenderer();
    // labelRenderer.setSize(threeCanvas.clientWidth, threeCanvas.clientHeight, false);
    labelRenderer.domElement.style.position = 'absolute';
    labelRenderer.domElement.style.top = '0px';
    labelRenderer.domElement.style.pointerEvents = 'none';
    container.appendChild(labelRenderer.domElement);

    const hemiLight = new HemisphereLight(viewportSettings.lighting, 0x8d8d8d, 1);
    hemiLight.position.set(0, 20, 0);
    scene.add(hemiLight);

    const dirLight = new DirectionalLight(viewportSettings.lighting, 1);
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

    const groundMat = new MeshPhongMaterial({ color: viewportSettings.ground, depthWrite: true });
    const groundMesh = new Mesh(new PlaneGeometry(1000, 1000), groundMat);
    groundMesh.rotation.x = - Math.PI / 2;
    groundMesh.receiveShadow = true;
    scene.add(groundMesh);

    const orbitControls = new OrbitControls(camera, renderer.domElement);
    orbitControls.minDistance = 0;
    orbitControls.maxDistance = 100;
    orbitControls.target.set(0, 0, 0);
    orbitControls.update();

    camera.position.z = 5;

    return {
        Camera: camera,
        OrbitControls: orbitControls,
        Renderer: renderer,
        Scene: scene,
        ViewerScene: viewerContainer,
        Ground: {
            Material: groundMat,
            Mesh: groundMesh
        },
        Lighting: {
            DirLight: dirLight,
            HemiLight: hemiLight
        },
        Composer: composer,
        LabelRenderer: labelRenderer,
        Fog: fog
    };
}