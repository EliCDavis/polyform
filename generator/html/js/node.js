import * as THREE from 'three';
import { TransformControls } from 'three/addons/controls/TransformControls.js';


class NodeBasicParameter {
    constructor(nodeManager, id, parameterData, guiFolder, guiFolderData) {
        this.id = id;
        this.guiFolder = guiFolder;
        this.guiFolderData = guiFolderData;

        guiFolderData[id] = parameterData.currentValue;

        let setting = null;

        if (parameterData.type === "coloring.WebColor") {
            setting = guiFolder.addColor(guiFolderData, id)
        } else {
            setting = guiFolder.add(guiFolderData, id)
        }

        setting = setting.name(parameterData.name);

        if (parameterData.type === "int") {
            setting = setting.step(1)
        }

        this.setting = setting
            .listen()
            .onChange((newData) => {
                nodeManager.nodeParameterChanged({
                    id: id,
                    data: newData
                });
            });
    }

    update(parameterData) {
        this.guiFolderData[this.id] = parameterData.currentValue;
    }
}

class NodeVector3Parameter {
    constructor(nodeManager, id, parameterData, app) {
        const control = new TransformControls(app.Camera, app.Renderer.domElement);
        control.setMode('translate');

        this.mesh = new THREE.Group();
        this.mesh.position.x = parameterData.currentValue.x;
        this.mesh.position.y = parameterData.currentValue.y;
        this.mesh.position.z = parameterData.currentValue.z;

        // control.addEventListener('change', () => {
        // });

        control.addEventListener('dragging-changed', (event) => {
            app.OrbitControls.enabled = !event.value;

            if (app.OrbitControls.enabled) {

                const newData = {
                    x: this.mesh.position.x,
                    y: this.mesh.position.y,
                    z: this.mesh.position.z,
                }
                nodeManager.nodeParameterChanged({
                    id: id,
                    data: newData
                });
            }
        });

        app.Scene.add(this.mesh);
        control.attach(this.mesh);
        app.Scene.add(control)
    }

    update(parameterData) {
        this.mesh.position.x = parameterData.currentValue.x;
        this.mesh.position.y = parameterData.currentValue.y;
        this.mesh.position.z = parameterData.currentValue.z;
    }
}

class NodeVector3ArryParameter {
    constructor(nodeManager, id, parameterData, app, guiFolderData) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.guiFolder = app.MeshGenFolder;
        this.guiFolderData = guiFolderData;
        this.app = app;
        this.scene = app.Scene;
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

        const mesh = new THREE.Group();
        mesh.position.x = pos.x;
        mesh.position.y = pos.y;
        mesh.position.z = pos.z;

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

        control.attach(mesh);
        this.scene.add(mesh);
        this.scene.add(control)
    }

    clearPositionControls() {
        this.allPositionControls.forEach((v) => {
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

function BuildParameter(nodeManager, id, parameterData, app, guiFolderData) {
    switch (parameterData.type) {
        case "float64":
        case "float32":
        case "int":
        case "bool":
        case "string":
        case "coloring.WebColor":
            return new NodeBasicParameter(nodeManager, id, parameterData, app.MeshGenFolder, guiFolderData);

        case "vector3.Vector[float64]":
        case "vector3.Vector[float32]":
            return new NodeVector3Parameter(nodeManager, id, parameterData, app);

        case "[]vector3.Vector[float64]":
        case "[]vector3.Vector[float32]":
            return new NodeVector3ArryParameter(nodeManager, id, parameterData, app, guiFolderData);

        default:
            throw new Error("unimplemented type: " + parameterData.type)
    }
}

export class PolyNode {
    constructor(nodeManager, id, nodeData, app, guiFolderData) {
        this.app = app;
        this.guiFolderData = guiFolderData;
        this.nodeManager = nodeManager;

        this.id = id;
        this.name = "";
        this.outputs = [];
        this.version = 0;
        this.dependencies = [];
        this.parameter = null;

        this.update(nodeData);
    }

    update(nodeData) {
        this.name = nodeData.name;
        this.outputs = nodeData.outputs;
        this.version = nodeData.version;
        this.dependencies = nodeData.dependencies;

        if (nodeData.parameter) {
            if (!this.parameter) {
                this.parameter = BuildParameter(this.nodeManager, this.id, nodeData.parameter, this.app, this.guiFolderData);
            } else {
                this.parameter.update(nodeData.parameter)
            }
        }
    }
}