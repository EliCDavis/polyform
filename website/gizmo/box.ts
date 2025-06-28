import { BoxHelper } from '../box.js';
import { Group, PerspectiveCamera, Scene, Vector3 } from 'three';
import { TransformGizmo } from '../gizmo/transform.js';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { Observable, Subject } from 'rxjs';

export interface AABB {
    center: {
        x: number,
        y: number,
        z: number,
    },
    extents: {
        x: number,
        y: number,
        z: number,
    }
}

export interface BoxGizmoConfig {
    camera: PerspectiveCamera;
    parent: Group;
    domElement: HTMLElement;
    scene: Scene;
    orbitControls: OrbitControls;
    initial: AABB;
}

export class BoxGizmo {

    private controlMesh: Group;

    private box: BoxHelper;

    private up: TransformGizmo;

    private down: TransformGizmo;

    private left: TransformGizmo;

    private right: TransformGizmo;

    private forward: TransformGizmo;

    private backward: TransformGizmo;

    private change$: Subject<AABB>;

    constructor(config: BoxGizmoConfig) {
        this.change$ = new Subject<AABB>();
        const aabb = config.initial;
        this.box = new BoxHelper(this.controlMesh);
        this.box.setBounds(
            new Vector3(
                aabb.center.x - aabb.extents.x,
                aabb.center.y - aabb.extents.y,
                aabb.center.z - aabb.extents.z,
            ),
            new Vector3(
                aabb.center.x + aabb.extents.x,
                aabb.center.y + aabb.extents.y,
                aabb.center.z + aabb.extents.z,
            )
        );

        config.parent.add(this.box);

        this.up = this.createControl(
            config, false, true, false,
            {
                x: aabb.center.x,
                y: aabb.center.y + aabb.extents.y,
                z: aabb.center.z
            },
        );

        this.down = this.createControl(
            config, false, true, false,
            {
                x: aabb.center.x,
                y: aabb.center.y - aabb.extents.y,
                z: aabb.center.z
            },
        );

        this.left = this.createControl(
            config, true, false, false,
            {
                x: aabb.center.x - aabb.extents.x,
                y: aabb.center.y,
                z: aabb.center.z
            },
        );

        this.right = this.createControl(
            config, true, false, false,
            {
                x: aabb.center.x + aabb.extents.x,
                y: aabb.center.y,
                z: aabb.center.z
            },
        );

        this.forward = this.createControl(
            config, false, false, true,
            {
                x: aabb.center.x,
                y: aabb.center.y,
                z: aabb.center.z + aabb.extents.z
            },
        );

        this.backward = this.createControl(
            config, false, false, true,
            {
                x: aabb.center.x,
                y: aabb.center.y,
                z: aabb.center.z - aabb.extents.z
            },
        );

        this.setEnabled(false);
    }

    setEnabled(enabled: boolean): void {
        this.box.visible = enabled;
        this.up.setEnabled(enabled);
        this.down.setEnabled(enabled);
        this.left.setEnabled(enabled);
        this.right.setEnabled(enabled);
        this.forward.setEnabled(enabled);
        this.backward.setEnabled(enabled);
    }

    aabb(): AABB {
        const extents = {
            x: Math.abs(this.right.x() - this.left.x()) / 2,
            y: Math.abs(this.up.y() - this.down.y()) / 2,
            z: Math.abs(this.forward.z() - this.backward.z()) / 2
        }
        return {
            extents,
            center: {
                x: this.left.x() + extents.x,
                y: this.down.y() + extents.y,
                z: this.backward.z() + extents.z,
            }
        };
    }

    private recalcBounds() {
        const aabb = this.aabb();
        this.box.setBounds(
            new Vector3(
                aabb.center.x - aabb.extents.x,
                aabb.center.y - aabb.extents.y,
                aabb.center.z - aabb.extents.z,
            ),
            new Vector3(
                aabb.center.x + aabb.extents.x,
                aabb.center.y + aabb.extents.y,
                aabb.center.z + aabb.extents.z,
            )
        );
    }

    setForward(value: number): void {
        this.forward.setZ(value);
        this.recalcBounds();
    }

    setBackwards(value: number): void {
        this.backward.setZ(value);
        this.recalcBounds();
    }

    setLeft(value: number): void {
        this.left.setX(value);
        this.recalcBounds();
    }

    setRight(value: number): void {
        this.right.setX(value);
        this.recalcBounds();
    }

    setUp(value: number): void {
        this.up.setY(value);
        this.recalcBounds();
    }

    setDown(value: number): void {
        this.down.setY(value);
        this.recalcBounds();
    }

    createControl(
        config: BoxGizmoConfig,
        showX: boolean,
        showY: boolean,
        showZ: boolean,
        position: { x: number, y: number, z: number, }
    ): TransformGizmo {
        const gizmo = new TransformGizmo({
            camera: config.camera,
            domElement: config.domElement,
            orbitControls: config.orbitControls,
            parent: config.parent,
            scene: config.scene,
            initialPosition: {
                x: position.x,
                y: position.y,
                z: position.z
            },
            hideX: !showX,
            hideY: !showY,
            hideZ: !showZ,
        });
        gizmo.position$().subscribe(() => {
            this.change$.next(this.aabb());
            this.recalcBounds();
        })
        return gizmo;
    }

    set(aabb: AABB) {
        this.up.setPosition(
            aabb.center.x,
            aabb.center.y + aabb.extents.y,
            aabb.center.z
        );

        this.down.setPosition(
            aabb.center.x,
            aabb.center.y - aabb.extents.y,
            aabb.center.z
        );

        this.left.setPosition(
            aabb.center.x - aabb.extents.x,
            aabb.center.y,
            aabb.center.z
        );

        this.right.setPosition(
            aabb.center.x + aabb.extents.x,
            aabb.center.y,
            aabb.center.z
        );

        this.forward.setPosition(
            aabb.center.x,
            aabb.center.y,
            aabb.center.z + aabb.extents.z
        );

        this.backward.setPosition(
            aabb.center.x,
            aabb.center.y,
            aabb.center.z - aabb.extents.z
        );

        this.recalcBounds();
    }

    aabb$(): Observable<AABB> {
        return this.change$.asObservable();
    }

    dispose(): void {
        this.setEnabled(false);
        this.box.removeFromParent();
        this.change$.complete();
        this.up.dispose();
        this.down.dispose();
        this.left.dispose();
        this.right.dispose();
        this.forward.dispose();
        this.backward.dispose();
    }

}     