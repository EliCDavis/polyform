import { FlowNode, FlowNodeConfig, NodeFlowGraph, Publisher } from "@elicdavis/node-flow";
import { InstanceIDProperty, PolyNodeController, camelCaseToWords } from "./nodes/node.js";
import { RequestManager } from "./requests.js";
import { GraphInstance, GraphInstanceNodes, NodeInstance } from "./schema.js";
import { NodeType } from './schema';
import { ThreeApp } from "./three_app.js";
import { ProducerViewManager } from './ProducerView/producer_view_manager';

const ColorSchemes = {
    Green: {
        Background: "#233",
        Title: "#355"
    },
}

interface NodeParameterChangeEvent {
    // Node ID
    id: string,

    // New Parameter Data
    data: any,

    // Whether or not the parameter data is binary
    binary: boolean
}

export class NodeManager {

    // CONFIG >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
    nodeFlowGraph: NodeFlowGraph;

    requestManager: RequestManager;

    nodesPublisher: Publisher;

    app: ThreeApp;

    producerViewManager: ProducerViewManager;

    // RUNTIME >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
    nodeIdToNode: Map<string, PolyNodeController>;

    subscribers: Array<(change: NodeParameterChangeEvent) => void>;

    producerTypes: Map<string, string>;

    nodeTypeToFlowNodePath: Map<string, string>;

    nodeTypeToNodeDefinition: Map<string, NodeType>;

    serverUpdatingNodeConnections: boolean;

    constructor(
        nodeFlowGraph: NodeFlowGraph,
        requestManager: RequestManager,
        nodesPublisher: Publisher,
        app: ThreeApp,
        producerViewManager: ProducerViewManager,
        nodeTypes: Array<NodeType>
    ) {
        this.nodeFlowGraph = nodeFlowGraph;
        this.requestManager = requestManager;
        this.nodesPublisher = nodesPublisher;
        this.app = app;
        this.producerViewManager = producerViewManager;

        this.nodeIdToNode = new Map<string, PolyNodeController>();
        this.nodeTypeToFlowNodePath = new Map<string, string>();
        this.nodeTypeToNodeDefinition = new Map<string, NodeType>();
        this.producerTypes = new Map<string, string>();
        this.subscribers = [];
        this.serverUpdatingNodeConnections = false;
        // this.registerSpecialParameterPolyformNodes();

        nodeTypes.forEach(type => {
            // We should have custom nodes already defined for parameters
            if (type.parameter) {
                return;
            }
            this.registerCustomNodeType(type)
        });

        nodeFlowGraph.addOnNodeAddedListener(this.onNodeAddedCallback.bind(this));
        nodeFlowGraph.addOnNodeRemovedListener(this.nodeRemoved.bind(this));
    }

    nodeRemoved(flowNode: FlowNode): void {
        if (this.serverUpdatingNodeConnections) {
            return;
        }

        this.requestManager.deleteNode(flowNode.getProperty(InstanceIDProperty))
    }

    onNodeAddedCallback(flowNode: FlowNode): void {
        if (this.serverUpdatingNodeConnections) {
            return;
        }

        const nodeType = flowNode.metadata().typeData.type

        this.requestManager.createNode(nodeType, (resp) => {
            const nodeID = resp.nodeID
            const nodeData = resp.data;

            flowNode.setProperty(InstanceIDProperty, nodeID);

            let producerOutPort: string = null
            if (this.producerTypes.has(nodeType)) {
                producerOutPort = this.producerTypes.get(nodeType);
            }

            this.nodeIdToNode.set(
                nodeID,
                new PolyNodeController(
                    flowNode,
                    this,
                    nodeID,
                    nodeData,
                    this.nodeTypeToNodeDefinition.get(nodeType),
                    this.app,
                    producerOutPort,
                    this.requestManager,
                    this.producerViewManager
                )
            );
        })
    }

    sortNodesByName(nodesToSort: GraphInstanceNodes): Array<{ id: string, node: NodeInstance }> {
        const sortable = new Array<{ id: string, node: NodeInstance }>();
        for (let nodeID in nodesToSort) {
            sortable.push({
                id: nodeID,
                node: nodesToSort[nodeID]
            });
        }

        sortable.sort((a, b) => {
            const textA = a.node.name.toUpperCase();
            const textB = b.node.name.toUpperCase();
            return (textA < textB) ? -1 : (textA > textB) ? 1 : 0;
        });
        return sortable;
    }

