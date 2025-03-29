import { VRButton } from 'three/examples/jsm/webxr/VRButton.js';
import { XRControllerModelFactory } from 'three/examples/jsm/webxr/XRControllerModelFactory.js';
import { RepresentationType, WebSocketRepresentationManager } from "./websocket.js";
import { AdditiveBlending, BufferGeometry, CircleGeometry, Float32BufferAttribute, Group, Line, LineBasicMaterial, Matrix4, Mesh, MeshBasicMaterial, Object3D, Quaternion, Raycaster, RingGeometry, Vector3, XRGripSpace, XRTargetRaySpace } from 'three';
import { ThreeApp } from './three_app.js';
import { UpdateEntry, UpdateManager } from './update_manager.js';

interface Controller {
    target: XRTargetRaySpace;
    grip: XRGripSpace;
}

let INTERSECTION: Vector3;

const tempMatrix = new Matrix4();

// The XRControllerModelFactory will automatically fetch controller models
// that match what the user is holding as closely as possible. The models
// should be attached to the object returned from getControllerGrip in
// order to match the orientation of the held device.
const controllerModelFactory = new XRControllerModelFactory();

export class XRManager {

    threeApp: ThreeApp;

    updateManager: UpdateManager;

    controller1: Controller;

    controller2: Controller;

    marker: Mesh

    baseReferenceSpace: XRReferenceSpace

    raycaster: Raycaster;

    updater: UpdateEntry;

    constructor(threeApp: ThreeApp, representationManager: WebSocketRepresentationManager, updateManager: UpdateManager) {
        this.threeApp = threeApp;
        this.updateManager = updateManager;
        this.raycaster = new Raycaster();
        this.updater = {
            name: "XR Controller",
            loop: this.update
        }

        threeApp.Renderer.xr.addEventListener('sessionstart', this.sessionStarted);
        threeApp.Renderer.xr.addEventListener('sessionend', this.sessionEnd);

        this.controller1 = buildController(0, threeApp, representationManager, this.baseReferenceSpace);
        this.controller2 = buildController(1, threeApp, representationManager, this.baseReferenceSpace);

        document.body.appendChild(VRButton.createButton(threeApp.Renderer));
    }

    sessionStarted(): void {
        this.baseReferenceSpace = this.threeApp.Renderer.xr.getReferenceSpace();
        this.updateManager.addToUpdate(this.updater);

        if (!this.marker) {
            this.marker = new Mesh(
                new CircleGeometry(0.25, 32).rotateX(- Math.PI / 2),
                new MeshBasicMaterial({ color: 0xbcbcbc })
            );
            this.threeApp.Scene.add(this.marker);
        }
    }

    sessionEnd(): void {
        this.updateManager.removeFromUpdate(this.updater);

        if (this.marker) {
            this.threeApp.Scene.remove(this.marker);
            this.marker = null;
        }
    }

    intersectControllerUpdate(controller: XRTargetRaySpace): void {
        tempMatrix.identity().extractRotation(controller.matrixWorld);

        this.raycaster.ray.origin.setFromMatrixPosition(controller.matrixWorld);
        this.raycaster.ray.direction.set(0, 0, - 1).applyMatrix4(tempMatrix);

        const intersects = this.raycaster.intersectObjects([this.threeApp.Ground.Mesh]);
        if (intersects.length > 0) {
            INTERSECTION = intersects[0].point;
        }
    }

    update(): void {
        INTERSECTION = undefined;

        if (this.controller1.target.userData.isSelecting === true) {
            this.intersectControllerUpdate(this.controller1.target);
        } else if (this.controller2.target.userData.isSelecting === true) {
            this.intersectControllerUpdate(this.controller2.target);
        }

        if (INTERSECTION) this.marker.position.copy(INTERSECTION);

        this.marker.visible = INTERSECTION !== undefined;
    }

}

function buildController(index: number, threeApp: ThreeApp, representationManager: WebSocketRepresentationManager, baseReferenceSpace: XRReferenceSpace): Controller {
    function onSelectStart() {
        this.userData.isSelecting = true;
    }

    function onSelectEnd() {
        this.userData.isSelecting = false;

        if (INTERSECTION) {
            const offsetPosition = { x: - INTERSECTION.x, y: - INTERSECTION.y, z: - INTERSECTION.z, w: 1 };
            const offsetRotation = new Quaternion();
            const transform = new XRRigidTransform(offsetPosition, offsetRotation);
            const teleportSpaceOffset = baseReferenceSpace.getOffsetReferenceSpace(transform);

            threeApp.Renderer.xr.setReferenceSpace(teleportSpaceOffset);
        }
    }

    const controller = threeApp.Renderer.xr.getController(index);
    controller.addEventListener('selectstart', onSelectStart);
    controller.addEventListener('selectend', onSelectEnd);
    controller.addEventListener('connected', (event) => {
        const rep = buildControllerRepresentation(event.data);
        representationManager.AddRepresentation(RepresentationType.LeftHand, controller);
        controller.add(rep);
    });
    controller.addEventListener('disconnected', () => {
        representationManager.RemoveRepresentation(RepresentationType.LeftHand, controller);
        controller.remove(controller.children[0]);
    });
    threeApp.Scene.add(controller);

    const grip = threeApp.Renderer.xr.getControllerGrip(index);
    grip.add(controllerModelFactory.createControllerModel(grip));
    threeApp.Scene.add(grip);

    return {
        target: controller,
        grip: grip
    };
}

function buildControllerRepresentation(data: XRInputSource): Object3D {
    let geometry, material;

    console.log(data)
    switch (data.targetRayMode) {
        case 'tracked-pointer':
            geometry = new BufferGeometry();
            geometry.setAttribute('position', new Float32BufferAttribute([0, 0, 0, 0, 0, - 1], 3));
            geometry.setAttribute('color', new Float32BufferAttribute([0.5, 0.5, 0.5, 0, 0, 0], 3));

            material = new LineBasicMaterial({ vertexColors: true, blending: AdditiveBlending });

            return new Line(geometry, material);

        case 'gaze':
            geometry = new RingGeometry(0.02, 0.04, 32).translate(0, 0, - 1);
            material = new MeshBasicMaterial({ opacity: 0.5, transparent: true });
            return new Mesh(geometry, material);

        default:
            console.warn("unrecognized target ray mode: " + data.targetRayMode);
            return new Group();
    }
}


