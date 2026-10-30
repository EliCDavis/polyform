import { NodeManager } from "../node_manager";
import { FlowNode } from "@elicdavis/node-flow";
import type { Vector4NodeParameter } from "../../types/parameter";

const components = ["x", "y", "z", "w"] as const;

export class Vector4ParameterNodeController {

    id: string;

    nodeManager: NodeManager;

    flowNode: FlowNode;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: Vector4NodeParameter) {
        this.nodeManager = nodeManager;
        this.id = id;

        this.flowNode = flowNode;
        this.flowNode.setTitle(parameterData.name);

        this.setProperties(parameterData);
        for (const c of components) {
            this.flowNode.addPropertyEditListener(c, this.propertyChange.bind(this));
        }
    }

    private setProperties(parameterData: Vector4NodeParameter): void {
        const curVal = parameterData.currentValue;
        for (const c of components) {
            this.flowNode.setProperty(c, curVal?.[c] ?? 0);
        }
    }

    propertyChange() {
        this.nodeManager.nodeParameterChanged({
            id: this.id,
            data: {
                x: this.flowNode.getProperty("x"),
                y: this.flowNode.getProperty("y"),
                z: this.flowNode.getProperty("z"),
                w: this.flowNode.getProperty("w"),
            },
            binary: false
        });
    }

    update(parameterData: Vector4NodeParameter) {
        this.setProperties(parameterData);
    }
}
