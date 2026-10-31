import { BasicParameterNodeController } from './basic_parameter';
import { Vector3ParameterNodeController } from './vector3_parameter';
import { ImageParameterNodeController } from './image_parameter';
import { AABBParameterNodeController } from './aabb_parameter';
import { NodeManager } from '../node_manager.js';
import { FileParameterNodeController } from './file_parameter';
import { getFileExtension, getLastSegmentOfURL } from '../utils.js';
import { Vector2ParameterNodeController } from './vector2_parameter';
import { Vector4ParameterNodeController } from './vector4_parameter';
import { ArrayParameterNodeController, colorItem, floatItem, intItem, stringItem, trsItem, vec2Item, vec2IntItem, vec3Item, vec3IntItem } from './array_parameter';
import { GradientParameterNodeController } from './gradient_parameter';
import { TRSParameterNodeController } from './trs_parameter';
import { QuaternionParameterNodeController } from './quaternion_parameter';
import { NodeInstance, NodeInstanceAssignedInput, NodeInstanceOutput, NodeDefinition, NodeInput, NodeOutput, ExecutionReport, subGraphBoundaryKind } from '../schema.js';
import { RequestManager, saveFileToDisk } from '../requests.js';
import { FlowNode, GlobalWidgetFactory, ImageWidget, MessageType, Port, ToggleWidget, type Widget } from '@elicdavis/node-flow';
import { ThreeApp } from '../three_app.js';
import { ProducerViewManager } from '../ProducerView/producer_view_manager.js';
import type { NodeParameter, ScalarNodeParameter } from '../../types/parameter.js';
import { SubGraphBoundaryNodeController } from './boundary_node_controller.js';
import { isSubGraphRuntimeType, resolveDynamicType, shortDynamicPattern } from '../portTypes.js';
import { useGraphTabStore } from '../../stores/graphTabStore.js';
import { nodeOutputUrl } from '../../features/preview/PreviewPanel.js';
import { thumbnails } from '../../features/preview/thumbnail_renderer.js';
import { lockIfReadOnly } from './read_only.js';

export const InstanceIDProperty: string = "instanceID"

// Output types the server serves as a GLB, previewed in a tab of their own.
const PREVIEW_3D_TYPES = new Set([
    "github.com/EliCDavis/polyform/modeling.Mesh",
    "[]github.com/EliCDavis/polyform/modeling.Mesh",
    "*github.com/EliCDavis/polyform/formats/gltf.PolyformModel",
    "[]*github.com/EliCDavis/polyform/formats/gltf.PolyformModel",
]);

