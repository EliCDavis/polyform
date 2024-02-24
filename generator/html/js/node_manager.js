import { PolyNode } from "./node.js";

export class NodeManager {
    constructor(app) {
        this.app = app;
        this.guiFolderData = {};
        this.nodes = new Map();
        this.subscribers = [];
    }

    updateNodes(newNodes) {
        let sortable = [];
        for (var nodeID in newNodes) {
            sortable.push([nodeID, newNodes[nodeID]]);
        }

        sortable.sort((a, b) => {
            var textA = a[1].name.toUpperCase();
            var textB = b[1].name.toUpperCase();
            return (textA < textB) ? -1 : (textA > textB) ? 1 : 0;
        });


        for (let node of sortable) {
            const nodeID = node[0];
            const nodeData = node[1];

            if (this.nodes.has(nodeID)) {
                const nodeToUpdate = this.nodes.get(nodeID);
                nodeToUpdate.update(nodeData);
            } else {
                this.nodes.set(nodeID, new PolyNode(this, nodeID, nodeData, this.app, this.guiFolderData));
            }
        }
    }

    subscribeToParameterChange(subscriber) {
        this.subscribers.push(subscriber)
    }

    nodeParameterChanged(change) {
        this.subscribers.forEach((e) => e(change))
    }
}