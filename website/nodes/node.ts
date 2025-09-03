import { BasicParameterNodeController } from './basic_parameter.js';
import { Vector3ParameterNodeController } from './vector3_parameter.js';
import { Vector3ArrayParameterNodeController } from './vector3_array_parameter.js';
import { ImageParameterNodeController } from './image_parameter.js';
import { AABBParameterNodeController } from './aabb_parameter.js';
import { NodeManager } from '../node_manager.js';
import { FileParameterNodeController } from './file_parameter.js';
import { getFileExtension, getLastSegmentOfURL } from '../utils.js';
import { Vector2ParameterNodeController } from './vector2_parameter.js';
import { NodeInstance, NodeInstanceAssignedInput, NodeInstanceOutput, NodeDefinition, ExecutionReport } from '../schema.js';
import { RequestManager, saveFileToDisk } from '../requests.js';
import { FlowNode, GlobalWidgetFactory, ImageWidget, MessageType, StringWidget } from '@elicdavis/node-flow';
import { ThreeApp } from '../three_app.js';
import { ProducerViewManager } from '../ProducerView/producer_view_manager.js';

export const InstanceIDProperty: string = "instanceID"

interface ParameterController {
    update(parameterData: any): void
}

function formatNanoseconds(ns: number): string {
    const ms = ns / 1_000_000;
    if (ms < 1000) {
        return `${ms.toFixed(ms < 10 ? 2 : 1)}ms`;
    }

    const s = ms / 1000;
    return `${s.toFixed(s < 10 ? 2 : 1)}s`;
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
        case "coloring.Color":
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

    nodeDefinition: NodeDefinition;

    serializableOutputTypes: Array<string>;

    nodeOutputWidgets: Map<string, ImageWidget>;

    constructor(
        flowNode: FlowNode,
        nodeManager: NodeManager,
        id: string,
        nodeData: NodeInstance,
        app: ThreeApp,
        producerOutput: string,
        requestManager: RequestManager,
        producerViewManager: ProducerViewManager,
        nodeDefinition: NodeDefinition,
        serializableOutputTypes: Array<string>
    ) {
        // console.log(liteNode)
        this.flowNode = flowNode;
        this.id = id;
        this.app = app;
        this.nodeManager = nodeManager;
        this.isProducer = !!producerOutput;
        this.requestManager = requestManager;
        this.producerViewManager = producerViewManager;
        this.nodeDefinition = nodeDefinition;
        this.serializableOutputTypes = serializableOutputTypes;
        this.nodeOutputWidgets = new Map<string, ImageWidget>();

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

        if (!!producerOutput) {
            const ext = getFileExtension(nodeData.name);
            if (ext === "png") {
                const imageWidget = GlobalWidgetFactory.create(flowNode, "image", {}) as ImageWidget;
                flowNode.addWidget(imageWidget);
                producerViewManager.SubscribeToProducerRefresh((url, image) => {
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
                    saveFileToDisk("./zip/" + this.id + "/" + producerOutput, this.id);
                }
            });
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

    fetchOutput(outputKey: string, version: number): void {

        // if (version === -1) {
        //     return;
        // }

        // Don't do anything if we're up to date
        if (outputKey in this.outputs && this.outputs[outputKey].version === version) {
            return;
        }

        let recognizeOutput = this.nodeDefinition?.outputs && outputKey in this.nodeDefinition.outputs;
        if (!recognizeOutput) {
            return
        }

        for (let i = 0; i < this.flowNode.outputs(); i++) {
            const outputPort = this.flowNode.outputPort(i);
            if (outputPort.getDisplayName() !== outputKey) {
                continue;
            }
            if (outputPort.connections().length === 0) {

                // If we were once connected and are no longer, clear image widget
                if (this.nodeOutputWidgets.has(outputKey)) {
                    this.flowNode.removeWidget(this.nodeOutputWidgets.get(outputKey));
                }
                return;
            }
            console.log("Woo", outputPort.connections().length)
        }

        let found = false;
        for (let i = 0; i < this.serializableOutputTypes.length; i++) {
            if (this.nodeDefinition.outputs[outputKey].type == this.serializableOutputTypes[i]) {
                found = true;
                break;
            }
        }

        if (!found) {
            return;
        }

        // const stringWidget = GlobalWidgetFactory.create(this.flowNode, "string", {}) as StringWidget;
        // this.flowNode.addWidget(stringWidget);
        // stringWidget.Set("" + version);

        if (!this.nodeOutputWidgets.has(outputKey)) {
            const imageWidget = GlobalWidgetFactory.create(this.flowNode, "image", {}) as ImageWidget;
            this.flowNode.addWidget(imageWidget);
            this.nodeOutputWidgets.set(outputKey, imageWidget)
        }

        this.nodeOutputWidgets.get(outputKey).SetUrl(`./node/output/${this.id}/${outputKey}`);
    }

    update(nodeData: NodeInstance): void {
        this.name = nodeData.name;
        this.dependencies = nodeData.assignedInput;

        if (nodeData.metadata) {
            if (nodeData.metadata.position) {
                this.flowNode.setPosition(nodeData.metadata.position);
            }
        }

        if (nodeData.parameter) {
            this.parameter.update(nodeData.parameter)
        }
    }

    setOutputPortReport(portName: string, report: ExecutionReport) {
        if (report.selfTime !== undefined) {
            this.flowNode.addMessage({
                message: `${portName}: ${formatNanoseconds(report.selfTime)}`,
                alwaysShow: true
            })
        } else if (report.totalTime !== 0) {
            this.flowNode.addMessage({
                message: `${portName}: total ${formatNanoseconds(report.totalTime)}`,
                alwaysShow: true
            })
        }

        if (report.errors) {
            for (let errI = 0; errI < report.errors.length; errI++) {
                this.flowNode.addMessage({
                    message: `${portName}: ${report.errors[errI]}`,
                    type: MessageType.Error,
                    alwaysShow: true
                })
            }
        }
    }

    updateConnections(nodeData: NodeInstance) {

        // Fetch updates for changed output
        for (let outputKey in nodeData.output) {
            this.fetchOutput(outputKey, nodeData.output[outputKey].version);
        }
        this.outputs = nodeData.output;

    }
}