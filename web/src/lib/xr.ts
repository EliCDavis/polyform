import { XRControllerModelFactory } from 'three/examples/jsm/webxr/XRControllerModelFactory.js';
import { AdditiveBlending, BufferGeometry, CircleGeometry, Float32BufferAttribute, Group, Line, LineBasicMaterial, Matrix4, Mesh, MeshBasicMaterial, Object3D, Quaternion, Raycaster, RingGeometry, Vector2, Vector3, XRGripSpace, XRTargetRaySpace } from 'three';
import type { NodeFlowGraph } from '@elicdavis/node-flow';
import type { TransformControls } from 'three/examples/jsm/controls/TransformControls.js';
import { RepresentationType, WebSocketRepresentationManager } from "./websocket";
import { ThreeApp } from './three_app';
import { UpdateEntry, UpdateManager } from './update_manager';
import { AnimationFrameQueue } from './xr_animation_frame_queue';
import { XRPagePanel } from './xr_page_panel';

const RAY_LENGTH = 5;
const WHEEL_PIXELS_PER_SECOND = 900;
const STICK_DEADZONE = 0.15;
const MIN_MOVE_PIXELS = 0.5;
const NODE_FLOW_ZOOM_STEP = 0.05;
const PAGE_PUSH_METERS_PER_SECOND = 1.5;

enum XRStandardButton {
    Trigger = 0,
    Squeeze = 1,
    Touchpad = 2,
    Thumbstick = 3,
    AX = 4,
    BY = 5,
}

enum XRStandardAxis {
    TouchpadX = 0,
    TouchpadY = 1,
    ThumbstickX = 2,
    ThumbstickY = 3,
}

const controllerModelFactory = new XRControllerModelFactory();

interface Controller {
    target: XRTargetRaySpace;
    grip: XRGripSpace;
    source: XRInputSource | null;
    pointer: Object3D | null;
    teleporting: boolean;
    teleportPoint: Vector3 | null;
    pressing: boolean;
    menuPressed: boolean;
    uv: Vector2 | null;
    bar: { point: Vector3, distance: number } | null;
    movingPage: boolean;
    gizmo: TransformControls | null;
}

export interface XRNodeCanvas {
    graph: NodeFlowGraph;
    canvas: HTMLCanvasElement;
}

export class XRManager {
    #threeApp: ThreeApp;
    #updateManager: UpdateManager;
    #representationManager: WebSocketRepresentationManager;
    #nodeCanvas: XRNodeCanvas | null;
    #frames = new AnimationFrameQueue();
    #raycaster = new Raycaster();
    #rotation = new Matrix4();
    #controllers: Array<Controller>;
    #page: XRPagePanel | null = null;
    #hidden: Array<{ element: HTMLElement, display: string }> = [];
    #marker: Mesh | null = null;
    #baseReferenceSpace: XRReferenceSpace | null = null;
    #lastUV: Vector2 | null = null;
    #pinchDistance: number | null = null;
    #updater: UpdateEntry;

    constructor(threeApp: ThreeApp, representationManager: WebSocketRepresentationManager, updateManager: UpdateManager, nodeCanvas: XRNodeCanvas | null) {
        this.#threeApp = threeApp;
        this.#updateManager = updateManager;
        this.#representationManager = representationManager;
        this.#nodeCanvas = nodeCanvas;
        this.#updater = {
            name: "XR Controller",
            loop: this.#update,
        };

        threeApp.Renderer.xr.addEventListener('sessionstart', this.#sessionStarted);
        threeApp.Renderer.xr.addEventListener('sessionend', this.#sessionEnded);

        this.#controllers = [this.#buildController(0), this.#buildController(1)];
    }

    #sessionStarted = (): void => {
        this.#baseReferenceSpace = this.#threeApp.Renderer.xr.getReferenceSpace();
        this.#updateManager.addToUpdate(this.#updater);

        if (!this.#marker) {
            this.#marker = new Mesh(
                new CircleGeometry(0.25, 32).rotateX(- Math.PI / 2),
                new MeshBasicMaterial({ color: 0xbcbcbc })
            );
            this.#marker.visible = false;
            this.#threeApp.Scene.add(this.#marker);
        }