    findIndexOfInputPortWithName(node: FlowNode, portName: string): number {
        for (let i = 0; i < node.inputs(); i++) {
            if (node.inputPort(i).getDisplayName() === portName) {
                return i;
            }
        }
        return -1;
    }

    findIndexOfOutputPortWithName(node: FlowNode, portName: string): number {
        for (let i = 0; i < node.outputs(); i++) {
            if (node.outputPort(i).getDisplayName() === portName) {
                return i;
            }
        }
        return -1;
    }

    updateNodeConnections(nodes: Array<{ id: string, node: NodeInstance }>): void {
        for (let node of nodes) {
            const nodeID = node.id;
            const nodeData = node.node;
            const nodeController = this.nodeIdToNode.get(nodeID);

            for (const dirtyinputPortName in nodeData.assignedInput) {
                let cleanedInputPortName = dirtyinputPortName;

                // Inputs that are elements in array are named "Input.N"
                if (cleanedInputPortName.indexOf(".") !== -1) {
                    cleanedInputPortName = cleanedInputPortName.split(".")[0]
                }

                const inputPort = nodeData.assignedInput[dirtyinputPortName];
                const inputPortIndex = this.findIndexOfInputPortWithName(nodeController.flowNode, cleanedInputPortName);
                if (inputPortIndex === -1) {
                    console.error("failed to find source for ", inputPort)
                    continue;
                }

                const otherNode = this.nodeIdToNode.get(inputPort.id);
                const outputPortIndex = this.findIndexOfOutputPortWithName(otherNode.flowNode, inputPort.port);
                if (outputPortIndex === -1) {
                    console.error("failed to find output port", inputPort.port, "on node", otherNode)
                    continue;
                }

                this.nodeFlowGraph.connectNodes(
                    otherNode.flowNode, outputPortIndex,
                    nodeController.flowNode, inputPortIndex,
                )
            }
        }
    }

    nodeTypeIsProducer(typeData: NodeType): string {
        if (typeData.outputs) {
            for (const output in typeData.outputs) {
                if (typeData.outputs[output].type === "github.com/EliCDavis/polyform/generator/manifest.Manifest") {
                    return output;
                }
            }
        }
        return null
    }

    convertPathToUppercase(dirtyPath: string): string {
        let cleanPath = "";
        let capatilize = true;
        for (let i = 0; i < dirtyPath.length; i++) {
            const char = dirtyPath.charAt(i);
            if (capatilize) {
                cleanPath += char.toUpperCase();
                capatilize = false;
                continue;
            }

            if (char === "/") {
                capatilize = true;
            }

            cleanPath += char;
        }
        return cleanPath;
    }

    registerCustomNodeType(typeData: NodeType): void {
        const nodeConfig: FlowNodeConfig = {
            title: typeData.displayName, //camelCaseToWords(typeData.displayName),
            subTitle: typeData.path,
            info: typeData.info,
            inputs: [],
            outputs: [],
            metadata: {
                typeData: typeData
            },
            canEditTitle: false,
            style: null
        };

        this.nodeTypeToNodeDefinition.set(typeData.type, typeData);

        for (let inputName in typeData.inputs) {
            nodeConfig.inputs.push({
                name: inputName,
                type: typeData.inputs[inputName].type,
                array: typeData.inputs[inputName].isArray,
                description: typeData.inputs[inputName].description
            });
        }

        const isVariable = typeData.path === "generator/variable";
        const isProducer = this.nodeTypeIsProducer(typeData);
        if (isProducer) {
            this.producerTypes.set(typeData.type, isProducer);
        }

        if (typeData.outputs) {
            for (let outName in typeData.outputs) {
                nodeConfig.outputs.push({
                    name: outName,
                    type: typeData.outputs[outName].type,
                    description: typeData.outputs[outName].description
                });
            }
        }

        if (isProducer) {
            const ManifestNodeBackgroundColor = "#332233";
            nodeConfig.style = {
                title: {
                    color: "#4a3355"
                },
                idle: {
                    color: ManifestNodeBackgroundColor,
                },
                mouseOver: {
                    color: ManifestNodeBackgroundColor,
                },
                grabbed: {
                    color: ManifestNodeBackgroundColor,
                },
                selected: {
                    color: ManifestNodeBackgroundColor,
                }
            }
            nodeConfig.canEditTitle = true;
        }

        if (isVariable) {
            nodeConfig.style = {
                title: {
                    color: ColorSchemes.Green.Title
                },
                idle: {
                    color: ColorSchemes.Green.Background,
                },
                mouseOver: {
                    color: ColorSchemes.Green.Background,
                },
                grabbed: {
                    color: ColorSchemes.Green.Background,
                },
                selected: {
                    color: ColorSchemes.Green.Background,
                }
            }
        }

        // nm.onNodeCreateCallback(this, typeData.type);

        // const category = this.convertPathToUppercase(typeData.path) + "/" + camelCaseToWords(typeData.displayName);
        const category = this.convertPathToUppercase(typeData.path) + "/" + typeData.displayName;
        this.nodeTypeToFlowNodePath.set(typeData.type, category);
        this.nodesPublisher.register(category, nodeConfig);
    }

