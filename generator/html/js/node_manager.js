import { PolyNodeController, camelCaseToWords } from "./nodes/node.js";


const ParameterNodeOutputPortName = "Out";
const ParameterNodeColor = "#233";
const ParameterNodeBackgroundColor = "#355";

/**
 * 
 * @param {string} dataType 
 * @returns {string}
 */
function ParameterNodeType(dataType) {
    return "github.com/EliCDavis/polyform/parameter.Value[" + dataType + "]";
}


/**
 * 
 * @param {string} subCategory 
 * @returns {string}
 */
function ParameterNamespace(subCategory) {
    return "parameter/" + subCategory
}


export class NodeManager {
    constructor(app) {
        this.app = app;
        this.guiFolderData = {};
        this.nodeIdToNode = new Map();
        this.nodeTypeToLitePath = new Map();
        this.subscribers = [];
        this.initializedNodeTypes = false
        // this.registerSpecialParameterPolyformNodes();
    }

    onNodeCreateCallback(liteNode, nodeType) {
        if (this.app.ServerUpdatingNodeConnections) {
            return;
        }
        this.app.RequestManager.createNode(nodeType, (resp) => {
            const isProducer = false;
            const nodeID = resp.nodeID
            const nodeData = resp.data;
            liteNode.nodeInstanceID = nodeID;
            this.nodeIdToNode.set(nodeID, new PolyNodeController(liteNode, this, nodeID, nodeData, this.app, isProducer));
        })
    }

    sortNodesByName(nodesToSort) {
        let sortable = [];
        for (var nodeID in nodesToSort) {
            sortable.push([nodeID, nodesToSort[nodeID]]);
        }

        sortable.sort((a, b) => {
            var textA = a[1].name.toUpperCase();
            var textB = b[1].name.toUpperCase();
            return (textA < textB) ? -1 : (textA > textB) ? 1 : 0;
        });
        return sortable;
    }

    updateNodeConnections(nodes) {
        for (let node of nodes) {
            const nodeID = node[0];
            const nodeData = node[1];
            const source = this.nodeIdToNode.get(nodeID);

            for (let i = 0; i < nodeData.dependencies.length; i++) {
                const dep = nodeData.dependencies[i];
                const target = this.nodeIdToNode.get(dep.dependencyID);

                let sourceInput = -1;
                // console.log(source.liteNode)
                for (let sourceInputIndex = 0; sourceInputIndex < source.liteNode.inputs(); sourceInputIndex++) {
                    if (source.liteNode.inputPort(sourceInputIndex).getDisplayName() === dep.name) {
                        sourceInput = sourceInputIndex;
                    }
                }

                // connectNodes(nodeOut: FlowNode, outPort: number, nodeIn: FlowNode, inPort): Connection | undefined {

                if (sourceInput === -1) {
                    console.error("failed to find source")
                    continue;
                }

                // TODO: This only works for nodes with one output
                this.app.NodeFlowGraph.connectNodes(
                    target.liteNode, 0,
                    source.liteNode, sourceInput,
                )
                // target.liteNode.connect(0, source.liteNode, sourceInput)
                // source.lightNode.connect(i, target.lightNode, 0);
            }
        }
    }


    buildCustomNodeType(typeData) {
        const nodeConfig = {
            title: camelCaseToWords(typeData.displayName),
            inputs: [],
            outputs:[]
        }

        for (var inputName in typeData.inputs) {
            nodeConfig.inputs.push({
                name: inputName, 
                type: typeData.inputs[inputName].type
            });
        }

        if (typeData.outputs) {
            typeData.outputs.forEach((o) => {
                nodeConfig.outputs.push({
                    name: o.name, 
                    type: o.type
                });
            })
        }
        // nm.onNodeCreateCallback(this, typeData.type);
        
        const category = typeData.path + "/" + typeData.displayName;
        this.nodeTypeToLitePath.set(typeData.type, category);
        PolyformNodesPublisher.register(category, nodeConfig);
    }

    newLiteNode(nodeData) {
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
                this.buildCustomNodeType(type)
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
                let isProducer = false;
                for (const [key, value] of Object.entries(newSchema.producers)) {
                    if (value.nodeID === nodeID) {
                        isProducer = true;
                    }
                }

                const liteNode = this.newLiteNode(nodeData);
                this.app.NodeFlowGraph.addNode(liteNode);
                liteNode.nodeInstanceID = nodeID;

                this.nodeIdToNode.set(nodeID, new PolyNodeController(liteNode, this, nodeID, nodeData, this.app, isProducer));
                nodeAdded = true;
            }
        }

        this.updateNodeConnections(sortedNodes);

        if (nodeAdded) {
            nodeFlowGraph.organize();
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