        this.#hideViewport();
        this.#page = new XRPagePanel(document.body);
        this.#threeApp.Scene.add(this.#page.mesh);
    };

    #sessionEnded = (): void => {
        this.#updateManager.removeFromUpdate(this.#updater);
        this.#baseReferenceSpace = null;

        if (this.#marker) {
            this.#threeApp.Scene.remove(this.#marker);
            this.#marker = null;
        }

        for (const controller of this.#controllers) {
            controller.teleporting = false;
            controller.pressing = false;
            controller.uv = null;
            controller.bar = null;
            controller.movingPage = false;
            if (controller.gizmo) {
                controller.gizmo.pointerUp(null);
                controller.gizmo = null;
            }
        }
        this.#lastUV = null;
        this.#pinchDistance = null;

        this.#page?.dispose();
        this.#page = null;
        this.#restoreViewport();
    };

    #hideViewport(): void {
        const viewer = document.getElementById('three-viewer-container');
        const pane = viewer?.parentElement;
        if (!pane) {
            return;
        }
        for (const element of [pane, pane.nextElementSibling]) {
            if (element instanceof HTMLElement) {
                this.#hidden.push({ element, display: element.style.display });
                element.style.display = 'none';
            }
        }
        this.#fitNodeCanvas();
    }

    #restoreViewport(): void {
        this.#hidden.forEach(({ element, display }) => element.style.display = display);
        this.#hidden = [];
        this.#fitNodeCanvas();
    }

    // node-flow sizes its canvas from a ResizeObserver, which doesn't fire while presenting.
    #fitNodeCanvas(): void {
        const canvas = this.#nodeCanvas?.canvas;
        const parent = canvas?.parentElement;
        if (!canvas || !parent) {
            return;
        }
        const rect = parent.getBoundingClientRect();
        canvas.width = rect.width;
        canvas.height = rect.height;
        this.#nodeCanvas?.graph.requestRender();
    }

    #aim(controller: XRTargetRaySpace): void {
        this.#rotation.identity().extractRotation(controller.matrixWorld);
        this.#raycaster.ray.origin.setFromMatrixPosition(controller.matrixWorld);
        this.#raycaster.ray.direction.set(0, 0, - 1).applyMatrix4(this.#rotation);
    }

    #update = (delta: number): void => {
        const page = this.#page;
        if (this.#frames.flush()) {
            page?.requestRedraw();
        }
        const viewer = this.#threeApp.Renderer.xr.getCamera();
        page?.update(viewer);
        const eye = viewer.getWorldPosition(new Vector3());

        let teleportPoint: Vector3 | null = null;
        let barHovered = false;

        for (const controller of this.#controllers) {
            this.#aim(controller.target);
            let length = RAY_LENGTH;

            if (page && controller.movingPage) {
                const stick = controller.source?.gamepad?.axes[XRStandardAxis.ThumbstickY] ?? 0;
                if (Math.abs(stick) > STICK_DEADZONE) {
                    page.pushMove(-stick * PAGE_PUSH_METERS_PER_SECOND * delta);
                }
                page.moveAlong(this.#raycaster.ray, eye);
            }

            const hit = page?.intersect(this.#raycaster) ?? null;
            controller.uv = hit?.uv ?? null;
            if (hit) {
                length = hit.distance;
            }

            controller.bar = null;
            const bar = page?.intersectBar(this.#raycaster) ?? null;
            if (bar && (!hit || bar.distance < hit.distance) && !controller.pressing) {
                controller.bar = bar;
                controller.uv = null;
                length = bar.distance;
                barHovered = true;
            }

            if (controller.gizmo) {
                this.#aimGizmo(controller.gizmo);
                controller.gizmo.pointerMove(null);
            }

            controller.teleportPoint = null;
            if (controller.teleporting) {
                const intersects = this.#raycaster.intersectObject(this.#threeApp.Ground.Mesh);
                if (intersects.length > 0) {
                    controller.teleportPoint = intersects[0].point;
                    teleportPoint = controller.teleportPoint;
                    length = intersects[0].distance;
                }
            }

            if (controller.pointer) {
                controller.pointer.scale.z = length;
            }
        }

        if (this.#marker) {
            this.#marker.visible = teleportPoint !== null;
            if (teleportPoint) {
                this.#marker.position.copy(teleportPoint);
            }
        }

        page?.setBarHovered(barHovered);
        this.#hoverGizmos();

        if (page) {
            this.#updatePageInput(page, delta);
        }
    };

    #gizmos(): Array<TransformControls> {
        const gizmos: Array<TransformControls> = [];
        for (const child of this.#threeApp.Scene.children) {
            const root = child as Object3D & { isTransformControlsRoot?: boolean, controls?: TransformControls };
            if (root.isTransformControlsRoot && root.controls?.enabled) {
                gizmos.push(root.controls);
            }
        }
        return gizmos;
    }

    #aimGizmo(gizmo: TransformControls): void {
        gizmo.getRaycaster().ray.copy(this.#raycaster.ray);
    }

    #hoverGizmos(): void {
        if (this.#controllers.some((c) => c.gizmo)) {
            return;
        }
        const pointing = this.#controllers.find((c) => !c.uv && !c.bar && !c.pressing && !c.movingPage);
        for (const gizmo of this.#gizmos()) {
            if (!pointing) {
                gizmo.axis = null;
                continue;
            }
            this.#aim(pointing.target);
            this.#aimGizmo(gizmo);
            gizmo.pointerHover(null);
        }
    }

    #grabGizmo(controller: Controller): boolean {
        for (const gizmo of this.#gizmos()) {
            this.#aim(controller.target);
            this.#aimGizmo(gizmo);
            gizmo.pointerHover(null);
            if (gizmo.axis === null) {
                continue;
            }
            gizmo.pointerDown(null);
            if (gizmo.dragging) {
                controller.gizmo = gizmo;
                return true;
            }
        }
        return false;
    }

    #updatePageInput(page: XRPagePanel, delta: number): void {
        if (this.#pinchDistance !== null) {
            this.#updatePinch(page);
            return;
        }

        const active =
            this.#controllers.find((c) => c.pressing) ??
            this.#controllers.find((c) => c.uv !== null);

        for (const controller of this.#controllers) {
            const pressed = controller.source?.gamepad?.buttons[XRStandardButton.AX]?.pressed ?? false;
            const justPressed = pressed && !controller.menuPressed;
            controller.menuPressed = pressed;
            if (justPressed && controller === active && controller.uv) {
                page.contextMenu(controller.uv);
            }
        }

        const uv = active?.uv ?? null;
        if (!active || !uv) {
            this.#lastUV = null;
            page.leave();
            return;
        }

        if (!this.#lastUV || page.pixelDistance(this.#lastUV, uv) > MIN_MOVE_PIXELS) {
            this.#lastUV = uv.clone();
            page.move(uv);
        }

        const stick = active.source?.gamepad?.axes[XRStandardAxis.ThumbstickY] ?? 0;
        if (Math.abs(stick) > STICK_DEADZONE) {
            page.wheel(uv, stick * WHEEL_PIXELS_PER_SECOND * delta);
        }
    }

    #updatePinch(page: XRPagePanel): void {
        const [a, b] = this.#controllers.map((c) => c.uv);
        const nodeCanvas = this.#nodeCanvas;
        if (!a || !b || !nodeCanvas || this.#pinchDistance === null) {
            return;
        }
        const distance = page.pixelDistance(a, b);
        if (this.#pinchDistance > 0 && distance > 0) {
            const rect = nodeCanvas.canvas.getBoundingClientRect();
            const mid = page.client(a).add(page.client(b)).multiplyScalar(0.5);
            const factor = distance / this.#pinchDistance;
            nodeCanvas.graph.zoom((factor - 1) / NODE_FLOW_ZOOM_STEP, { x: mid.x - rect.left, y: mid.y - rect.top });
            nodeCanvas.graph.requestRender();
        }
        this.#pinchDistance = distance;
    }

    #onNodeCanvas(uv: Vector2): boolean {
        return this.#page !== null && this.#page.elementAt(uv) === this.#nodeCanvas?.canvas;
    }

    #teleport(point: Vector3): void {
        if (!this.#baseReferenceSpace) {
            return;
        }
        const offsetPosition = { x: - point.x, y: - point.y, z: - point.z, w: 1 };
        const transform = new XRRigidTransform(offsetPosition, new Quaternion());
        this.#threeApp.Renderer.xr.setReferenceSpace(this.#baseReferenceSpace.getOffsetReferenceSpace(transform));
    }

    #buildController(index: number): Controller {
        const xr = this.#threeApp.Renderer.xr;
        const target = xr.getController(index);
        const grip = xr.getControllerGrip(index);
        const controller: Controller = {
            target,
            grip,
            source: null,
            pointer: null,
            teleporting: false,
            teleportPoint: null,
            pressing: false,
            menuPressed: false,
            uv: null,
            bar: null,
            movingPage: false,
            gizmo: null,
        };

        target.addEventListener('selectstart', () => {
            const page = this.#page;
            if (page && controller.bar && !page.moving && page.mesh.parent === this.#threeApp.Scene) {
                controller.movingPage = true;
                page.beginMove(controller.bar.point, controller.bar.distance);
                return;
            }
            if (!page || !controller.uv) {
                if (!this.#grabGizmo(controller)) {
                    controller.teleporting = true;
                }
                return;
            }

            const other = this.#controllers.find((c) => c !== controller && c.pressing);
            controller.pressing = true;
            if (other?.uv && this.#onNodeCanvas(other.uv) && this.#onNodeCanvas(controller.uv)) {
                page.cancelPress();
                this.#pinchDistance = page.pixelDistance(other.uv, controller.uv);
                return;
            }
            if (other) {
                controller.pressing = false;
                return;
            }

            this.#lastUV = controller.uv.clone();
            page.down(controller.uv);
        });

        target.addEventListener('selectend', () => {
            if (controller.movingPage) {
                controller.movingPage = false;
                this.#page?.endMove();
                return;
            }
            if (controller.gizmo) {
                this.#aim(controller.target);
                this.#aimGizmo(controller.gizmo);
                controller.gizmo.pointerUp(null);
                controller.gizmo = null;
                return;
            }
            if (controller.pressing) {
                controller.pressing = false;
                if (this.#pinchDistance !== null) {
                    this.#pinchDistance = null;
                    this.#controllers.forEach((c) => c.pressing = false);
                    return;
                }
                this.#page?.up(controller.uv);
                return;
            }
            controller.teleporting = false;
            if (controller.teleportPoint) {
                this.#teleport(controller.teleportPoint);
            }
        });

        target.addEventListener('squeezestart', () => {
            if (this.#page && controller.uv) {
                target.attach(this.#page.mesh);
            }
        });

        target.addEventListener('squeezeend', () => {
            if (this.#page?.mesh.parent === target) {
                this.#threeApp.Scene.attach(this.#page.mesh);
            }
        });

        let representation: number = RepresentationType.LeftHand;
        target.addEventListener('connected', (event) => {
            const source = event.data as XRInputSource;
            controller.source = source;
            controller.pointer = buildControllerRepresentation(source);
            target.add(controller.pointer);
            representation = source.handedness === 'right' ? RepresentationType.RightHand : RepresentationType.LeftHand;
            this.#representationManager.AddRepresentation(representation, target);
        });

        target.addEventListener('disconnected', () => {
            this.#representationManager.RemoveRepresentation(representation, target);
            if (controller.pointer) {
                target.remove(controller.pointer);
            }
            controller.pointer = null;
            controller.source = null;
        });

        this.#threeApp.Scene.add(target);

        grip.add(controllerModelFactory.createControllerModel(grip));
        this.#threeApp.Scene.add(grip);

        return controller;
    }
}

function buildControllerRepresentation(data: XRInputSource): Object3D {
    switch (data.targetRayMode) {
        case 'tracked-pointer': {
            const geometry = new BufferGeometry();
            geometry.setAttribute('position', new Float32BufferAttribute([0, 0, 0, 0, 0, - 1], 3));
            geometry.setAttribute('color', new Float32BufferAttribute([0.5, 0.5, 0.5, 0, 0, 0], 3));
            const material = new LineBasicMaterial({ vertexColors: true, blending: AdditiveBlending });
            return new Line(geometry, material);
        }

        case 'gaze': {
            const geometry = new RingGeometry(0.02, 0.04, 32).translate(0, 0, - 1);
            const material = new MeshBasicMaterial({ opacity: 0.5, transparent: true });
            return new Mesh(geometry, material);
        }

        default:
            console.warn("unrecognized target ray mode: " + data.targetRayMode);
            return new Group();
    }
}