interface ParameterController {
    update(parameterData: NodeParameter): void
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
    parameterData: NodeParameter,
    app: ThreeApp
): ParameterController {
    if (nodeManager.enumMembers(parameterData.type)) {
        return new BasicParameterNodeController(flowNode, nodeManager, id, parameterData as ScalarNodeParameter);
    }

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

        case "vector4.Vector[float64]":
        case "vector4.Vector[float32]":
            return new Vector4ParameterNodeController(flowNode, nodeManager, id, parameterData);

        case "[]vector3.Vector[float64]":
        case "[]vector3.Vector[float32]":
            return new ArrayParameterNodeController(flowNode, nodeManager, id, parameterData, vec3Item, app);

        case "[]float64":
            return new ArrayParameterNodeController(flowNode, nodeManager, id, parameterData, floatItem);

        case "[]int":
            return new ArrayParameterNodeController(flowNode, nodeManager, id, parameterData, intItem);

        case "[]string":
            return new ArrayParameterNodeController(flowNode, nodeManager, id, parameterData, stringItem);

        case "[]vector2.Vector[float64]":
            return new ArrayParameterNodeController(flowNode, nodeManager, id, parameterData, vec2Item, app);

        case "[]vector2.Vector[int]":
            return new ArrayParameterNodeController(flowNode, nodeManager, id, parameterData, vec2IntItem, app);

        case "[]vector3.Vector[int]":
            return new ArrayParameterNodeController(flowNode, nodeManager, id, parameterData, vec3IntItem, app);

        case "quaternion.Quaternion":
            return new QuaternionParameterNodeController(flowNode, nodeManager, id, parameterData, app);

        case "trs.TRS":
            return new TRSParameterNodeController(flowNode, nodeManager, id, parameterData, app);

        case "[]trs.TRS":
            return new ArrayParameterNodeController(flowNode, nodeManager, id, parameterData, trsItem, app);

        case "[]coloring.Color":
            return new ArrayParameterNodeController(flowNode, nodeManager, id, parameterData, colorItem);

        case "coloring.Gradient[github.com/EliCDavis/polyform/drawing/coloring.Color]":
            return new GradientParameterNodeController(flowNode, nodeManager, id, parameterData);

        case "image.Image":
            return new ImageParameterNodeController(flowNode, nodeManager, requestManager, id, parameterData);

        case "[]uint8":
            return new FileParameterNodeController(flowNode, nodeManager, id, parameterData);

        case "geometry.AABB":
            return new AABBParameterNodeController(flowNode, nodeManager, id, parameterData, app);

        default:
            throw new Error("build parameter: unimplemented parameter type: " + (parameterData as any).type)
    }
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

    boundary: SubGraphBoundaryNodeController;

    forEach: boolean;

    forEachToggle: ToggleWidget | null;

    requestManager: RequestManager;

    app: ThreeApp;

    producerViewManager: ProducerViewManager;

    nodeDefinition: NodeDefinition;

    serializableOutputTypes: Array<string>;

    nodeOutputWidgets: Map<string, ImageWidget>;

    previewButtons: Map<string, Widget>;

    thumbnailVersions: Map<string, number>;

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
        serializableOutputTypes: Array<string>,
    ) {
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
        this.previewButtons = new Map<string, Widget>();
        this.thumbnailVersions = new Map<string, number>();

        this.name = "";
        this.outputs = {};
        this.dependencies = {};

        this.parameter = null;
        this.boundary = null;
        this.forEach = false;
        this.forEachToggle = null;

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
        } else if (subGraphBoundaryKind(nodeData)) {
            this.boundary = new SubGraphBoundaryNodeController(
                flowNode,
                this.nodeManager,
                this.requestManager,
                this.id,
                nodeData,
            );
        } else if (isSubGraphRuntimeType(nodeData.type) && nodeData.subGraphId) {
            const subGraphId = nodeData.subGraphId;
            const subGraphLabel = nodeData.name || nodeData.subGraphId;
            const openSubGraphButton = GlobalWidgetFactory.create(flowNode, "button", {
                text: "Open Sub-Graph",
                callback: () => {
                    useGraphTabStore.getState().openSubGraphTab(subGraphId, subGraphLabel);
                },
            });
            flowNode.addWidget(openSubGraphButton);

            this.forEach = !!nodeData.forEach;
            this.forEachToggle = GlobalWidgetFactory.create(flowNode, "toggle", {
                label: "For Each",
                value: this.forEach,
                callback: (each) => {
                    this.forEach = each;
                    this.requestManager.setSubGraphInstanceForEach(
                        this.id,
                        { each, scope: this.nodeManager.getScopeApiPath() },
                    );
                },
            });
            flowNode.addWidget(this.forEachToggle);
        }

        this.flowNode.addTitleChangeListener((_, __, newTitle) => {

            if (this.boundary) {
                return;
            }

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
        if (!this.isProducer && this.parameter) {
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
                const imageWidget = GlobalWidgetFactory.create(flowNode, "image", {});
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

        this.flowNode.widgets().forEach(lockIfReadOnly);

        this.attachInputConnectionListeners();

        this.update(nodeData);
    }

    attachInputConnectionListeners(): void {
        for (let i = 0; i < this.flowNode.inputs(); i++) {
            const port = this.flowNode.inputPort(i);
            port.addConnectionAddedListener((connection, connectionIndex, port, portType, node) => {
                if (this.nodeManager.serverUpdatingNodeConnections) {
                    return;
                }
                console.log("connection ADDED", connection, connectionIndex, port, portType, node);

                let inputPort = connection.inPort().getDisplayName();
                const array = portType === "INPUTARRAY";
                if (array) {
                    inputPort += "." + connectionIndex;
                }

                this.requestManager.setNodeInputConnection(
                    this.flowNode.getProperty(InstanceIDProperty),
                    inputPort,
                    connection.outNode().getProperty(InstanceIDProperty),
                    connection.outPort().getDisplayName(),
                    undefined,
                    array,
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
    }

    // A thumbnail of the geometry sits above a button that opens it in a tab.
    private ensurePreviewButton(outputKey: string): void {
        if (this.previewButtons.has(outputKey)) {
            return;
        }
        const thumbnail = GlobalWidgetFactory.create(this.flowNode, "image", { maxWidth: 180 });
        this.flowNode.addWidget(thumbnail);
        this.nodeOutputWidgets.set(outputKey, thumbnail);

        const single = Object.keys(this.nodeDefinition.outputs).length === 1;
        const button = GlobalWidgetFactory.create(this.flowNode, "button", {
            text: single ? "Open" : `Open ${outputKey}`,
            callback: () => {
                const title = this.name || this.nodeDefinition.displayName || this.id;
                useGraphTabStore.getState().openPreviewTab(
                    this.id,
                    outputKey,
                    single ? title : `${title} · ${outputKey}`,
                    this.nodeManager.getScopeApiPath(),
                );
            },
        });
        this.flowNode.addWidget(button);
        this.previewButtons.set(outputKey, button);
    }

    private refreshThumbnail(outputKey: string, version: number): void {
        if (this.thumbnailVersions.get(outputKey) === version) {
            return;
        }
        this.thumbnailVersions.set(outputKey, version);
        const scope = this.nodeManager.getScopeApiPath();
        const url = nodeOutputUrl(this.id, outputKey, scope);
        thumbnails
            .request(`${scope ?? "root"}:${this.id}:${outputKey}`, `${url}${url.includes("?") ? "&" : "?"}v=${version}`)
            .then((dataUrl) => {
                const widget = this.nodeOutputWidgets.get(outputKey);
                if (!widget || dataUrl === null) {
                    return;
                }
                widget.SetUrl(dataUrl);
            })
            .catch((err) => console.warn(`thumbnail for ${this.id}.${outputKey}:`, err));
    }

    private outputConnected(outputKey: string): boolean {
        for (let i = 0; i < this.flowNode.outputs(); i++) {
            const outputPort = this.flowNode.outputPort(i);
            if (outputPort.getDisplayName() === outputKey) {
                return outputPort.connections().length > 0;
            }
        }
        return false;
    }

    private removePreview(outputKey: string): void {
        const image = this.nodeOutputWidgets.get(outputKey);
        if (image) {
            this.flowNode.removeWidget(image);
            this.nodeOutputWidgets.delete(outputKey);
        }
        const button = this.previewButtons.get(outputKey);
        if (button) {
            this.flowNode.removeWidget(button);
            this.previewButtons.delete(outputKey);
        }
        this.thumbnailVersions.delete(outputKey);
    }

    fetchOutput(outputKey: string, version: number): void {
        let recognizeOutput = this.nodeDefinition?.outputs && outputKey in this.nodeDefinition.outputs;
        if (!recognizeOutput) {
            return
        }

        if (!this.outputConnected(outputKey)) {
            this.removePreview(outputKey);
            return;
        }

        if (PREVIEW_3D_TYPES.has(this.nodeDefinition.outputs[outputKey].type)) {
            this.ensurePreviewButton(outputKey);
            this.refreshThumbnail(outputKey, version);
            return;
        }

        if (!this.serializableOutputTypes.includes(this.nodeDefinition.outputs[outputKey].type)) {
            return;
        }

        if (this.thumbnailVersions.get(outputKey) === version) {
            return;
        }
        this.thumbnailVersions.set(outputKey, version);

        const url = nodeOutputUrl(this.id, outputKey, this.nodeManager.getScopeApiPath());
        fetch(url)
            .then(async (response) => {
                if (this.thumbnailVersions.get(outputKey) !== version) {
                    return;
                }
                if (response.status === 204) {
                    this.removePreview(outputKey);
                    this.thumbnailVersions.set(outputKey, version);
                    return;
                }
                if (!response.ok) {
                    throw new Error(`${response.status} fetching ${url}`);
                }
                const blob = await response.blob();
                let widget = this.nodeOutputWidgets.get(outputKey);
                if (!widget) {
                    widget = GlobalWidgetFactory.create(this.flowNode, "image", {}) as ImageWidget;
                    this.flowNode.addWidget(widget);
                    this.nodeOutputWidgets.set(outputKey, widget);
                }
                widget.SetBlob(blob);
            })
            .catch((err) => console.warn(`preview for ${this.id}.${outputKey}:`, err));
    }

    // A dynamic port has no type until something is wired into it
    applyDynamicTypes(nodeData: NodeInstance): void {
        const definition = this.nodeDefinition;
        if (!definition) {
            return;
        }

        const bound = nodeData.dynamicTypes ?? {};
        const carrying = new Set<string>();

        const settle = (port: Port, declared: NodeInput | NodeOutput | undefined) => {
            if (!declared?.dynamic) {
                return;
            }
            const concrete = resolveDynamicType(declared.type, bound);
            port.setDataType(concrete ?? shortDynamicPattern(declared.type));
            port.setAnyType(concrete === undefined);
            if (concrete) {
                carrying.add(concrete);
            }
        };

        for (let i = 0; i < this.flowNode.inputs(); i++) {
            const port = this.flowNode.inputPort(i);
            settle(port, definition.inputs?.[port.getDisplayName()]);
        }

        for (let i = 0; i < this.flowNode.outputs(); i++) {
            const port = this.flowNode.outputPort(i);
            settle(port, definition.outputs?.[port.getDisplayName()]);

            // A lifted output turns into an array as soon as an array is
            // wired into its node, so the type belongs to the instance
            // rather than to the node type every instance was stamped from.
            const actual = nodeData.output?.[port.getDisplayName()]?.type;
            if (actual && actual !== port.getDataType()) {
                port.settleOn(actual);
            }
        }

        this.showDynamicType(carrying);
    }

    // A node with dynamic ports looks identical whatever it is carrying, so
    // the types it settled on go in its info where they can be read.
    showDynamicType(carrying: Set<string>): void {
        if (carrying.size === 0) {
            return;
        }

        const short = Array.from(carrying).map((type) => type.split("/").pop()).join(", ");
        const info = this.nodeDefinition?.info ?? "";
        this.flowNode.setInfo(info ? `${info}\n\nCarrying ${short}` : `Carrying ${short}`);
    }

    update(nodeData: NodeInstance): void {
        this.name = nodeData.name;
        this.dependencies = nodeData.assignedInput;
        this.applyDynamicTypes(nodeData);

        if (nodeData.metadata) {
            if (nodeData.metadata.position) {
                this.flowNode.setPosition(nodeData.metadata.position);
            }
        }

        if (nodeData.parameter) {
            this.parameter.update(nodeData.parameter)
        }

        if (this.boundary) {
            this.boundary.update(nodeData);
        }

        if (this.forEachToggle) {
            this.forEach = !!nodeData.forEach;
            this.forEachToggle.setValue(this.forEach);
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