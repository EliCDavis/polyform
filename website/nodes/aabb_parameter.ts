import { TransformControls } from 'three/examples/jsm/controls/TransformControls.js';
import { BoxHelper } from '../box.js';
import { FlowNode } from '@elicdavis/node-flow';
import { NodeManager } from '../node_manager.js';
import { Group, Vector3 } from 'three';
import { ThreeApp } from '../three_app.js';

interface BoxSideController {
    mesh: Group,
    control: TransformControls,
    helper: any,
}

export class AABBParameterNodeController {

    addControl(
        parameterData: any, 
        app: ThreeApp, 
        pos: { x: number; y: number; z: number; }, 
        showX: boolean, 
        showY: boolean, 
        showZ: boolean
    ): BoxSideController {
        const control = new TransformControls(app.Camera, app.Renderer.domElement);
        control.setMode('translate');
        control.setSpace("local");
        control.showX = showX;
        control.showY = showY;
        control.showZ = showZ;

        const controlMesh = new Group();
        app.ViewerScene.add(controlMesh);

        controlMesh.position.set(pos.x, pos.y, pos.z);

        const helper = control.getHelper();
        app.Scene.add(helper);
        control.attach(controlMesh);

        // control.visible = false;
        // control.enabled = false;

        control.addEventListener('dragging-changed', (event) => {
            app.OrbitControls.enabled = !event.value;

            if (app.OrbitControls.enabled) {
                this.recalcBounds();
            }
        });

        return {
            mesh: controlMesh,
            control: control,
            helper: helper,
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

    id: string;

    lightNode: FlowNode;

    controlMesh: Group;

    nodeManager: NodeManager;

    updating: boolean;

    box: BoxHelper;

    up: BoxSideController;

    down: BoxSideController;

    left: BoxSideController;

    right: BoxSideController;

    forward: BoxSideController;

    backward: BoxSideController;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData, app: ThreeApp) {
        this.lightNode = flowNode;
        this.lightNode.setTitle(parameterData.name);
        this.updating = false;
        this.nodeManager = nodeManager;
        this.id = id;

        const curVal = parameterData.currentValue;
        this.box = new BoxHelper(this.controlMesh);
        this.box.setBounds(
            new Vector3(
                curVal.center.x - curVal.extents.x,
                curVal.center.y - curVal.extents.y,
                curVal.center.z - curVal.extents.z,
            ),
            new Vector3(
                curVal.center.x + curVal.extents.x,
                curVal.center.y + curVal.extents.y,
                curVal.center.z + curVal.extents.z,
            )
        );
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

        this.up.helper.visible = false;
        this.up.helper.enabled = false;
        this.down.helper.visible = false;
        this.down.helper.enabled = false;
        this.left.helper.visible = false;
        this.left.helper.enabled = false;
        this.right.helper.visible = false;
        this.right.helper.enabled = false;
        this.forward.helper.visible = false;
        this.forward.helper.enabled = false;
        this.backward.helper.visible = false;
        this.backward.helper.enabled = false;
        this.box.visible = false;

        this.up.control.enabled = false;
        this.down.control.enabled = false;
        this.left.control.enabled = false;
        this.right.control.enabled = false;
        this.forward.control.enabled = false;
        this.backward.control.enabled = false;

        this.lightNode.addSelectListener(() => {
            this.box.visible = true;
            this.up.helper.visible = true;
            this.up.helper.enabled = true;
            this.down.helper.visible = true;
            this.down.helper.enabled = true;
            this.left.helper.visible = true;
            this.left.helper.enabled = true;
            this.right.helper.visible = true;
            this.right.helper.enabled = true;
            this.forward.helper.visible = true;
            this.forward.helper.enabled = true;
            this.backward.helper.visible = true;
            this.backward.helper.enabled = true;

            this.up.control.enabled = true;
            this.down.control.enabled = true;
            this.left.control.enabled = true;
            this.right.control.enabled = true;
            this.forward.control.enabled = true;
            this.backward.control.enabled = true;
        });

        this.lightNode.addUnselectListener(() => {
            this.box.visible = false;
            this.up.helper.visible = false;
            this.up.helper.enabled = false;
            this.down.helper.visible = false;
            this.down.helper.enabled = false;
            this.left.helper.visible = false;
            this.left.helper.enabled = false;
            this.right.helper.visible = false;
            this.right.helper.enabled = false;
            this.forward.helper.visible = false;
            this.forward.helper.enabled = false;
            this.backward.helper.visible = false;
            this.backward.helper.enabled = false;

            this.up.control.enabled = false;
            this.down.control.enabled = false;
            this.left.control.enabled = false;
            this.right.control.enabled = false;
            this.forward.control.enabled = false;
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
            new Vector3(
                center.x - extents.x,
                center.y - extents.y,
                center.z - extents.z,
            ),
            new Vector3(
                center.x + extents.x,
                center.y + extents.y,
                center.z + extents.z,
            )
        );

        this.lightNode.setProperty("min-x", center.x - extents.x);
        this.lightNode.setProperty("min-y", center.y - extents.y);
        this.lightNode.setProperty("min-z", center.z - extents.z);
        this.lightNode.setProperty("max-x", center.x + extents.x);
        this.lightNode.setProperty("max-y", center.y + extents.y);
        this.lightNode.setProperty("max-z", center.z + extents.z);
    }
}     