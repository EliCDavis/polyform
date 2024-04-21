import * as THREE from 'three';
import { NodeBasicParameter } from './basic_parameter.js';
import { NodeVector3Parameter } from './vector3_parameter.js';
import { NodeVector3ArryParameter } from './vector3_array_parameter.js';
import { ImageParameterNode } from './image_parameter.js';
import { NodeAABBParameter } from './aabb_parameter.js';
import { ColorParameter } from './color_parameter.js';
import { NodeManager } from '../node_manager.js';


function BuildParameter(nodeManager, id, parameterData, app, guiFolderData) {
    switch (parameterData.type) {
        case "float64":
        case "float32":
        case "int":
        case "bool":
        case "string":
            return new NodeBasicParameter(app, nodeManager, id, parameterData);

        case "coloring.WebColor":
            return new ColorParameter(nodeManager, id, parameterData, app);

        case "vector3.Vector[float64]":
        case "vector3.Vector[float32]":
            return new NodeVector3Parameter(nodeManager, id, parameterData, app);

        case "[]vector3.Vector[float64]":
        case "[]vector3.Vector[float32]":
            return new NodeVector3ArryParameter(nodeManager, id, parameterData, app);

        case "image.Image":
            return new ImageParameterNode(nodeManager, id, parameterData, app);

        case "geometry.AABB":
            return new NodeAABBParameter(nodeManager, id, parameterData, app);

        default:
            throw new Error("build parameter: unimplemented parameter type: " + parameterData.type)
    }
}

// https://stackoverflow.com/a/35953318/4974261
export function camelCaseToWords(str) {
    var result = str
        .replace(/(_)+/g, ' ')
        .replace(/([a-z])([A-Z][a-z])/g, "$1 $2")
        .replace(/([A-Z][a-z])([A-Z])/g, "$1 $2")
        .replace(/([a-z])([A-Z]+[a-z])/g, "$1 $2")
        .replace(/([A-Z]+)([A-Z][a-z][a-z])/g, "$1 $2")
        .replace(/([a-z]+)([A-Z0-9]+)/g, "$1 $2")
        .replace(/([A-Z]+)([A-Z][a-rt-z][a-z]*)/g, "$1 $2")
        .replace(/([0-9])([A-Z][a-z]+)/g, "$1 $2")
        .replace(/([A-Z]{2,})([0-9]{2,})/g, "$1 $2")
        .replace(/([0-9]{2,})([A-Z]{2,})/g, "$1 $2")
        .trim();

    let title = result.charAt(0).toUpperCase() + result.slice(1);
    if (title.endsWith(" Node")) {
        title = title.substring(0, title.length - 5);
    }
    if (title.endsWith(" Node Data")) {
        title = title.substring(0, title.length - 10);
    }
    return title;
}

// function BuildCustomNodeType(app, nodeData, isProducer) {
//     function CustomNode() {
//         for (var inputName in nodeData.inputs) {
//             this.addInput(inputName, nodeData.inputs[inputName].type);
//         }

//         if (!isProducer) {
//             nodeData.outputs.forEach((o) => {
//                 this.addOutput(o.name, o.type);
//             })
//         } else {
//             this.color = "#232";
//             this.bgcolor = "#353";
//             this.addWidget("button", "Download", null, () => {
//                 console.log("presed");
//                 saveFileToDisk("/producer/" + nodeData.name, nodeData.name);
//             })
//         }
//         this.title = camelCaseToWords(nodeData.name);

//         // this.properties = { precision: 1 };
//     }

//     const nodeName = "polyform/" + nodeData.name;
//     LiteGraph.registerNodeType(nodeName, CustomNode);

//     const node = LiteGraph.createNode(nodeName);
//     node.setSize(node.computeSize());

//     // node.pos = [200, app.LightGraph._nodes.length * 100];
//     app.LightGraph.add(node);
//     return node;
// }

export class PolyNode {

    /**
     * 
     * @param {NodeManager} nodeManager 
     * @param {*} id 
     * @param {*} nodeData 
     * @param {*} app 
     * @param {*} guiFolderData 
     * @param {*} isProducer 
     */
    constructor(nodeManager, id, nodeData, app, guiFolderData, isProducer) {
        this.app = app;
        this.guiFolderData = guiFolderData;
        this.nodeManager = nodeManager;
        this.isProducer = isProducer;

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

        let created = false;

        if (nodeData.parameter) {
            if (!this.parameter) {
                this.parameter = BuildParameter(this.nodeManager, this.id, nodeData.parameter, this.app, this.guiFolderData);
                this.lightNode = this.parameter.lightNode;
                created = true;
            } else {
                this.parameter.update(nodeData.parameter)
            }
        } else if (!this.lightNode) {
            // this.lightNode = BuildCustomNodeType(this.app, nodeData, this.isProducer)
            const nodeIdentifier = this.nodeManager.nodeTypeToLitePath.get(nodeData.type)
            const node = LiteGraph.createNode(nodeIdentifier);
            node.setSize(node.computeSize());

            if (this.isProducer) {
                node.color = "#232";
                node.bgcolor = "#353";
                node.addWidget("button", "Download", null, () => {
                    console.log("presed");
                    saveFileToDisk("/producer/" + this.name, this.name);
                })
            }

            // node.pos = [200, app.LightGraph._nodes.length * 100];
            this.app.LightGraph.add(node);

            this.lightNode = node;
            created = true;
        }

        if (created) {
            this.lightNode.nodeInstanceID = this.id;
            // this.lightNode.onConnectInput = (a, b, c, d, e, f, g) => {
            //     console.log("onConnectInput", a, b, c, d, e, f, g)
            // }

            this.lightNode.onConnectionsChange = (inOrOut, slot /* string or number */, connected, linkInfo, inputInfo) => {
                if (this.app.ServerUpdatingNodeConnections) {
                    return;
                }

                const input = inOrOut === LiteGraph.INPUT;
                const output = inOrOut === LiteGraph.OUTPUT;

                console.log("onConnectionsChange", {
                    "input": input,
                    "slot": slot,
                    "connected": connected,
                    "linkInfo": linkInfo,
                    "inputInfo": inputInfo
                })

                if (input && !connected) {
                    this.app.RequestManager.deleteNodeInput(this.id, inputInfo.name)
                }

                if(input && connected) {
                    // console.log(LiteGraph)
                    // console.log(lgraphInstance)

                    const link = lgraphInstance.links[linkInfo.id];
                    const outNode = lgraphInstance.getNodeById(link.origin_id);
                    const inNode = lgraphInstance.getNodeById(link.target_id);
                    // console.log(link)
                    // console.log("out?", outNode)
                    // console.log("in?", inNode)

                    this.app.RequestManager.setNodeInputConnection(
                        inNode.nodeInstanceID,
                        inNode.inputs[link.target_slot].name,
                        outNode.nodeInstanceID,
                        outNode.outputs[link.origin_slot].name,
                    )
                }
            }

            // this.lightNode.onNodeCreated = () => {
            //     if (this.app.ServerUpdatingNodeConnections) {
            //         return;
            //     }
            //     console.log("node created: ", typeData.type)
            //     this.app.RequestManager.createNode(typeData.type, (data) => {
            //         this.lightNode.nodeInstanceID = data.nodeID;
            //     })
            // }
        }
    }

    updateConnections() {

    }
}