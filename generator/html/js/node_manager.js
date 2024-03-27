import { PolyNode } from "./nodes/node.js";

function ImageParameterNode() {
    //     this.addInput(inputName, nodeData.inputs[inputName].type);
    this.addOutput("Value", "image.Image");
    // this.properties = { precision: 1 };
    this.title = "Image";
    this.color = "#233";
    this.bgcolor = "#355";

    // const w = this.addWidget("image", "Image", true, { property: "surname" }); //this will modify the node.properties
    // w.draw = (ctx, node, widget_width, y, H) => {

    //     // const H = LiteGraph.NODE_WIDGET_HEIGHT;
    //     var show_text = lightCanvas.ds.scale > 0.5;
    //     // ctx.save();
    //     // ctx.globalAlpha = this.editor_alpha;
    //     const outline_color = LiteGraph.WIDGET_OUTLINE_COLOR;
    //     const background_color = LiteGraph.WIDGET_BGCOLOR;
    //     const text_color = LiteGraph.WIDGET_TEXT_COLOR;
    //     const secondary_text_color = LiteGraph.WIDGET_SECONDARY_TEXT_COLOR;
    //     const margin = 15;

    //     ctx.textAlign = "left";
    //     ctx.strokeStyle = outline_color;
    //     ctx.fillStyle = background_color;
    //     ctx.beginPath();
    //     if (show_text)
    //         ctx.roundRect(margin, y, widget_width - margin * 2, H, [H * 0.5]);
    //     else
    //         ctx.rect(margin, y, widget_width - margin * 2, H);
    //     ctx.fill();
    //     if (show_text && !w.disabled)
    //         ctx.stroke();
    //     ctx.fillStyle = w.value ? "#89A" : "#333";
    //     ctx.beginPath();
    //     ctx.arc(widget_width - margin * 2, y + H * 0.5, H * 0.36, 0, Math.PI * 2);
    //     ctx.fill();
    //     if (show_text) {
    //         ctx.fillStyle = secondary_text_color;
    //         const label = w.label || w.name;
    //         if (label != null) {
    //             ctx.fillText(label, margin * 2, y + H * 0.7);
    //         }
    //         ctx.fillStyle = w.value ? text_color : secondary_text_color;
    //         ctx.textAlign = "right";
    //         ctx.fillText(
    //             w.value
    //                 ? w.options.on || "true"
    //                 : w.options.off || "false",
    //             widget_width - 40,
    //             y + H * 0.7
    //         );
    //     }
    // }

    // w.mouse = (event, pos, node) => {
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