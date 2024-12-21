import * as THREE from 'three';
import { TransformControls } from 'three/addons/controls/TransformControls.js';
import { BoxHelper } from '../box.js';

export class AABBParameterNodeController {

    addControl(parameterData, app, pos, showX, showY, showZ) {
        const control = new TransformControls(app.Camera, app.Renderer.domElement);
        control.setMode('translate');
        control.setSpace("local");
        control.showX = showX;
        control.showY = showY;
        control.showZ = showZ;

        const controlMesh = new THREE.Group();
        app.ViewerScene.add(controlMesh);

        controlMesh.position.set(pos.x, pos.y, pos.z);

        app.Scene.add(control)
        control.attach(controlMesh);

        control.visible = false;
        control.enabled = false;

        control.addEventListener('dragging-changed', (event) => {
            app.OrbitControls.enabled = !event.value;

            if (app.OrbitControls.enabled) {
                this.recalcBounds();
            }
        });

        return {
            mesh: controlMesh,
            control: control,
        };
    }

    recalcBounds() {
        const extents = {
            x: Math.abs(this.right.mesh.position.x - this.left.mesh.position.x) / 2,
            y: Math.abs(this.up.mesh.position.y - this.down.mesh.position.y) / 2,
            z: Math.abs(this.forward.mesh.position.z - this.backward.mesh.position.z) / 2
        }
        const newData = {
            extents,
            center: {
                x: this.left.mesh.position.x + extents.x,
                y: this.down.mesh.position.y + extents.y,
                z: this.backward.mesh.position.z + extents.z,
            }
        }
        this.nodeManager.nodeParameterChanged({
            id: this.id,
            data: newData,
            binary: false
        });
    }

    constructor(lightNode, nodeManager, id, parameterData, app) {
        this.lightNode = lightNode;
        this.lightNode.setTitle(parameterData.name);      // const centerControl = new TransformControls(app.Camera, app.Renderer.domElement);
        this.updating = false;
        this.nodeManager = nodeManager;
        this.id = id;

        const curVal = parameterData.currentValue;
        this.box = new BoxHelper(this.controlMesh);
        this.box.setBounds(
            {
                x: curVal.center.x - curVal.extents.x,
                y: curVal.center.y - curVal.extents.y,
                z: curVal.center.z - curVal.extents.z,
            },
            {
                x: curVal.center.x + curVal.extents.x,
                y: curVal.center.y + curVal.extents.y,
                z: curVal.center.z + curVal.extents.z,
            }
        )
        this.lightNode.setProperty("min-x", curVal.center.x - curVal.extents.x);
        this.lightNode.setProperty("min-y", curVal.center.y - curVal.extents.y);
        this.lightNode.setProperty("min-z", curVal.center.z - curVal.extents.z);
        this.lightNode.setProperty("max-x", curVal.center.x + curVal.extents.x);
        this.lightNode.setProperty("max-y", curVal.center.y + curVal.extents.y);
        this.lightNode.setProperty("max-z", curVal.center.z + curVal.extents.z);
        this.lightNode.addPropertyChangeListener("min-x", this.propertyChange.bind(this));
        this.lightNode.addPropertyChangeListener("min-y", this.propertyChange.bind(this));
        this.lightNode.addPropertyChangeListener("min-z", this.propertyChange.bind(this));
        this.lightNode.addPropertyChangeListener("max-x", this.propertyChange.bind(this));
        this.lightNode.addPropertyChangeListener("max-y", this.propertyChange.bind(this));
        this.lightNode.addPropertyChangeListener("max-z", this.propertyChange.bind(this));
        app.ViewerScene.add(this.box);

        this.up = this.addControl(parameterData, app, {
            x: curVal.center.x,
            y: curVal.center.y + curVal.extents.y,
            z: curVal.center.z
        }, false, true, false);

        this.down = this.addControl(parameterData, app, {
            x: curVal.center.x,
            y: curVal.center.y - curVal.extents.y,
            z: curVal.center.z
        }, false, true, false);

        this.left = this.addControl(parameterData, app, {
            x: curVal.center.x - curVal.extents.x,
            y: curVal.center.y,
            z: curVal.center.z
        }, true, false, false);

        this.right = this.addControl(parameterData, app, {
            x: curVal.center.x + curVal.extents.x,
            y: curVal.center.y,
            z: curVal.center.z
        }, true, false, false);

        this.forward = this.addControl(parameterData, app, {
            x: curVal.center.x,
            y: curVal.center.y,
            z: curVal.center.z + curVal.extents.z
        }, false, false, true);

        this.backward = this.addControl(parameterData, app, {
            x: curVal.center.x,
            y: curVal.center.y,
            z: curVal.center.z - curVal.extents.z
        }, false, false, true);

        this.up.control.visible = false;
        this.up.control.enabled = false;
        this.down.control.visible = false;
        this.down.control.enabled = false;
        this.left.control.visible = false;
        this.left.control.enabled = false;
        this.right.control.visible = false;
        this.right.control.enabled = false;
        this.forward.control.visible = false;
        this.forward.control.enabled = false;
        this.backward.control.visible = false;
        this.backward.control.enabled = false;
        this.box.visible = false;

        this.lightNode.addSelectListener(() => {
            this.box.visible = true;
            this.up.control.visible = true;
            this.up.control.enabled = true;
            this.down.control.visible = true;
            this.down.control.enabled = true;
            this.left.control.visible = true;
            this.left.control.enabled = true;
            this.right.control.visible = true;
            this.right.control.enabled = true;
            this.forward.control.visible = true;
            this.forward.control.enabled = true;
            this.backward.control.visible = true;
            this.backward.control.enabled = true;
        });

        this.lightNode.addUnselectListener(() => {
            this.box.visible = false;
            this.up.control.visible = false;
            this.up.control.enabled = false;
            this.down.control.visible = false;
            this.down.control.enabled = false;
            this.left.control.visible = false;
            this.left.control.enabled = false;
            this.right.control.visible = false;
            this.right.control.enabled = false;
            this.forward.control.visible = false;
            this.forward.control.enabled = false;
            this.backward.control.visible = false;
            this.backward.control.enabled = false;
        });
    }

