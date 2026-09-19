import { NodeManager } from "../node_manager";
import { FlowNode } from "@elicdavis/node-flow";
import type { Vector4NodeParameter } from "../../types/parameter";

const components = ["x", "y", "z", "w"] as const;

export class Vector4ParameterNodeController {

    id: string;

    nodeManager: NodeManager;

    updating: boolean;

    flowNode: FlowNode;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: Vector4NodeParameter) {
        this.nodeManager = nodeManager;
        this.id = id;
        this.updating = false;

        this.flowNode = flowNode;
        this.flowNode.setTitle(parameterData.name);

        this.setProperties(parameterData);
        for (const c of components) {
            this.flowNode.addPropertyChangeListener(c, this.propertyChange.bind(this));
        }
    }

    private setProperties(parameterData: Vector4NodeParameter): void {
        const curVal = parameterData.currentValue;
        for (const c of components) {
            this.flowNode.setProperty(c, curVal?.[c] ?? 0);
        }
    }

    propertyChange() {
        if (this.updating) {
            return
        }
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
        this.updating = true;
        this.setProperties(parameterData);
        this.updating = false;
    }
}
