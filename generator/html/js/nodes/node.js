import * as THREE from 'three';
import { TransformControls } from 'three/addons/controls/TransformControls.js';
import { NodeBasicParameter } from './basic_parameter.js';
import { NodeVector3Parameter } from './vector3_parameter.js';



class NodeVector3ArryParameter {
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

function BuildParameter(nodeManager, id, parameterData, app, guiFolderData) {
    switch (parameterData.type) {
        case "float64":
        case "float32":
        case "int":
        case "bool":
        case "string":
        case "coloring.WebColor":
            return new NodeBasicParameter(app, nodeManager, id, parameterData, app.MeshGenFolder, guiFolderData);

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

function BuildCustomNode(app, nodeData) {
    function CustomNode() {
        nodeData.inputs.forEach((i) => {
            this.addInput(i.name, i.type);
        })

        nodeData.outputs.forEach((o) => {
            this.addOutput(o.name, o.type);
        })
        // this.properties = { precision: 1 };
    }
    CustomNode.title = nodeData.name;

    const nodeName = "polyform/" + nodeData.name;
    LiteGraph.registerNodeType(nodeName, CustomNode);

    const node = LiteGraph.createNode(nodeName);
    console.log(node)
    // node.pos = [200, app.LightGraph._nodes.length * 100];
    app.LightGraph.add(node);
    return node;
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
        this.lightNode = null;

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
                this.lightNode = this.parameter.lightNode;
            } else {
                this.parameter.update(nodeData.parameter)
            }
        } else if(!this.lightNode) {
            this.lightNode = BuildCustomNode(this.app, nodeData)
        }
    }

    updateConnections() {

    }
}