import * as THREE from 'three';
import { NodeBasicParameter } from './basic_parameter.js';
import { NodeVector3Parameter } from './vector3_parameter.js';
import { NodeVector3ArryParameter } from './vector3_array_parameter.js';


function BuildImageParameterNode(app) {
    const node = LiteGraph.createNode("polyform/Image");
    console.log(node)
    app.LightGraph.add(node);
    return node;
}

class ImageParameterNode {

    constructor(nodeManager, id, parameterData, app) {
        this.lightNode = BuildImageParameterNode(app);
        this.lightNode.title = parameterData.name;
        app.LightGraph.add(this.lightNode);

        this.lightNode.onDropFile = (file) => {
            // console.log(file)
            var reader = new FileReader();
            reader.onload = (evt) => {
                console.log(evt.target.result)
                nodeManager.nodeParameterChanged({
                    id: id,
                    data: evt.target.result,
                    binary: true
                });
            }
            reader.readAsArrayBuffer(file);
        }
    }

    update(parameterData) {
        const curVal = parameterData.currentValue;
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

        case "image.Image":
            return new ImageParameterNode(nodeManager, id, parameterData, app);

        default:
            throw new Error("build parameter: unimplemented parameter type: " + parameterData.type)
    }
}

// https://stackoverflow.com/a/35953318/4974261
function camelCaseToWords(str) {
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
    return title;
}

/**
 * 
 * @param {*} app 
 * @param {*} nodeData 
 * @param {bool} isProducer 
 * @returns 
 */
function BuildCustomNode(app, nodeData, isProducer) {
    function CustomNode() {
        for (var inputName in nodeData.inputs) {
            this.addInput(inputName, nodeData.inputs[inputName].type);
        }

        if (!isProducer) {
            nodeData.outputs.forEach((o) => {
                this.addOutput(o.name, o.type);
            })
        } else {
            this.color = "#232";
            this.bgcolor = "#353";
            this.addWidget("button", "Download", null, () => {
                console.log("presed");
                saveFileToDisk("/producer/" + nodeData.name, nodeData.name);
            })
        }
        this.title = camelCaseToWords(nodeData.name);

        // this.properties = { precision: 1 };
    }

    const nodeName = "polyform/" + nodeData.name;
    LiteGraph.registerNodeType(nodeName, CustomNode);

    const node = LiteGraph.createNode(nodeName);
    console.log(node)
    // node.pos = [200, app.LightGraph._nodes.length * 100];
    app.LightGraph.add(node);
    return node;
}

export class PolyNode {
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

        if (nodeData.parameter) {
            if (!this.parameter) {
                this.parameter = BuildParameter(this.nodeManager, this.id, nodeData.parameter, this.app, this.guiFolderData);
                this.lightNode = this.parameter.lightNode;
            } else {
                this.parameter.update(nodeData.parameter)
            }
        } else if (!this.lightNode) {
            this.lightNode = BuildCustomNode(this.app, nodeData, this.isProducer)
        }
    }

    updateConnections() {

    }
}