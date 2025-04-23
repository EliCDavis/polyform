import { BasicParameterNodeController } from './basic_parameter.js';
import { Vector3ParameterNodeController } from './vector3_parameter.js';
import { Vector3ArrayParameterNodeController } from './vector3_array_parameter.js';
import { ImageParameterNodeController } from './image_parameter.js';
import { AABBParameterNodeController } from './aabb_parameter.js';
import { NodeManager } from '../node_manager.js';
import { FileParameterNodeController } from './file_parameter.js';
import { getFileExtension, getLastSegmentOfURL } from '../utils.js';
import { Vector2ParameterNodeController } from './vector2_parameter.js';
import { NodeInstance, NodeInstanceAssignedInput, NodeInstanceOutput, NodeType } from '../schema.js';
import { RequestManager, saveFileToDisk } from '../requests.js';
import { FlowNode, GlobalWidgetFactory, ImageWidget } from '@elicdavis/node-flow';
import { ThreeApp } from '../three_app.js';
import { ProducerViewManager } from '../ProducerView/producer_view_manager.js';

export const InstanceIDProperty: string = "instanceID"

interface ParameterController {
    update(parameterData: any): void
}

function BuildParameter(
    id: string,
    flowNode: FlowNode,
    nodeManager: NodeManager,
    requestManager: RequestManager,
    parameterData,
    app: ThreeApp
): ParameterController {
    switch (parameterData.type) {
        case "float64":
        case "float32":
        case "int":
        case "bool":
        case "string":
        case "coloring.WebColor":
            return new BasicParameterNodeController(flowNode, nodeManager, id, parameterData);

        case "vector2.Vector[float64]":
        case "vector2.Vector[float32]":
            return new Vector2ParameterNodeController(flowNode, nodeManager, id, parameterData);

        case "vector3.Vector[float64]":
        case "vector3.Vector[float32]":
            return new Vector3ParameterNodeController(flowNode, nodeManager, id, parameterData, app);

        case "[]vector3.Vector[float64]":
        case "[]vector3.Vector[float32]":
            return new Vector3ArrayParameterNodeController(flowNode, nodeManager, id, parameterData, app);

        case "image.Image":
            return new ImageParameterNodeController(flowNode, nodeManager, requestManager, id, parameterData);

        case "[]uint8":
            return new FileParameterNodeController(flowNode, nodeManager, id, parameterData);

        case "geometry.AABB":
            return new AABBParameterNodeController(flowNode, nodeManager, id, parameterData, app);

        default:
            throw new Error("build parameter: unimplemented parameter type: " + parameterData.type)
    }
}

// https://stackoverflow.com/a/35953318/4974261
export function camelCaseToWords(str: string): string {
    let title = str;

    const endMatches = str.match(/\[[^\]]*\]$/g);
    let end = "";
    if (endMatches !== null && endMatches.length > 0) {
        end = " " + endMatches[0];
        title = title.substring(0, title.length - endMatches[0].length);
    }

    title = title
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

    title = title.charAt(0).toUpperCase() + title.slice(1);
    if (title.endsWith(" Node")) {
        title = title.substring(0, title.length - 5);
    }
    if (title.endsWith(" Node Data")) {
        title = title.substring(0, title.length - 10);
    }
    if (title.endsWith("NodeData")) {
        title = title.substring(0, title.length - 8);
    }
    return title + end;
}

export class PolyNodeController {

    flowNode: FlowNode;

    nodeManager: NodeManager;

    id: string;

    name: string;

    isProducer: boolean;

    outputs: NodeInstanceOutput;

    dependencies: NodeInstanceAssignedInput;

    parameter: ParameterController;

    requestManager: RequestManager;

    app: ThreeApp;

    producerViewManager: ProducerViewManager;

