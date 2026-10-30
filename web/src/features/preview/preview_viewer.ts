import {
  Box3,
  Color,
  DirectionalLight,
  HemisphereLight,
  GridHelper,
  Group,
  PerspectiveCamera,
  Scene,
  Vector3,
  WebGLRenderer,
} from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader.js";
import { roomEnvironment } from "./environment";

// A small stand-alone viewer for one node output, kept apart from the main
// viewport so a preview never disturbs the producer scene.
export class PreviewViewer {
  private readonly renderer: WebGLRenderer;
  private readonly scene = new Scene();
  private readonly camera: PerspectiveCamera;
  private readonly controls: OrbitControls;
  private readonly model = new Group();
  private readonly grid = new GridHelper(10, 10, 0x777777, 0x8a8a8a);
  private readonly resize: ResizeObserver;
  private frame = 0;
  private latest = 0;

  constructor(private readonly container: HTMLElement) {
    this.renderer = new WebGLRenderer({ antialias: true });
    this.renderer.setPixelRatio(window.devicePixelRatio);
    container.appendChild(this.renderer.domElement);

    this.scene.background = new Color("#a0a0a0");
    this.scene.environment = roomEnvironment(this.renderer);
    this.scene.environmentIntensity = 0.1;
    this.scene.add(new HemisphereLight(0xffffff, 0x8d8d8d, 0.3));
    const sun = new DirectionalLight(0xffffff, 3);
    sun.position.set(5, 10, 7);
    this.scene.add(sun);
    this.scene.add(this.grid);
    this.scene.add(this.model);

    this.camera = new PerspectiveCamera(50, 1, 0.01, 10000);
    this.camera.position.set(5, 4, 5);
    this.controls = new OrbitControls(this.camera, this.renderer.domElement);

    this.resize = new ResizeObserver(() => this.fit());
    this.resize.observe(container);
    this.fit();
    this.render();
  }

  load(url: string): void {
    const request = ++this.latest;
    new GLTFLoader().load(url, (gltf) => {
      if (request !== this.latest) return;
      this.model.clear();
      this.model.add(gltf.scene);
      this.frameModel();
    });
  }

  dispose(): void {
    cancelAnimationFrame(this.frame);
    this.resize.disconnect();
    this.controls.dispose();
    this.scene.environment?.dispose();
    this.renderer.dispose();
    this.renderer.domElement.remove();
  }

  private frameModel(): void {
    const bounds = new Box3().setFromObject(this.model);
    if (bounds.isEmpty()) return;
    const center = bounds.getCenter(new Vector3());
    const size = bounds.getSize(new Vector3()).length();
    const span = Math.max(size, 0.01);
    this.grid.scale.setScalar(span / 10);
    this.grid.position.set(center.x, bounds.min.y, center.z);
    this.camera.near = span / 1000;
    this.camera.far = span * 100;
    this.camera.updateProjectionMatrix();
    this.camera.position.copy(center).add(new Vector3(0.8, 0.6, 0.8).multiplyScalar(span));
    this.controls.target.copy(center);
    this.controls.update();
  }

  private fit(): void {
    const width = Math.max(1, this.container.clientWidth);
    const height = Math.max(1, this.container.clientHeight);
    this.renderer.setSize(width, height, false);
    this.camera.aspect = width / height;
    this.camera.updateProjectionMatrix();
  }

  private render = (): void => {
    this.controls.update();
    this.renderer.render(this.scene, this.camera);
    this.frame = requestAnimationFrame(this.render);
  };
}
