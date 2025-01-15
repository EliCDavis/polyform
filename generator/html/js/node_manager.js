import { PolyNodeController, camelCaseToWords } from "./nodes/node.js";


export class NodeManager {
    constructor(app) {
        this.app = app;
        this.guiFolderData = {};
        this.nodeIdToNode = new Map();
        this.nodeTypeToLitePath = new Map();
        this.subscribers = [];
        this.initializedNodeTypes = false
        // this.registerSpecialParameterPolyformNodes();

        this.app.NodeFlowGraph.addOnNodeAddedListener(this.onNodeAddedCallback.bind(this));
    }

    onNodeAddedCallback(flowNode) {

        if (this.app.ServerUpdatingNodeConnections) {
            return;
        }

        console.log(flowNode.metadata())
        const nodeType = flowNode.metadata().typeData.type
        console.log(nodeType, flowNode)

        this.app.RequestManager.createNode(nodeType, (resp) => {
            const isProducer = false;
            const nodeID = resp.nodeID
            const nodeData = resp.data;

            // TODO: This is an ugo hack. Consider somehow making this apart of the metadata.
            flowNode.nodeInstanceID = nodeID;

            this.nodeIdToNode.set(nodeID, new PolyNodeController(flowNode, this, nodeID, nodeData, this.app, isProducer));
        })
    }

    sortNodesByName(nodesToSort) {
        const sortable = [];
        for (let nodeID in nodesToSort) {
            sortable.push([nodeID, nodesToSort[nodeID]]);
        }

        sortable.sort((a, b) => {
            const textA = a[1].name.toUpperCase();
            const textB = b[1].name.toUpperCase();
            return (textA < textB) ? -1 : (textA > textB) ? 1 : 0;
        });
        return sortable;
    }

    updateNodeConnections(nodes) {
        // console.log("nodes", nodes)
        for (let node of nodes) {
            const nodeID = node[0];
            const nodeData = node[1];
            const inNode = this.nodeIdToNode.get(nodeID);

            for (let i = 0; i < nodeData.dependencies.length; i++) {
                const dep = nodeData.dependencies[i];
                let dependencyName = dep.name;

                // Inputs that are elements in array are named "Input.N"
                if (dependencyName.indexOf(".") !== -1) {
                    dependencyName = dependencyName.split(".")[0]
                }

                const outNode = this.nodeIdToNode.get(dep.dependencyID);

                let sourceInput = -1;
                // console.log(source.flowNode)
                for (let sourceInputIndex = 0; sourceInputIndex < inNode.flowNode.inputs(); sourceInputIndex++) {
                    if (inNode.flowNode.inputPort(sourceInputIndex).getDisplayName() === dependencyName) {
                        sourceInput = sourceInputIndex;
                    }
                }

                if (sourceInput === -1) {
                    console.error("failed to find source for ", dep)
                    continue;
                }

                // connectNodes(nodeOut: FlowNode, outPort: number, nodeIn: FlowNode, inPort): Connection | undefined {
                // TODO: This only works for nodes with one output
                this.app.NodeFlowGraph.connectNodes(
                    outNode.flowNode, 0,
                    inNode.flowNode, sourceInput,
                )
            }
        }
    }


    registerCustomNodeType(typeData) {
        const nodeConfig = {
            title: camelCaseToWords(typeData.displayName),
            inputs: [],
            outputs: [],
            metadata: {
                typeData: typeData
            }
        }

        for (let inputName in typeData.inputs) {
            nodeConfig.inputs.push({
                name: inputName,
                type: typeData.inputs[inputName].type,
                array: typeData.inputs[inputName].isArray
            });
        }

        let isProducer = false;

        if (typeData.outputs) {
            typeData.outputs.forEach((o) => {
                nodeConfig.outputs.push({ name: o.name, type: o.type });

                if (o.type === "github.com/EliCDavis/polyform/generator/artifact.Artifact") {
                    isProducer = true;
                }
            })
        }

        if (isProducer) {
            const ParameterNodeBackgroundColor = "#332233";
            const ParameterStyle = {
                title: {
                    color: "#4a3355"
                },
                idle: {
                    color: ParameterNodeBackgroundColor,
                },
                mouseOver: {
                    color: ParameterNodeBackgroundColor,
                },
                grabbed: {
                    color: ParameterNodeBackgroundColor,
                },
                selected: {
                    color: ParameterNodeBackgroundColor,
                }
            }
            nodeConfig.style = ParameterStyle

            nodeConfig.canEditTitle = true;
        }

        // nm.onNodeCreateCallback(this, typeData.type);

        const category = typeData.path + "/" + typeData.displayName;
        this.nodeTypeToLitePath.set(typeData.type, category);
        PolyformNodesPublisher.register(category, nodeConfig);
    }

    newNode(nodeData) {
        const isParameter = !!nodeData.parameter;

        // Not a parameter, just create a node that adhere's to the server's
        // reflection.
        // if (!isParameter) {
        //     const nodeIdentifier = this.nodeTypeToLitePath.get(nodeData.type)
        //     return LiteGraph.createNode(nodeIdentifier);
        // }

        if (!isParameter) {
            const nodeIdentifier = this.nodeTypeToLitePath.get(nodeData.type)
            return PolyformNodesPublisher.create(nodeIdentifier);
        }

        const parameterType = nodeData.parameter.type;
        return PolyformNodesPublisher.create("parameters/" + parameterType);
    }

    updateNodes(newSchema) {

        // Only need to do this once, since types are set at compile time. If
        // that ever changes, god.
        if (this.initializedNodeTypes === false) {
            this.initializedNodeTypes = true;
            newSchema.types.forEach(type => {
                // We should have custom nodes already defined for parameters
                if (type.parameter) {
                    return;
                }
                this.registerCustomNodeType(type)
            })
        }

        const sortedNodes = this.sortNodesByName(newSchema.nodes);

        this.app.ServerUpdatingNodeConnections = true;

        let nodeAdded = false;
        for (let node of sortedNodes) {
            const nodeID = node[0];
            const nodeData = node[1];

            if (this.nodeIdToNode.has(nodeID)) {
                const nodeToUpdate = this.nodeIdToNode.get(nodeID);
                nodeToUpdate.update(nodeData);
            } else {
                const flowNode = this.newNode(nodeData);

                let isProducer = false;
                for (const [key, value] of Object.entries(newSchema.producers)) {
                    if (value.nodeID === nodeID) {
                        isProducer = true;
                        flowNode.setTitle(key);
                    }
                }

                this.app.NodeFlowGraph.addNode(flowNode);

                // TODO: This is an ugo hack. Consider somehow making this
                // apart of the metadata.
                flowNode.nodeInstanceID = nodeID;

                this.nodeIdToNode.set(nodeID, new PolyNodeController(flowNode, this, nodeID, nodeData, this.app, isProducer));
                nodeAdded = true;
            }
        }

        this.updateNodeConnections(sortedNodes);

        if (nodeAdded) {
            // nodeFlowGraph.organize();
        }
        this.app.ServerUpdatingNodeConnections = false;
    }

    subscribeToParameterChange(subscriber) {
        this.subscribers.push(subscriber)
    }

    nodeParameterChanged(change) {
        this.subscribers.forEach((e) => e(change))
    }
}