import { FlowNode, FlowNodeConfig, FlowNodeStyle, NodeFlowGraph, Publisher } from "@elicdavis/node-flow";
import { InstanceIDProperty, PolyNodeController } from "./nodes/node";
import { RequestManager } from "./requests";
import {
  BoundaryType,
  GraphExecutionReport,
  GraphInstance,
  GraphInstanceNodes,
  NodeDefinition,
  NodeInstance,
  RegisteredTypes,
  subGraphBoundaryInfo,
  subGraphBoundaryKind,
} from "./schema";
import { ThreeApp } from "./three_app";
import { ProducerViewManager } from './ProducerView/producer_view_manager';
import { getScopedNodes, getScopedProducers } from "./graphScope";
import {
  GraphScopeKind,
  ROOT_SCOPE,
  SUBGRAPH_INPUT_TYPE,
  SUBGRAPH_OUTPUT_TYPE,
    isSubGraphRuntimeType,
    scopeToApiPath,
    subGraphRuntimeType,
    type GraphScope,
} from "./portTypes";
import {
  SubGraphRuntimeStyle,
  buildBoundaryFlowNodeConfig,
  subGraphNodeConfigs,
} from "@/features/nodeFlow/subGraphNodeConfigs";
import { portTypePickerActions } from "@/stores/portTypePickerStore";

export const GeneratorVariablePublisherPath = "Generator/Variable/";


const VariableNodeBackgroundColor = "#233";
const VariableColorScheme: FlowNodeStyle = {
    title: {
        color: "#355"
    },
    idle: {
        color: VariableNodeBackgroundColor,
    },
    mouseOver: {
        color: VariableNodeBackgroundColor,
    },
    grabbed: {
        color: VariableNodeBackgroundColor,
    },
    selected: {
        color: VariableNodeBackgroundColor,
    }
}

