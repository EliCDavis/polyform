import * as THREE from 'three';
import { BasicParameterNodeController } from './basic_parameter.js';
import { Vector3ParameterNodeController } from './vector3_parameter.js';
import { Vector3ArrayParameterNodeController } from './vector3_array_parameter.js';
import { ImageParameterNodeController } from './image_parameter.js';
import { AABBParameterNodeController } from './aabb_parameter.js';
import { NodeManager } from '../node_manager.js';
import { FileParameterNodeController } from './file_parameter.js';
import { getFileExtension, getLastSegmentOfURL } from '../utils.js';
import { Vector2ParameterNodeController } from './vector2_parameter.js';


function BuildParameter(liteNode, nodeManager, id, parameterData, app) {
    switch (parameterData.type) {
        case "float64":
        case "float32":
        case "int":
        case "bool":
        case "string":
        case "coloring.WebColor":
            return new BasicParameterNodeController(liteNode, nodeManager, id, parameterData);

        case "vector2.Vector[float64]":
        case "vector2.Vector[float32]":
            return new Vector2ParameterNodeController(liteNode, nodeManager, id, parameterData, app);

        case "vector3.Vector[float64]":
        case "vector3.Vector[float32]":
            return new Vector3ParameterNodeController(liteNode, nodeManager, id, parameterData, app);

        case "[]vector3.Vector[float64]":
        case "[]vector3.Vector[float32]":
            return new Vector3ArrayParameterNodeController(liteNode, nodeManager, id, parameterData, app);

        case "image.Image":
            return new ImageParameterNodeController(liteNode, nodeManager, id, parameterData, app);

        case "[]uint8":
            return new FileParameterNodeController(liteNode, nodeManager, id, parameterData, app);

        case "geometry.AABB":
            return new AABBParameterNodeController(liteNode, nodeManager, id, parameterData, app);

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
    if (title.endsWith("NodeData")) {
        title = title.substring(0, title.length - 8);
    }
    return title;
}

export class PolyNodeController {

    /**
     * 
     * @param {*} liteNode 
     * @param {NodeManager} nodeManager 
     * @param {string} id 
     * @param {*} nodeData 
     * @param {*} app 
     * @param {boolean} isProducer 
     */
    constructor(liteNode, nodeManager, id, nodeData, app, isProducer) {
        // console.log(liteNode)
        this.liteNode = liteNode;
        this.id = id;
        this.app = app;
        this.nodeManager = nodeManager;
        this.isProducer = isProducer;

        this.name = "";
        this.outputs = [];
        this.version = 0;
        this.dependencies = [];

        this.parameter = null;

        if (nodeData.parameter) {
            this.parameter = BuildParameter(liteNode, this.nodeManager, this.id, nodeData.parameter, this.app);
        }

        if (this.isProducer) {
            const ext = getFileExtension(nodeData.name);
            if (ext === "png") {
                const imageWidget = GlobalWidgetFactory.create(liteNode, "image", {});
                liteNode.addWidget(imageWidget);
                app.SchemaRefreshManager.Subscribe((url, image) => {
                    // console.log(url, image)
                    // imageWidget.SetBlob(image);
                    // console.log(nodeData.name, getLastSegmentOfURL(url), image);
                    if (getLastSegmentOfURL(url) === nodeData.name) {
                        imageWidget.SetUrl(url)
                    }
                });
            }

            this.liteNode.color = "#232";
            this.liteNode.bgcolor = "#353";
            const downloadButton = GlobalWidgetFactory.create(liteNode, "button", {
                text: "Download",
                callback: () => {
                    saveFileToDisk("/producer/" + this.name, this.name);
                }
            })
            this.liteNode.addWidget(downloadButton);
        }

        // TODO: FOIXM AEJPRGEAIPGNREAGI
        this.liteNode.onConnectionsChange = this.onConnectionChange.bind(this);

        this.update(nodeData);
    }

    /**
     * 
     * @param {number} inOrOut 
     * @param {string|number} slot 
     * @param {boolean} connected 
     * @param {*} linkInfo 
     * @param {*} inputInfo 
     * @returns {void} 
     */
    onConnectionChange(inOrOut, slot /* string or number */, connected, linkInfo, inputInfo) {
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

        if (input && connected) {
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

    update(nodeData) {
        this.name = nodeData.name;
        this.outputs = nodeData.outputs;
        this.version = nodeData.version;
        this.dependencies = nodeData.dependencies;

        if (nodeData.parameter) {
            this.parameter.update(nodeData.parameter)
        }
    }

    updateConnections() {

    }
}