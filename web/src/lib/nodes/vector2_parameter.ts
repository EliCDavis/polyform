import { NodeManager } from "../node_manager";
import { FlowNode } from "@elicdavis/node-flow";
import type { Vector2NodeParameter } from "../../types/parameter";

export class Vector2ParameterNodeController {

    id: string;

    nodeManager: NodeManager;

    flowNode: FlowNode;

    constructor(lightNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: Vector2NodeParameter) {
        this.nodeManager = nodeManager;
        this.id = id;

        this.flowNode = lightNode;
        this.flowNode.setTitle(parameterData.name);

        const curVal = parameterData.currentValue;
        this.flowNode.setProperty("x", curVal.x);
        this.flowNode.setProperty("y", curVal.y);

        this.flowNode.addPropertyEditListener("x", this.propertyChange.bind(this));
        this.flowNode.addPropertyEditListener("y", this.propertyChange.bind(this));
    }

    propertyChange() {
        this.nodeManager.nodeParameterChanged({
            id: this.id,
            data: {
                x: this.flowNode.getProperty("x"),
                y: this.flowNode.getProperty("y"),
            },
            binary: false
        });
    }

    update(parameterData: Vector2NodeParameter) {
        const curVal = parameterData.currentValue;
        this.flowNode.setProperty("x", curVal.x);
        this.flowNode.setProperty("y", curVal.y);
    }
}
