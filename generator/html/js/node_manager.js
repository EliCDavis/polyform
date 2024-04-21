import { PolyNode, camelCaseToWords } from "./nodes/node.js";


const ParameterNodeOutputPortName = "Out";
const ParameterNodeColor = "#233";
const ParameterNodeBackgroundColor = "#355";

/**
 * 
 * @param {string} dataType 
 * @returns {string}
 */
function ParameterNodeType(dataType) {
    return "github.com/EliCDavis/polyform/generator.ParameterNode[" + dataType + "]";
}

function OnNodeCreateCallback(app, nodeType) {
    return () => {
        if (app.ServerUpdatingNodeConnections) {
            return;
        }
        app.RequestManager.createNode(ParameterNodeType(nodeType))
    }
}


export class NodeManager {
    constructor(app) {
        this.app = app;
        this.guiFolderData = {};
        this.nodeIdToNode = new Map();
        this.nodeTypeToLitePath = new Map();
        this.subscribers = [];
        this.initializedNodeTypes = false
        this.registerSpecialParameterPolyformNodes();
    }


    registerSpecialParameterPolyformNodes() {
        const nm = this;
        function ImageParameterNode() {
            const nodeDataType = "image.Image";
            this.addOutput(ParameterNodeOutputPortName, nodeDataType);
            this.title = "Image";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;
            this.onNodeCreated = OnNodeCreateCallback(nm.app, nodeDataType);

            const H = LiteGraph.NODE_WIDGET_HEIGHT;

            const imgWidget = this.addWidget("image", "Image", true, { property: "surname" }); //this will modify the node.properties
            this.imgWidget = imgWidget;
            const margin = 15;
            this.imgWidget.draw = (ctx, node, widget_width, y, H) => {
                if (!imgWidget.image) {
                    return;
                }

                const adjustedWidth = widget_width - margin * 2
                ctx.drawImage(
                    imgWidget.image,
                    margin,
                    y,
                    adjustedWidth,
                    (adjustedWidth / imgWidget.image.width) * imgWidget.image.height
                );
            }

            this.imgWidget.computeSize = (width) => {
                if (!!imgWidget.image) {
                    const adjustedWidth = width - margin * 2
                    const newH = (adjustedWidth / imgWidget.image.width) * imgWidget.image.height;
                    return [width, newH]
                }
                return [width, 0];
            }
        }

        function ColorParameterNode() {
            const nodeDataType = "github.com/EliCDavis/polyform/drawing/coloring.WebColor";
            this.addOutput(ParameterNodeOutputPortName, nodeDataType);
            this.title = "Color";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;
            this.onNodeCreated = OnNodeCreateCallback(nm.app, nodeDataType);


            const imgWidget = this.addWidget("color", "Color", "#00FF00", {}); //this will modify the node.properties
            this.imgWidget = imgWidget;
            const margin = 15;
            this.imgWidget.draw = (ctx, node, widget_width, y, H) => {
                const adjustedWidth = widget_width - margin * 2
                ctx.beginPath(); // Start a new path
                ctx.rect(margin, y, adjustedWidth, H); // Add a rectangle to the current path
                ctx.fillStyle = this.imgWidget.value;
                ctx.fill(); // Render the path
            }

            // this.imgWidget.mouse = (event, pos, node) => {
            //     if (event.type == LiteGraph.pointerevents_method + "down") {
            //         w.value = !w.value;
            //         setTimeout(function () {
            //             inner_value_change(w, w.value);
            //         }, 20);
            //     }
            // }
        }

        function Vector3ParameterNode() {
            const nodeDataType = "github.com/EliCDavis/vector/vector3.Vector[float64]";
            this.addOutput(ParameterNodeOutputPortName, nodeDataType);
            this.title = "Vector3";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;
            this.onNodeCreated = OnNodeCreateCallback(nm.app, nodeDataType);
        }

        function Vector3ArrayParameterNode() {
            const nodeDataType = "[]github.com/EliCDavis/vector/vector3.Vector[float64]";
            this.addOutput(ParameterNodeOutputPortName, nodeDataType);
            this.title = "Vector3 Array";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;
            this.onNodeCreated = OnNodeCreateCallback(nm.app, nodeDataType);
        }


        function AABBParameterNode() {
            const nodeDataType = "github.com/EliCDavis/polyform/math/geometry.AABB";
            this.addOutput(ParameterNodeOutputPortName, nodeDataType);
            this.title = "AABB";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;
            this.onNodeCreated = OnNodeCreateCallback(nm.app, nodeDataType);
        }

        LiteGraph.registerNodeType("polyform/aabb", AABBParameterNode);
        LiteGraph.registerNodeType("polyform/vector3", Vector3ParameterNode);
        LiteGraph.registerNodeType("polyform/vector3[]", Vector3ArrayParameterNode);
        LiteGraph.registerNodeType("polyform/color", ColorParameterNode);
        LiteGraph.registerNodeType("polyform/Image", ImageParameterNode);
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
                for (let sourceInputIndex = 0; sourceInputIndex < source.lightNode.inputs.length; sourceInputIndex++) {
                    if (source.lightNode.inputs[sourceInputIndex].name === dep.name) {
                        sourceInput = sourceInputIndex;
                    }
                }

                // TODO: This only works for nodes with one output
                target.lightNode.connect(0, source.lightNode, sourceInput)
                // source.lightNode.connect(i, target.lightNode, 0);
            }
        }
    }

    buildCustomNodeType(typeData) {
        const nm = this;
        function FuckYou() {
            for (var inputName in typeData.inputs) {
                this.addInput(inputName, typeData.inputs[inputName].type);
            }

            if (typeData.outputs) {
                typeData.outputs.forEach((o) => {
                    this.addOutput(o.name, o.type);
                })
            }

            // if (producers.includes(nodeData.name)) {
            //     this.color = "#232";
            //     this.bgcolor = "#353";
            //     this.addWidget("button", "Download", null, () => {
            //         console.log("presed");
            //         saveFileToDisk("/producer/" + typeData.displayName, typeData.displayName);
            //     })
            // }
            this.title = camelCaseToWords(typeData.displayName);

            this.onNodeCreated = () => {
                if (nm.app.ServerUpdatingNodeConnections) {
                    return;
                }
                nm.app.RequestManager.createNode(typeData.type)
                console.log("node created: ", typeData.type)
            }
        }

        Object.defineProperty(FuckYou, "name", { value: typeData.displayName });

        const category = typeData.path + "/" + typeData.displayName;
        LiteGraph.registerNodeType(category, FuckYou);
        this.nodeTypeToLitePath.set(typeData.type, category);

        // const node = LiteGraph.createNode(nodeName);
        // node.setSize(node.computeSize());

        // app.LightGraph.add(node);
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
                this.nodeIdToNode.set(nodeID, new PolyNode(this, nodeID, nodeData, this.app, this.guiFolderData, isProducer));
                nodeAdded = true;
            }
        }

        this.updateNodeConnections(sortedNodes);

        if (nodeAdded) {
            lgraphInstance.arrange();
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