const ManifestNodeBackgroundColor = "#332233";
const ManifestColorScheme: FlowNodeStyle = {
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

    serverUpdatingNodeConnections: boolean;

    serializableOutputTypes: Array<string>;

    graphScope: GraphScope = ROOT_SCOPE;

    registeredTypes: RegisteredTypes;

    private onSchemaRefreshNeeded: (() => void) | null = null;

    private boundaryMenuRegistered = false;

    constructor(
        nodeFlowGraph: NodeFlowGraph,
        requestManager: RequestManager,
        nodesPublisher: Publisher,
        app: ThreeApp,
        producerViewManager: ProducerViewManager,
        registeredTypes: RegisteredTypes
    ) {
        this.nodeFlowGraph = nodeFlowGraph;
        this.requestManager = requestManager;
        this.nodesPublisher = nodesPublisher;
        this.app = app;
        this.producerViewManager = producerViewManager;

        this.nodeIdToNode = new Map<string, PolyNodeController>();
        this.nodeTypeToFlowNodePath = new Map<string, string>();
        this.producerTypes = new Map<string, string>();
        this.subscribers = [];
        this.serverUpdatingNodeConnections = false;

        this.registerNodeTypesFromRegistry(registeredTypes);

        nodeFlowGraph.addOnNodeAddedListener(this.onNodeAddedCallback.bind(this));
        nodeFlowGraph.addOnNodeRemovedListener(this.nodeRemoved.bind(this));

        this.serializableOutputTypes = registeredTypes.serializableOutputTypes;
    }

    /** Registers node types from /node-types, including dynamic sub-graph runtime types (subgraph/*). */
    registerNodeTypesFromRegistry(registeredTypes: RegisteredTypes): void {
        this.registeredTypes = registeredTypes;
        registeredTypes.nodeTypes.forEach((type) => {
            if (type.parameter) {
                return;
            }
            if (type.type === SUBGRAPH_INPUT_TYPE || type.type === SUBGRAPH_OUTPUT_TYPE) {
                return;
            }
            this.registerCustomNodeType(type);
        });
    }

    setGraphScope(scope: GraphScope): void {
        this.graphScope = scope;
        this.requestManager.setGraphScopePath(scopeToApiPath(scope));
        this.syncBoundaryNodeMenuAvailability();
    }

    /** Input/Output boundary nodes are only creatable while editing a sub-graph. */
    private syncBoundaryNodeMenuAvailability(): void {
        const shouldRegister = this.graphScope.kind === GraphScopeKind.SubGraph;
        if (shouldRegister === this.boundaryMenuRegistered) {
            return;
        }
        this.boundaryMenuRegistered = shouldRegister;

        for (const [path, config] of Object.entries(subGraphNodeConfigs)) {
            if (shouldRegister) {
                this.nodesPublisher.register(path, config);
            } else {
                this.nodesPublisher.unregister(path);
            }
        }
    }

    getScopeApiPath(): string | null {
        if (this.graphScope.kind === GraphScopeKind.Root) return null;
        return `subgraph/${this.graphScope.id}`;
    }

    setOnSchemaRefreshNeeded(callback: () => void): void {
        this.onSchemaRefreshNeeded = callback;
    }

    notifySubGraphDefinitionChanged(nodeType?: NodeDefinition): void {
        if (nodeType) {
            this.registerCustomNodeType(nodeType);
        }
        this.onSchemaRefreshNeeded?.();
    }

    unregisterRuntimeSubGraphType(subGraphId: string): void {
        const typePath = subGraphRuntimeType(subGraphId);
        const publisherPath = this.nodeTypeToFlowNodePath.get(typePath);
        if (publisherPath) {
            this.unregisterNodeType(publisherPath);
        }
    }

    refreshRuntimeSubGraphType(subGraphId: string, onComplete?: () => void): void {
        this.requestManager.getNodeTypes((types) => {
            const nodeType = types.nodeTypes.find(
                (entry) => entry.type === subGraphRuntimeType(subGraphId)
            );
            if (nodeType) {
                this.registerCustomNodeType(nodeType);
            }
            onComplete?.();
        });
    }

    clearCanvasNodes(): void {
        this.serverUpdatingNodeConnections = true;
        this.nodeIdToNode.forEach((controller) => {
            this.nodeFlowGraph.removeNode(controller.flowNode);
        });
        this.nodeIdToNode.clear();
        this.serverUpdatingNodeConnections = false;
    }

    switchGraphScope(scope: GraphScope, schema: GraphInstance): void {
        this.setGraphScope(scope);
        this.clearCanvasNodes();
        this.updateNodes(schema);
    }

    /**
     * Write live canvas positions into the in-memory schema for the current
     * scope. Drag-stop persists to the server, but tab switches rebuild from
     * schemaManager.currentGraph — which otherwise stays stale until refresh.
     */
    syncLivePositionsIntoSchema(schema: GraphInstance): void {
        const scopedNodes = getScopedNodes(schema, this.graphScope);
        this.nodeIdToNode.forEach((controller, nodeId) => {
            const nodeData = scopedNodes[nodeId];
            if (!nodeData) {
                return;
            }
            const pos = controller.flowNode.getPosition();
            if (!nodeData.metadata) {
                nodeData.metadata = {};
            }
            nodeData.metadata.position = {
                x: Math.round(pos.x),
                y: Math.round(pos.y),
            };
        });
    }

    refreshExecutionReport(): void {
        this.requestManager.getExecutionReport((executionReport: GraphExecutionReport) => {

            for (let nodeID in executionReport.nodes) {
                const nodeExecutionReport = executionReport.nodes[nodeID];
                if (!this.nodeIdToNode.has(nodeID)) {
                    return;
                }
                const nodeToUpdate = this.nodeIdToNode.get(nodeID);
                nodeToUpdate.flowNode.clearMessages();

                for (let output in nodeExecutionReport.output) {
                    nodeToUpdate.setOutputPortReport(output, nodeExecutionReport.output[output]);
                }
            }
        });
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

        const nodeType: string = flowNode.metadata().typeData.type

        if (nodeType === SUBGRAPH_INPUT_TYPE || nodeType === SUBGRAPH_OUTPUT_TYPE) {
            this.beginBoundaryNodeCreation(flowNode, nodeType);
            return;
        }

        this.finishNodeCreation(flowNode, nodeType);
    }

    private removeUnfinishedCanvasNode(flowNode: FlowNode): void {
        this.serverUpdatingNodeConnections = true;
        this.nodeFlowGraph.removeNode(flowNode);
        this.serverUpdatingNodeConnections = false;
    }

    private beginBoundaryNodeCreation(flowNode: FlowNode, nodeType: string): void {
        const portTypes = this.registeredTypes?.portTypes ?? [];
        if (portTypes.length === 0) {
            this.removeUnfinishedCanvasNode(flowNode);
            return;
        }

        const kind =
            nodeType === SUBGRAPH_INPUT_TYPE ? BoundaryType.Input : BoundaryType.Output;

        portTypePickerActions.show({
            title: kind === BoundaryType.Input ? "Input Port Type" : "Output Port Type",
            options: portTypes,
            current: portTypes[0],
            onCancel: () => {
                this.removeUnfinishedCanvasNode(flowNode);
            },
            onSelect: (portType) => {
                const typedNode = this.replacePlaceholderBoundaryNode(
                    flowNode,
                    kind,
                    portType,
                );
                this.finishNodeCreation(typedNode, nodeType, portType, (nodeID, node) => {
                    const portName = node.title();
                    if (!portName.trim()) {
                        return;
                    }
                    this.requestManager.setBoundaryNodeInfo(
                        nodeID,
                        { portName, scope: this.getScopeApiPath() },
                        (resp) => {
                            this.notifySubGraphDefinitionChanged(resp?.nodeType);
                        },
                    );
                });
            },
        });
    }

    /** Swap the menu-created placeholder for a node whose ports use the chosen type. */
    private replacePlaceholderBoundaryNode(
        oldFlowNode: FlowNode,
        kind: BoundaryType,
        portType: string,
    ): FlowNode {
        const position = oldFlowNode.getPosition();
        const title = oldFlowNode.title();
        const config = buildBoundaryFlowNodeConfig(kind, portType);

        this.serverUpdatingNodeConnections = true;
        this.nodeFlowGraph.removeNode(oldFlowNode);

        const newFlowNode = new FlowNode({
            ...config,
            title: title || config.title,
            position,
        });
        this.nodeFlowGraph.addNode(newFlowNode);
        newFlowNode.setProperty("portType", portType);
        this.serverUpdatingNodeConnections = false;
        return newFlowNode;
    }

    private finishNodeCreation(
        flowNode: FlowNode,
        nodeType: string,
        portType?: string,
        afterCreate?: (nodeID: string, flowNode: FlowNode) => void,
    ): void {
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
                    this.app,
                    producerOutPort,
                    this.requestManager,
                    this.producerViewManager,
                    flowNode.metadata().typeData,
                    this.serializableOutputTypes,
                )
            );

            afterCreate?.(nodeID, flowNode);
        }, portType)
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

        nodes.forEach(node => {
            const nodeID = node.id;
            const nodeData = node.node;

            if (this.nodeIdToNode.has(nodeID)) {
                const nodeToUpdate = this.nodeIdToNode.get(nodeID);
                nodeToUpdate.updateConnections(nodeData);
            }
        })
    }

    nodeTypeIsProducer(typeData: NodeDefinition): string {
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

    // getFlowNodeConfig(nodePublisherPath: string): FlowNodeConfig {
    //     return this.nodesPublisher.nodes().get(nodePublisherPath);
    // }

    // updateVariableInfo(originalPublisherID: string, newPublisherID: string, newName: string, newDescription: string): void {
    //     const originalDefinition = this.nodesPublisher.nodes().get(originalPublisherID);
    //     originalDefinition.title = newName;
    //     originalDefinition.info = newDescription;
    //     originalDefinition.metadata.typeData.type = newName;

    //     this.unregisterNodeType(originalPublisherID);
    //     this.nodesPublisher.register(newPublisherID, originalDefinition);
    //     this.nodeTypeToFlowNodePath.set(newName, "generator/variable/" + newName);
    //     console.log(originalDefinition);

    //     this.nodeIdToNode.forEach((controller, nodeId) => {
    //         controller.flowNode.metadata()
    //     })
    // }

    unregisterNodeType(nodePublisherPath: string): void {
        if (!this.nodesPublisher.unregister(nodePublisherPath)) {
            console.error("Failed to unregister", nodePublisherPath);
            return;
        }

        this.nodeTypeToFlowNodePath.forEach((nodeType: string, flowNodePath: string) => {
            if (flowNodePath === nodePublisherPath) {
                console.log(nodeType, flowNodePath)
                this.nodeTypeToFlowNodePath.delete(nodeType);
            }
        });
    }

    registerCustomNodeType(nodeDefinition: NodeDefinition): void {
        const isRuntimeSubGraph = isSubGraphRuntimeType(nodeDefinition.type);

        const existingPath = this.nodeTypeToFlowNodePath.get(nodeDefinition.type);
        if (existingPath) {
            this.unregisterNodeType(existingPath);
        }

        const nodeConfig: FlowNodeConfig = {
            title: nodeDefinition.displayName, //camelCaseToWords(typeData.displayName),
            subTitle: nodeDefinition.path,
            info: nodeDefinition.info,
            inputs: [],
            outputs: [],
            metadata: {
                typeData: nodeDefinition
            },
            canEditTitle: false,
            style: isRuntimeSubGraph ? SubGraphRuntimeStyle : null
        };

        for (let inputName in nodeDefinition.inputs) {
            nodeConfig.inputs.push({
                name: inputName,
                type: nodeDefinition.inputs[inputName].type,
                array: nodeDefinition.inputs[inputName].isArray,
                description: nodeDefinition.inputs[inputName].description
            });
        }

        const isVariable = nodeDefinition.path === "generator/variable";
        const isProducer = this.nodeTypeIsProducer(nodeDefinition);
        if (isProducer) {
            this.producerTypes.set(nodeDefinition.type, isProducer);
        }

        if (nodeDefinition.outputs) {
            for (let outName in nodeDefinition.outputs) {
                nodeConfig.outputs.push({
                    name: outName,
                    type: nodeDefinition.outputs[outName].type,
                    description: nodeDefinition.outputs[outName].description
                });
            }
        }

        if (isProducer) {
            nodeConfig.style = ManifestColorScheme;
            nodeConfig.canEditTitle = true;
        }

        if (isVariable) {
            nodeConfig.style = VariableColorScheme;
        }

        // nm.onNodeCreateCallback(this, typeData.type);

        // const category = this.convertPathToUppercase(typeData.path) + "/" + camelCaseToWords(typeData.displayName);
        if (isRuntimeSubGraph) {
            const category = "SubGraph/" + nodeDefinition.displayName;
            this.nodeTypeToFlowNodePath.set(nodeDefinition.type, category);
            this.nodesPublisher.register(category, nodeConfig);
            return;
        }

        const category = this.convertPathToUppercase(nodeDefinition.path) + "/" + nodeDefinition.displayName;
        this.nodeTypeToFlowNodePath.set(nodeDefinition.type, category);
        this.nodesPublisher.register(category, nodeConfig);
    }

    private boundaryNodeKind(nodeData: NodeInstance): BoundaryType | null {
        return subGraphBoundaryKind(nodeData) ?? null;
    }

    createBoundaryFlowNode(nodeData: NodeInstance): FlowNode {
        const kind = this.boundaryNodeKind(nodeData)!;
        const boundary = subGraphBoundaryInfo(nodeData);
        const portType = boundary?.portType ?? "";
        const config = buildBoundaryFlowNodeConfig(kind, portType);
        return new FlowNode({
            ...config,
            title: boundary?.portName || config.title,
        });
    }

    newNode(nodeData: NodeInstance): FlowNode {
        const isParameter = !!nodeData.parameter;
        const isVariable = !!nodeData.variable;

        const boundaryKind = this.boundaryNodeKind(nodeData);
        if (boundaryKind) {
            return this.createBoundaryFlowNode(nodeData);
        }

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
            return this.nodesPublisher.create(GeneratorVariablePublisherPath + nodeData.name);
        }

        throw new Error("what tf is this.")
    }

    updateNodes(newSchema: GraphInstance): void {
        const scopedNodes = getScopedNodes(newSchema, this.graphScope);
        const scopedProducers = getScopedProducers(newSchema, this.graphScope);
        const sortedNodes = this.sortNodesByName(scopedNodes);

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
                this.nodeIdToNode.get(nodeID).update(nodeData);
            } else {
                const flowNode = this.newNode(nodeData);

                for (const [key, value] of Object.entries(scopedProducers)) {
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
                    this.app,
                    producerOutPort,
                    this.requestManager,
                    this.producerViewManager,
                    flowNode.metadata().typeData,
                    this.serializableOutputTypes,
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