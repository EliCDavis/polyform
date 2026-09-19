import { Observable, Subject } from "rxjs";
import { Group, PerspectiveCamera, Scene } from "three";
import { TransformControls } from 'three/examples/jsm/controls/TransformControls.js';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';

export interface TRSValue {
    position: { x: number; y: number; z: number };
    rotation: { x: number; y: number; z: number; w: number };
    scale: { x: number; y: number; z: number };
}

export type TRSGizmoMode = "translate" | "rotate" | "scale";

export interface TRSGizmoConfig {
    camera: PerspectiveCamera;
    parent: Group;
    domElement: HTMLElement;
    scene: Scene;
    orbitControls: OrbitControls;
    initial?: TRSValue;
    mode?: TRSGizmoMode;
}

export class TRSGizmo {

    private mesh: Group;

    private controls: TransformControls;

    private helper: any;

    private change$: Subject<TRSValue>;

    constructor(config: TRSGizmoConfig) {
        this.change$ = new Subject<TRSValue>();

        this.controls = new TransformControls(config.camera, config.domElement);
        this.controls.setMode(config.mode ?? "translate");
        this.controls.setSpace("local");

        this.mesh = new Group();

        this.controls.addEventListener('dragging-changed', (event) => {
            config.orbitControls.enabled = !event.value;
            if (config.orbitControls.enabled) {
                this.change$.next(this.value());
            }
        });

        config.parent.add(this.mesh);
        if (config.initial) {
            this.setValue(config.initial);
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

    setMode(mode: TRSGizmoMode): void {
        this.controls.setMode(mode);
    }

    setValue(v: TRSValue): void {
        this.mesh.position.set(v.position.x, v.position.y, v.position.z);
        this.mesh.quaternion.set(v.rotation.x, v.rotation.y, v.rotation.z, v.rotation.w);
        this.mesh.scale.set(v.scale.x, v.scale.y, v.scale.z);
    }

    value(): TRSValue {
        const p = this.mesh.position;
        const q = this.mesh.quaternion;
        const s = this.mesh.scale;
        return {
            position: { x: p.x, y: p.y, z: p.z },
            rotation: { x: q.x, y: q.y, z: q.z, w: q.w },
            scale: { x: s.x, y: s.y, z: s.z },
        };
    }

    changes$(): Observable<TRSValue> {
        return this.change$.asObservable();
    }

    dispose(): void {
        this.setEnabled(false);
        this.change$.complete();
        this.mesh.removeFromParent();
        this.helper.removeFromParent();
        this.controls.detach();
        this.controls.dispose();
    }
}
