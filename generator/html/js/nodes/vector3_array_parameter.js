import { TransformControls } from 'three/addons/controls/TransformControls.js';
import * as THREE from 'three';


export class NodeVector3ArryParameter {
    constructor(lightNode, nodeManager, id, parameterData, app) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.app = app;
        this.scene = app.ViewerScene;
        this.allPositionControls = [];
        this.allPositionControlsMeshes = [];
        this.renderControls = false;

        parameterData.currentValue.forEach((ele) => {
            this.newPositionControl(ele);
        })

        this.lightNode = lightNode;
        this.lightNode.title = parameterData.name;
        this.lightNode.addWidget("button", "Add Point", "", () => {
            const paramData = this.buildParameterData();

            const oldEle = paramData[paramData.length - 1]
            const newEle = {
                x: oldEle.x + 1,
                y: oldEle.y,
                z: oldEle.z,
            }

            paramData.push(newEle)

            this.nodeManager.nodeParameterChanged({
                id: this.id,
                data: paramData,
            });
        });

        console.log(this.lightNode)

        this.lightNode.onSelected = (obj) => {
            this.renderControls = true;
            this.updateControlRendering();
        }

        this.lightNode.onDeselected = (obj) => {
            this.renderControls = false;
            this.updateControlRendering();
        }
    }

    updateControlRendering() {
        this.allPositionControls.forEach((v) => {
            v.visible = this.renderControls;
            v.enabled = this.renderControls;
        });
    }

    buildParameterData() {
        const data = [];

        this.allPositionControlsMeshes.forEach((ele) => {
            data.push({
                x: ele.position.x,
                y: ele.position.y,
                z: ele.position.z,
            })
        })

        return data
    }

    newPositionControl(pos) {
        const control = new TransformControls(this.app.Camera, this.app.Renderer.domElement);
        control.setMode('translate');
        control.setSpace("local");


        const mesh = new THREE.Group();

        control.addEventListener('dragging-changed', (event) => {
            this.app.OrbitControls.enabled = !event.value;

            if (this.app.OrbitControls.enabled) {
                this.nodeManager.nodeParameterChanged({
                    id: this.id,
                    data: this.buildParameterData()
                });
            }
        });

        this.allPositionControls.push(control);
        this.allPositionControlsMeshes.push(mesh);

        this.scene.add(mesh);
        this.app.Scene.add(control);
        mesh.position.set(pos.x, pos.y, pos.z);
        control.attach(mesh);

        control.visible = this.renderControls;
        control.enabled = this.renderControls;
    }

    clearPositionControls() {
        this.allPositionControls.forEach((v) => {
            this.app.Scene.remove(v);
        });
        this.allPositionControlsMeshes.forEach((v) => {
            this.scene.remove(v);
        })
        this.allPositionControls = [];
        this.allPositionControlsMeshes = [];
    }

    update(parameterData) {
        this.clearPositionControls();
        parameterData.currentValue.forEach((ele) => {
            this.newPositionControl(ele);
        })
    }
}