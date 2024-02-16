import { PolyNode } from "./node.js";

export class NodeManager {
    constructor(app) {
        this.app = app;
        this.guiFolderData = {};
        this.nodes = new Map();
        this.subscribers = [];
    }

    updateNodes(newNodes) {
        for (let nodeID of Object.keys(newNodes)) {
            console.log(nodeID)
            const nodeData = newNodes[nodeID];
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