    newNode(nodeData: NodeInstance): FlowNode {
        const isParameter = !!nodeData.parameter;
        const isVariable = !!nodeData.variable;

        // Not a parameter, just create a node that adhere's to the server's
        // reflection.
        if (!isParameter && !isVariable) {
            const nodeIdentifier = this.nodeTypeToFlowNodePath.get(nodeData.type)
            return this.nodesPublisher.create(nodeIdentifier);
        }

        if (isParameter) {
            let parameterType = nodeData.parameter.type;
            if (parameterType === "[]uint8") {
                parameterType = "File";
            }
            return this.nodesPublisher.create("Parameters/" + parameterType);
        }

        if (isVariable) {
            let parameterType = nodeData.variable.type;
            if (parameterType === "[]uint8") {
                parameterType = "File";
            }
            return this.nodesPublisher.create("Generator/Variable/" + nodeData.variable.name);
        }

        throw new Error("what tf is this.")
    }

    updateNodes(newSchema: GraphInstance): void {
        const sortedNodes = this.sortNodesByName(newSchema.nodes);

        const nodesSet = new Map<string, boolean>();
        this.nodeIdToNode.forEach((node, nodeId) => {
            nodesSet.set(nodeId, false);
        });

        this.serverUpdatingNodeConnections = true;

        for (let node of sortedNodes) {
            const nodeID = node.id;
            const nodeData = node.node;
            nodesSet.set(nodeID, true);

            if (this.nodeIdToNode.has(nodeID)) {
                const nodeToUpdate = this.nodeIdToNode.get(nodeID);
                nodeToUpdate.update(nodeData);
            } else {
                const flowNode = this.newNode(nodeData);

                for (const [key, value] of Object.entries(newSchema.producers)) {
                    if (value.nodeID === nodeID) {
                        flowNode.setTitle(key);
                    }
                }

                let producerOutPort: string = null
                if (this.producerTypes.has(nodeData.type)) {
                    producerOutPort = this.producerTypes.get(nodeData.type);
                }

                this.nodeFlowGraph.addNode(flowNode);

                flowNode.setProperty(InstanceIDProperty, nodeID);

                const controller = new PolyNodeController(
                    flowNode,
                    this,
                    nodeID,
                    nodeData,
                    this.nodeTypeToNodeDefinition.get(nodeData.type),
                    this.app,
                    producerOutPort,
                    this.requestManager,
                    this.producerViewManager
                );
                this.nodeIdToNode.set(nodeID, controller);
            }
        }

        this.updateNodeConnections(sortedNodes);

        nodesSet.forEach((set, nodeId) => {
            if (set) {
                return;
            }
            console.log("removing node..." + nodeId)
            const node = this.nodeIdToNode.get(nodeId)
            this.nodeFlowGraph.removeNode(node.flowNode);
            this.nodeIdToNode.delete(nodeId);
        });

        this.serverUpdatingNodeConnections = false;
    }

    subscribeToParameterChange(subscriber: (change: NodeParameterChangeEvent) => void): void {
        this.subscribers.push(subscriber)
    }

    nodeParameterChanged(change: NodeParameterChangeEvent): void {
        this.subscribers.forEach((e) => e(change))
    }
}