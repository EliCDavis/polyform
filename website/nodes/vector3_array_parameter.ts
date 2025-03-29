import { TransformControls } from 'three/examples/jsm/controls/TransformControls.js';
import { FlowNode, GlobalWidgetFactory } from '@elicdavis/node-flow';
import { NodeManager } from '../node_manager';
import { Group } from 'three';
import { ThreeApp } from '../three_app';


export class Vector3ArrayParameterNodeController {

    lightNode: FlowNode;
    nodeManager: NodeManager;
    id: string;

    renderControls: boolean;

    allPositionControls: Array<TransformControls>;

    allPositionControlHelpersMeshes: Array<Group>;

    allPositionControlHelpers: Array<any>;

    app: ThreeApp;

    constructor(lightNode: FlowNode, nodeManager: NodeManager, id: string, parameterData, app: ThreeApp) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.app = app;
        this.allPositionControlHelpers = [];
        this.allPositionControls = new Array<TransformControls>();
        this.allPositionControlHelpersMeshes = [];
        this.renderControls = false;

        parameterData.currentValue?.forEach((ele) => {
            this.newPositionControl(ele);
        })

        this.lightNode = lightNode;
        this.lightNode.setTitle(parameterData.name);
        const addPointButton = GlobalWidgetFactory.create(this.lightNode, "button", {
            text: "Add Point",
            callback: () => {
                const paramData = this.buildParameterData();

                if (paramData.length > 0) {
                    const oldEle = paramData[paramData.length - 1]
                    const newEle = {
                        x: oldEle.x + 1,
                        y: oldEle.y,
                        z: oldEle.z,
                    }

                    paramData.push(newEle)
                } else {
                    paramData.push({ x: 0, y: 0, z: 0 })
                }


                this.nodeManager.nodeParameterChanged({
                    id: this.id,
                    data: paramData,
                    binary: false
                });
            }
        })

        this.lightNode.addWidget(addPointButton);


        this.lightNode.addSelectListener(() => {
            this.renderControls = true;
            this.updateControlRendering();
        });

        this.lightNode.addUnselectListener(() => {
            this.renderControls = false;
            this.updateControlRendering();
        });
    }

    updateControlRendering() {
        this.allPositionControlHelpers.forEach((v) => {
            v.visible = this.renderControls;
            v.enabled = this.renderControls;
        });
        this.allPositionControls.forEach((v) => {
            v.enabled = this.renderControls;
        });
    }

    buildParameterData() {
        const data = [];

        this.allPositionControlHelpersMeshes.forEach((ele) => {
            data.push({
                x: ele.position.x,
                y: ele.position.y,
                z: ele.position.z,
            })
        })

        return data
    }

    newPositionControl(pos: { x: number; y: number; z: number; }): void {
        const control = new TransformControls(this.app.Camera, this.app.Renderer.domElement);
        control.setMode('translate');
        control.setSpace("local");


        const mesh = new Group();

        control.addEventListener('dragging-changed', (event) => {
            this.app.OrbitControls.enabled = !event.value;

            if (this.app.OrbitControls.enabled) {
                this.nodeManager.nodeParameterChanged({
                    id: this.id,
                    data: this.buildParameterData(),
                    binary: false
                });
            }
        });

        const helper = control.getHelper();
        this.allPositionControlHelpers.push(helper);
        this.allPositionControls.push(control);
        this.allPositionControlHelpersMeshes.push(mesh);

        this.app.ViewerScene.add(mesh);
        this.app.Scene.add(helper);
        mesh.position.set(pos.x, pos.y, pos.z);
        control.attach(mesh);

        helper.visible = this.renderControls;
        // helper.enabled = this.renderControls;
        control.enabled = this.renderControls;
    }

    clearPositionControls(): void {
        this.allPositionControlHelpers.forEach((v) => {
            this.app.Scene.remove(v);
        });
        this.allPositionControlHelpersMeshes.forEach((v) => {
            this.app.ViewerScene.remove(v);
        })
        this.allPositionControlHelpers = [];
        this.allPositionControlHelpersMeshes = [];
        this.allPositionControls = [];
    }

    update(parameterData): void {
        this.clearPositionControls();
        parameterData.currentValue?.forEach((ele) => {
            this.newPositionControl(ele);
        })
    }
}