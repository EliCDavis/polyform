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
    hideX?: boolean;
    hideY?: boolean;
    hideZ?: boolean;
}

export class TransformGizmo {

    private mesh: Group;

    private controls: TransformControls;

    private helper: any;

    private change$: Subject<Vector3>;

    constructor(config: TransformGizmoConfig) {
        this.change$ = new Subject<Vector3>();

        this.controls = new TransformControls(config.camera, config.domElement);
        this.controls.setMode('translate');
        this.controls.setSpace("local");

        this.controls.showX = !(config.hideX === true);
        this.controls.showY = !(config.hideY === true);
        this.controls.showZ = !(config.hideZ === true);

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

    setX(x: number): void {
        this.mesh.position.setX(x);
    }

    setY(y: number): void {
        this.mesh.position.setY(y);
    }

    setZ(z: number): void {
        this.mesh.position.setZ(z);
    }

    x(): number {
        return this.mesh.position.x;
    }

    y(): number {
        return this.mesh.position.y;
    }

    z(): number {
        return this.mesh.position.z;
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