    constructor(
        flowNode: FlowNode,
        nodeManager: NodeManager,
        id: string,
        nodeData: NodeInstance,
        nodeType: NodeType,
        app: ThreeApp,
        producerOutput: string,
        requestManager: RequestManager,
        producerViewManager: ProducerViewManager
    ) {
        // console.log(liteNode)
        this.flowNode = flowNode;
        this.id = id;
        this.app = app;
        this.nodeManager = nodeManager;
        this.isProducer = !!producerOutput;
        this.requestManager = requestManager;
        this.producerViewManager = producerViewManager;

        this.name = "";
        this.outputs = {};
        this.dependencies = {};

        this.parameter = null;

        if (nodeData.metadata) {
            if (nodeData.metadata.position) {
                // console.log("setting position....", nodeData.metadata.position)
                this.flowNode.setPosition(nodeData.metadata.position);
            }
        }

        this.flowNode.addDragStoppedListener((nodeChanged) => {

            // Round to decrease file size of json. Precision isn't needed
            const pos = this.flowNode.getPosition();
            pos.x = Math.round(pos.x);
            pos.y = Math.round(pos.y);

            this.requestManager.setNodeMetadata(
                this.flowNode.getProperty(InstanceIDProperty),
                "position",
                pos,
                (response) => {
                    console.log("set metadata response", response)
                }
            );
        });

        if (nodeData.parameter) {
            this.parameter = BuildParameter(
                this.id,
                flowNode,
                this.nodeManager,
                this.requestManager,
                nodeData.parameter,
                this.app
            );
            if (nodeData.parameter.description) {
                flowNode.setInfo(nodeData.parameter.description);
            }
        }

        // type TitleChangeCallback = (node: FlowNode, oldTitle: string, newTitle: string) => void
        this.flowNode.addTitleChangeListener((_, __, newTitle) => {

            // The only two times we can change title is if the node is a 
            // parameter, or if it's a producer.

            if (this.isProducer) {
                this.requestManager.setProducerTitle(
                    this.flowNode.getProperty(InstanceIDProperty),
                    {
                        nodePort: producerOutput,
                        producer: newTitle
                    },
                    () => { }
                );
            } else {
                this.requestManager.setParameterTitle(
                    this.flowNode.getProperty(InstanceIDProperty),
                    newTitle,
                    () => { }
                );
            }
        });


        // Only parameters can change their info
        if (!this.isProducer) {
            this.flowNode.addInfoChangeListener((_, __, newTitle) => {
                this.requestManager.setParameterInfo(
                    this.flowNode.getProperty(InstanceIDProperty),
                    newTitle,
                    () => { }
                );
            });
        }

        if (this.isProducer) {
            const ext = getFileExtension(nodeData.name);
            if (ext === "png") {
                const imageWidget = GlobalWidgetFactory.create(flowNode, "image", {}) as ImageWidget;
                flowNode.addWidget(imageWidget);
                producerViewManager.Subscribe((url, image) => {
                    // console.log(url, image)
                    // imageWidget.SetBlob(image);
                    // console.log(nodeData.name, getLastSegmentOfURL(url), image);
                    if (getLastSegmentOfURL(url) === nodeData.name) {
                        imageWidget.SetUrl(url)
                    }
                });
            }

            // this.flowNode.color = "#232";
            // this.flowNode.bgcolor = "#353";
            const downloadButton = GlobalWidgetFactory.create(flowNode, "button", {
                text: "Download",
                callback: () => {
                    saveFileToDisk("/producer/value/" + this.name, this.name);
                }
            })
            this.flowNode.addWidget(downloadButton);
        }

        // type ConnectionChangeCallback = (connection: Connection, connectionIndex: number, port: Port, portType: PortType, node: FlowNode) => void
        for (let i = 0; i < this.flowNode.inputs(); i++) {
            const port = this.flowNode.inputPort(i);
            port.addConnectionAddedListener((connection, connectionIndex, port, portType, node) => {
                if (this.nodeManager.serverUpdatingNodeConnections) {
                    return;
                }
                console.log("connection ADDED", connection, connectionIndex, port, portType, node);

                let inputPort = connection.inPort().getDisplayName();
                if (portType === "INPUTARRAY") {
                    inputPort += "." + connectionIndex;
                }

                this.requestManager.setNodeInputConnection(
                    this.flowNode.getProperty(InstanceIDProperty),
                    inputPort,
                    connection.outNode().getProperty(InstanceIDProperty),
                    connection.outPort().getDisplayName(),
                )
            });

            port.addConnectionRemovedListener((connection, connectionIndex, port, portType, node) => {
                if (this.nodeManager.serverUpdatingNodeConnections) {
                    return;
                }
                console.log("connection removed", {
                    "connection": connection,
                    "connectionIndex": connectionIndex,
                    "port": port,
                    "portType": portType,
                    "node": node
                })

                let inputPort = port.getDisplayName();
                if (portType === "INPUTARRAY") {
                    inputPort += "." + connectionIndex;
                }

                this.requestManager.deleteNodeInput(this.id, inputPort)
            });
        }

        this.update(nodeData);
    }


    update(nodeData: NodeInstance) {
        this.name = nodeData.name;
        this.outputs = nodeData.output;
        this.dependencies = nodeData.assignedInput;

        // console.log(nodeData);
        if (nodeData.metadata) {
            if (nodeData.metadata.position) {
                this.flowNode.setPosition(nodeData.metadata.position);
            }
        }

        if (nodeData.parameter) {
            this.parameter.update(nodeData.parameter)
        }
    }

    updateConnections() {

    }
}