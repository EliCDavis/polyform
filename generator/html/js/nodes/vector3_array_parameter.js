import { TransformControls } from 'three/addons/controls/TransformControls.js';

export class NodeVector3ArryParameter {
    constructor(nodeManager, id, parameterData, app, guiFolderData) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.guiFolder = app.MeshGenFolder;
        this.guiFolderData = guiFolderData;
        this.app = app;
        this.scene = app.ViewerScene;
        this.allPositionControls = [];
        this.allPositionControlsMeshes = [];

        parameterData.currentValue.forEach((ele) => {
            this.newPositionControl(ele);
        })

        this.guiFolderData[this.id] = () => {

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
        }

        this.setting = this.guiFolder.
            add(this.guiFolderData, this.id).
            name("Add to " + parameterData.name).
            listen()
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
        control.space = "local";

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