import { PolyNode } from "./nodes/node.js";

function ImageParameterNode() {
    //     this.addInput(inputName, nodeData.inputs[inputName].type);
    this.addOutput("Value", "image.Image");
    // this.properties = { precision: 1 };
    this.title = "Image";
    this.color = "#233";
    this.bgcolor = "#355";

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

    // this.imgWidget.mouse = (event, pos, node) => {
    //     if (event.type == LiteGraph.pointerevents_method + "down") {
    //         w.value = !w.value;
    //         setTimeout(function () {
    //             inner_value_change(w, w.value);
    //         }, 20);
    //     }
    // }
}

export class NodeManager {
    constructor(app) {
        this.app = app;
        this.guiFolderData = {};
        this.nodeIdToNode = new Map();
        this.subscribers = [];

        this.registerPolyformNodes();
    }

    registerPolyformNodes() {
        function Vector3ParameterNode() {
            //     this.addInput(inputName, nodeData.inputs[inputName].type);
            this.addOutput("Value", "github.com/EliCDavis/vector/vector3.Vector[float64]");
            // this.properties = { precision: 1 };
            this.title = "Vector3";
            this.color = "#233";
            this.bgcolor = "#355";
        }

        function AABBParameterNode() {
            //     this.addInput(inputName, nodeData.inputs[inputName].type);
            this.addOutput("Value", "github.com/EliCDavis/polyform/math/geometry.AABB");
            // this.properties = { precision: 1 };
            this.title = "AABB";
            this.color = "#233";
            this.bgcolor = "#355";
        }


        LiteGraph.registerNodeType("polyform/aabb", AABBParameterNode);
        LiteGraph.registerNodeType("polyform/vector3", Vector3ParameterNode);
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

    updateNodes(newSchema) {
        const sortedNodes = this.sortNodesByName(newSchema.nodes);

        let nodeAdded = false;
        for (let node of sortedNodes) {
            const nodeID = node[0];
            const nodeData = node[1];

            if (this.nodeIdToNode.has(nodeID)) {
                const nodeToUpdate = this.nodeIdToNode.get(nodeID);
                nodeToUpdate.update(nodeData);
            } else {
                const isProducer = newSchema.producers.includes(nodeData.name);
                this.nodeIdToNode.set(nodeID, new PolyNode(this, nodeID, nodeData, this.app, this.guiFolderData, isProducer));
                nodeAdded = true;
            }
        }

        this.updateNodeConnections(sortedNodes);

        if (nodeAdded) {
            lgraphInstance.arrange();
        }
    }

    subscribeToParameterChange(subscriber) {
        this.subscribers.push(subscriber)
    }

    nodeParameterChanged(change) {
        this.subscribers.forEach((e) => e(change))
    }
}