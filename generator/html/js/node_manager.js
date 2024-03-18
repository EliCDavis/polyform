import { PolyNode } from "./nodes/node.js";

export class NodeManager {
    constructor(app) {
        this.app = app;
        this.guiFolderData = {};
        this.nodeIdToNode = new Map();
        // this.nodeId = new Map();
        this.subscribers = [];
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

    updateNodes(newNodes) {
        const sortedNodes = this.sortNodesByName(newNodes);

        for (let node of sortedNodes) {
            const nodeID = node[0];
            const nodeData = node[1];

            if (this.nodeIdToNode.has(nodeID)) {
                const nodeToUpdate = this.nodeIdToNode.get(nodeID);
                nodeToUpdate.update(nodeData);
            } else {
                this.nodeIdToNode.set(nodeID, new PolyNode(this, nodeID, nodeData, this.app, this.guiFolderData));
            }
        }

        for (let node of sortedNodes) {
            const nodeID = node[0];
            const nodeData = node[1];
            const source = this.nodeIdToNode.get(nodeID);
            
            for (let i = 0; i < nodeData.dependencies.length; i ++) {
                const dep = nodeData.dependencies[i];
                const target = this.nodeIdToNode.get(dep.dependencyID);

                target.lightNode.connect(0, source.lightNode, i)
                // source.lightNode.connect(i, target.lightNode, 0);
            }
        }

        lgraphInstance.arrange();
    }

    subscribeToParameterChange(subscriber) {
        this.subscribers.push(subscriber)
    }

    nodeParameterChanged(change) {
        this.subscribers.forEach((e) => e(change))
    }
}