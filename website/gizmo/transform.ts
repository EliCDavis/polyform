import { Observable, Subject } from "rxjs";
import { Group, PerspectiveCamera, Scene, Vector3 } from "three";
import { TransformControls } from 'three/examples/jsm/controls/TransformControls.js';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';

export interface TransformGizmoConfig {
    camera: PerspectiveCamera;
    parent: Group;
    domElement: HTMLElement;
    scene: Scene;
    orbitControls: OrbitControls;
    initialPosition?: {
        x: number,
        y: number,
        z: number,
    }
}

export class TransformGizmo {

    mesh: Group;

    controls: TransformControls;

    helper: any;

    change$: Subject<Vector3>;

    constructor(config: TransformGizmoConfig) {
        this.change$ = new Subject<Vector3>();

        this.controls = new TransformControls(config.camera, config.domElement);
        this.controls.setMode('translate');
        this.controls.setSpace("local");

        this.mesh = new Group();

        this.controls.addEventListener('dragging-changed', (event) => {
            config.orbitControls.enabled = !event.value;
            if (!config.orbitControls.enabled) {
                return;
            }

            this.change$.next(this.mesh.position);
        });

        config.parent.add(this.mesh);
        if (config.initialPosition) {
            this.mesh.position.set(
                config.initialPosition.x,
                config.initialPosition.y,
                config.initialPosition.z
            );
        }

        this.helper = this.controls.getHelper();
        config.scene.add(this.helper)
        this.controls.attach(this.mesh);

        this.setEnabled(false);
    }

    setEnabled(enabled: boolean): void {
        this.helper.visible = enabled;
        this.controls.enabled = enabled;
    }

    setPosition(x: number, y: number, z: number): void {
        this.mesh.position.set(x, y, z);
    }

    position$(): Observable<Vector3> {
        return this.change$.asObservable();
    }

    dispose(): void {
        this.setEnabled(false);
        this.change$.complete();
        this.mesh.removeFromParent();
        this.helper.removeFromParent();
    }
}