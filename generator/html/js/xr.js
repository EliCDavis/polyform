import { VRButton } from 'three/addons/webxr/VRButton.js';
import { XRControllerModelFactory } from 'three/addons/webxr/XRControllerModelFactory.js';
import * as THREE from 'three';

let controller1, controller2;
let controllerGrip1, controllerGrip2;

let marker, floor, baseReferenceSpace, raycaster;

let INTERSECTION;
const tempMatrix = new THREE.Matrix4();


export const InitXR = (scene, renderer, updateManager, representationManager, ground) => {

    floor = ground;
    raycaster = new THREE.Raycaster();

    console.log(renderer.xr)
    renderer.xr.addEventListener('sessionstart', () => {
        baseReferenceSpace = renderer.xr.getReferenceSpace();

        if (!marker) {
            marker = new THREE.Mesh(
                new THREE.CircleGeometry(0.25, 32).rotateX(- Math.PI / 2),
                new THREE.MeshBasicMaterial({ color: 0xbcbcbc })
            );
            scene.add(marker);
        }
        updateManager.addToUpdate(intersectionUpdate);
    });

    renderer.xr.addEventListener('sessionend', () => {
        updateManager.removeFromUpdate(intersectionUpdate);
        scene.remove(marker);
        marker = null;
    });

    document.body.appendChild(VRButton.createButton(renderer));

    // controllers

    function onSelectStart() {
        this.userData.isSelecting = true;
    }

    function onSelectEnd() {
        this.userData.isSelecting = false;

        if (INTERSECTION) {
            const offsetPosition = { x: - INTERSECTION.x, y: - INTERSECTION.y, z: - INTERSECTION.z, w: 1 };
            const offsetRotation = new THREE.Quaternion();
            const transform = new XRRigidTransform(offsetPosition, offsetRotation);
            const teleportSpaceOffset = baseReferenceSpace.getOffsetReferenceSpace(transform);

            renderer.xr.setReferenceSpace(teleportSpaceOffset);
        }
    }

    controller1 = renderer.xr.getController(0);
    controller1.addEventListener('selectstart', onSelectStart);
    controller1.addEventListener('selectend', onSelectEnd);
    controller1.addEventListener('connected', (event) => {
        const rep = buildController(event.data);
        representationManager.AddRepresentation("left-hand", controller1);
        controller1.add(rep);
    });
    controller1.addEventListener('disconnected', () => {
        representationManager.RemoveRepresentation("left-hand", controller1);
        controller1.remove(controller1.children[0]);
    });
    scene.add(controller1);

    controller2 = renderer.xr.getController(1);
    controller2.addEventListener('selectstart', onSelectStart);
    controller2.addEventListener('selectend', onSelectEnd);
    controller2.addEventListener('connected', (event) => {
        const rep = buildController(event.data);
        representationManager.AddRepresentation("right-hand", controller2);
        controller2.add(rep);
    });
    controller2.addEventListener('disconnected', () => {
        representationManager.RemoveRepresentation("right-hand", controller2);
        controller2.remove(controller2.children[0]);
    });
    scene.add(controller2);

    // The XRControllerModelFactory will automatically fetch controller models
    // that match what the user is holding as closely as possible. The models
    // should be attached to the object returned from getControllerGrip in
    // order to match the orientation of the held device.
    const controllerModelFactory = new XRControllerModelFactory();

    controllerGrip1 = renderer.xr.getControllerGrip(0);
    controllerGrip1.add(controllerModelFactory.createControllerModel(controllerGrip1));
    scene.add(controllerGrip1);

    controllerGrip2 = renderer.xr.getControllerGrip(1);
    controllerGrip2.add(controllerModelFactory.createControllerModel(controllerGrip2));
    scene.add(controllerGrip2);

}

function buildController(data) {
    let geometry, material;

    console.log(data)
    switch (data.targetRayMode) {
        case 'tracked-pointer':
            geometry = new THREE.BufferGeometry();
            geometry.setAttribute('position', new THREE.Float32BufferAttribute([0, 0, 0, 0, 0, - 1], 3));
            geometry.setAttribute('color', new THREE.Float32BufferAttribute([0.5, 0.5, 0.5, 0, 0, 0], 3));

            material = new THREE.LineBasicMaterial({ vertexColors: true, blending: THREE.AdditiveBlending });

            return new THREE.Line(geometry, material);

        case 'gaze':
            geometry = new THREE.RingGeometry(0.02, 0.04, 32).translate(0, 0, - 1);
            material = new THREE.MeshBasicMaterial({ opacity: 0.5, transparent: true });
            return new THREE.Mesh(geometry, material);

        default:
            console.warn("unrecognized target ray mode: " + data.targetRayMode);
            return new THREE.Group();
    }
}

function intersectionUpdate() {
    INTERSECTION = undefined;

    if (controller1.userData.isSelecting === true) {
        tempMatrix.identity().extractRotation(controller1.matrixWorld);

        raycaster.ray.origin.setFromMatrixPosition(controller1.matrixWorld);
        raycaster.ray.direction.set(0, 0, - 1).applyMatrix4(tempMatrix);

        const intersects = raycaster.intersectObjects([floor]);

        if (intersects.length > 0) {
            INTERSECTION = intersects[0].point;
        }

    } else if (controller2.userData.isSelecting === true) {
        tempMatrix.identity().extractRotation(controller2.matrixWorld);

        raycaster.ray.origin.setFromMatrixPosition(controller2.matrixWorld);
        raycaster.ray.direction.set(0, 0, - 1).applyMatrix4(tempMatrix);

        const intersects = raycaster.intersectObjects([floor]);

        if (intersects.length > 0) {
            INTERSECTION = intersects[0].point;
        }
    }

    if (INTERSECTION) marker.position.copy(INTERSECTION);

    marker.visible = INTERSECTION !== undefined;
}