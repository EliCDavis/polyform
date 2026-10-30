import {
  Box3,
  Color,
  DirectionalLight,
  HemisphereLight,
  Group,
  Material,
  Mesh,
  PerspectiveCamera,
  Scene,
  Vector3,
  WebGLRenderer,
} from "three";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader.js";
import { roomEnvironment } from "./environment";

const WIDTH = 180;
const HEIGHT = 135;

// Above this a thumbnail costs more than it tells; the node keeps its Open button.
export const THUMBNAIL_BYTE_LIMIT = 8 * 1024 * 1024;

interface Job {
  url: string;
  resolve: (dataUrl: string | null) => void;
  reject: (err: unknown) => void;
}

const CACHE_LIMIT = 300;

// One offscreen renderer draws every node thumbnail, one at a time. A newer
// request for the same key replaces one still waiting, so a slider drag
// renders the final state rather than every step. Results are kept by URL,
// which carries the output version, so switching tabs redraws nothing.
class ThumbnailRenderer {
  private renderer: WebGLRenderer | null = null;
  private readonly scene = new Scene();
  private readonly camera = new PerspectiveCamera(40, WIDTH / HEIGHT, 0.01, 10000);
  private readonly model = new Group();
  private readonly loader = new GLTFLoader();
  private readonly queue = new Map<string, Job>();
  private readonly cache = new Map<string, string | null>();
  private running = false;

  // Starts held: the producer view releases it once the first model has loaded.
  private held = true;

  hold(): void {
    this.held = true;
  }

  release(): void {
    this.held = false;
    void this.drain();
  }

  request(key: string, url: string): Promise<string | null> {
    if (this.cache.has(url)) {
      return Promise.resolve(this.cache.get(url)!);
    }
    return new Promise((resolve, reject) => {
      this.queue.get(key)?.resolve(null);
      this.queue.set(key, { url, resolve, reject });
      void this.drain();
    });
  }

  private remember(url: string, dataUrl: string | null): void {
    this.cache.set(url, dataUrl);
    if (this.cache.size > CACHE_LIMIT) {
      this.cache.delete(this.cache.keys().next().value as string);
    }
  }

  private async drain(): Promise<void> {
    if (this.running || this.held) return;
    this.running = true;
    try {
      while (this.queue.size > 0 && !this.held) {
        const [key, job] = this.queue.entries().next().value as [string, Job];
        this.queue.delete(key);
        try {
          const dataUrl = this.cache.has(job.url) ? this.cache.get(job.url)! : await this.render(job.url);
          this.remember(job.url, dataUrl);
          job.resolve(dataUrl);
        } catch (err) {
          job.reject(err);
        }
      }
    } finally {
      this.running = false;
    }
  }

  private async render(url: string): Promise<string | null> {
    const response = await fetch(url);
    if (!response.ok) throw new Error(`${response.status} fetching ${url}`);
    const bytes = await response.arrayBuffer();
    if (bytes.byteLength > THUMBNAIL_BYTE_LIMIT) return null;

    const gltf = await this.loader.parseAsync(bytes, "");
    const renderer = this.ensureRenderer();
    this.model.add(gltf.scene);
    try {
      this.frame();
      renderer.render(this.scene, this.camera);
      return renderer.domElement.toDataURL("image/png");
    } finally {
      this.model.remove(gltf.scene);
      gltf.scene.traverse((o) => {
        if (o instanceof Mesh) {
          o.geometry.dispose();
          const materials = Array.isArray(o.material) ? o.material : [o.material];
          materials.forEach((m: Material) => m.dispose());
        }
      });
    }
  }

  private ensureRenderer(): WebGLRenderer {
    if (this.renderer) return this.renderer;
    this.renderer = new WebGLRenderer({ antialias: true, preserveDrawingBuffer: true });
    this.renderer.setPixelRatio(1);
    this.renderer.setSize(WIDTH, HEIGHT, false);
    this.scene.background = new Color("#a0a0a0");
    this.scene.environment = roomEnvironment(this.renderer);
    this.scene.environmentIntensity = 0.1;
    this.scene.add(new HemisphereLight(0xffffff, 0x8d8d8d, 0.3));
    const sun = new DirectionalLight(0xffffff, 3);
    sun.position.set(5, 10, 7);
    this.scene.add(sun);
    this.scene.add(this.model);
    return this.renderer;
  }

  private frame(): void {
    const bounds = new Box3().setFromObject(this.model);
    const center = bounds.isEmpty() ? new Vector3() : bounds.getCenter(new Vector3());
    const span = bounds.isEmpty() ? 1 : Math.max(bounds.getSize(new Vector3()).length(), 0.01);
    this.camera.near = span / 1000;
    this.camera.far = span * 100;
    this.camera.position.copy(center).add(new Vector3(0.8, 0.6, 0.8).multiplyScalar(span));
    this.camera.lookAt(center);
    this.camera.updateProjectionMatrix();
  }
}

export const thumbnails = new ThumbnailRenderer();