    propertyChange() {
        if (this.updating) {
            return
        }
        this.right.mesh.position.setX(this.lightNode.getProperty("max-x"));
        this.left.mesh.position.setX(this.lightNode.getProperty("min-x"));
        this.up.mesh.position.setY(this.lightNode.getProperty("max-y"));
        this.down.mesh.position.setY(this.lightNode.getProperty("min-y"));
        this.forward.mesh.position.setZ(this.lightNode.getProperty("max-z"));
        this.backward.mesh.position.setZ(this.lightNode.getProperty("min-z"));
        this.recalcBounds();
    }

    update(parameterData) {
        this.updating = true;
        const curVal = parameterData.currentValue;

        this.up.mesh.position.set(
            curVal.center.x,
            curVal.center.y + curVal.extents.y,
            curVal.center.z
        );

        this.down.mesh.position.set(
            curVal.center.x,
            curVal.center.y - curVal.extents.y,
            curVal.center.z
        );

        this.left.mesh.position.set(
            curVal.center.x - curVal.extents.x,
            curVal.center.y,
            curVal.center.z
        );

        this.right.mesh.position.set(
            curVal.center.x + curVal.extents.x,
            curVal.center.y,
            curVal.center.z
        );

        this.forward.mesh.position.set(
            curVal.center.x,
            curVal.center.y,
            curVal.center.z + curVal.extents.z
        );

        this.backward.mesh.position.set(
            curVal.center.x,
            curVal.center.y,
            curVal.center.z - curVal.extents.z
        );

        this.updateBox(curVal.center, curVal.extents);
        
        this.updating = false;
    }

    updateBox(center, extents) {
        this.box.setBounds(
            {
                x: center.x - extents.x,
                y: center.y - extents.y,
                z: center.z - extents.z,
            },
            {
                x: center.x + extents.x,
                y: center.y + extents.y,
                z: center.z + extents.z,
            }
        )

        this.lightNode.setProperty("min-x", center.x - extents.x);
        this.lightNode.setProperty("min-y", center.y - extents.y);
        this.lightNode.setProperty("min-z", center.z - extents.z);
        this.lightNode.setProperty("max-x", center.x + extents.x);
        this.lightNode.setProperty("max-y", center.y + extents.y);
        this.lightNode.setProperty("max-z", center.z + extents.z);
    }
}     