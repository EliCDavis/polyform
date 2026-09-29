import { FlowNode, InternalConnection, NodeFlowGraph, Publisher } from "@elicdavis/node-flow";
import { InstanceIDProperty } from "./nodes/node";
import { RequestManager } from "./requests";
import { NodeInstance } from "./schema";

interface Vec2 {
    x: number;
    y: number;
}

interface CopiedNode {
    nodeType: string;
    offset: Vec2;
    parameter?: unknown;
}

interface CopiedConnection {
    fromNode: number;
    fromPort: string;
    toNode: number;
    toPort: string;
}

interface CopiedGraph {
    nodes: Array<CopiedNode>;
    connections: Array<CopiedConnection>;
}

export type AfterNodeCreated = (nodeID: string | null, flowNode: FlowNode) => void;

export class GraphClipboard {
    private clipboard: CopiedGraph | null = null;

    private selectOnceLoaded = new Set<string>();

    private pendingAfterCreate = new Map<FlowNode, AfterNodeCreated>();

    private lastNodeData = new Map<string, NodeInstance>();

    constructor(
        private readonly nodeFlowGraph: NodeFlowGraph,
        private readonly nodesPublisher: Publisher,
        private readonly requestManager: RequestManager,
        private readonly flowNodePathOf: (nodeType: string) => string | undefined,
    ) { }

    rememberNodeData(nodeID: string, data: NodeInstance): void {
        this.lastNodeData.set(nodeID, data);
    }

    forgetNode(nodeID: string): void {
        this.lastNodeData.delete(nodeID);
        this.selectOnceLoaded.delete(nodeID);
    }

    takePendingCreate(flowNode: FlowNode): AfterNodeCreated | undefined {
        const afterCreate = this.pendingAfterCreate.get(flowNode);
        this.pendingAfterCreate.delete(flowNode);
        return afterCreate;
    }

    claimPastedNode(nodeID: string): boolean {
        return this.selectOnceLoaded.delete(nodeID);
    }

    copy(nodes: Array<FlowNode>, connections: Array<InternalConnection>): void {
        const copied: Array<CopiedNode> = [];

        // Positions are kept relative to the top left of the selection, so a
        // paste lands as a group wherever it is dropped.
        let originX = Infinity;
        let originY = Infinity;
        for (const node of nodes) {
            originX = Math.min(originX, node.getPosition().x);
            originY = Math.min(originY, node.getPosition().y);
        }

        const copiedIndexOf = new Map<number, number>();

        nodes.forEach((node, index) => {
            const nodeType: string | undefined = node.metadata()?.typeData?.type;
            if (!nodeType) {
                return;
            }

            const instanceID = node.getProperty(InstanceIDProperty) as string;
            const data = instanceID ? this.lastNodeData.get(instanceID) : undefined;

            copiedIndexOf.set(index, copied.length);
            copied.push({
                nodeType,
                offset: {
                    x: node.getPosition().x - originX,
                    y: node.getPosition().y - originY,
                },
                parameter: data?.parameter?.currentValue,
            });
        });

        if (copied.length === 0) {
            this.clipboard = null;
            return;
        }

        const wires: Array<CopiedConnection> = [];
        for (const connection of connections) {
            const from = copiedIndexOf.get(connection.fromNode);
            const to = copiedIndexOf.get(connection.toNode);
            if (from === undefined || to === undefined) {
                continue;
            }
            wires.push({
                fromNode: from,
                fromPort: nodes[connection.fromNode].outputPort(connection.fromPort).getDisplayName(),
                toNode: to,
                toPort: nodes[connection.toNode].inputPort(connection.toPort).getDisplayName(),
            });
        }

        this.clipboard = { nodes: copied, connections: wires };
    }

    paste(position: Vec2): void {
        const clipboard = this.clipboard;
        if (clipboard === null) {
            return;
        }

        this.nodeFlowGraph.unselectAllNodes();

        const created = new Array<string | null>(clipboard.nodes.length).fill(null);
        let outstanding = clipboard.nodes.length;

        const finished = () => {
            outstanding--;
            if (outstanding !== 0) {
                return;
            }

            for (const wire of clipboard.connections) {
                const outID = created[wire.fromNode];
                const inID = created[wire.toNode];
                if (outID === null || inID === null) {
                    continue;
                }
                this.requestManager.setNodeInputConnection(inID, wire.toPort, outID, wire.fromPort);
            }
        };

        clipboard.nodes.forEach((copy, index) => {
            const at = {
                x: Math.round(position.x + copy.offset.x),
                y: Math.round(position.y + copy.offset.y),
            };

            const publisherPath = this.flowNodePathOf(copy.nodeType);

            if (!publisherPath) {
                this.requestManager.createNode(copy.nodeType, (resp) => {
                    created[index] = resp.nodeID;
                    this.selectOnceLoaded.add(resp.nodeID);
                    if (copy.parameter !== undefined) {
                        this.requestManager.setParameter(resp.nodeID, copy.parameter, false);
                    }
                    this.requestManager.setNodeMetadata(resp.nodeID, "position", at);
                    finished();
                }, undefined, finished);
                return;
            }

            const flowNode = this.nodesPublisher.create(publisherPath);
            flowNode.setPosition(at);

            this.pendingAfterCreate.set(flowNode, (nodeID, pasted) => {
                created[index] = nodeID;
                if (nodeID !== null) {
                    if (copy.parameter !== undefined) {
                        this.requestManager.setParameter(nodeID, copy.parameter, false);
                    }
                    pasted.select();
                }
                finished();
            });

            this.nodeFlowGraph.addNode(flowNode);
        });
    }
}
