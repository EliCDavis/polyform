import { NodeManager } from "../node_manager";
import { FlowNode } from "@elicdavis/node-flow";
import type { ScalarNodeParameter } from "../../types/parameter";

export class BasicParameterNodeController {

    flowNode: FlowNode;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: ScalarNodeParameter) {
        this.flowNode = flowNode;
        this.flowNode.setTitle(parameterData.name);
        this.flowNode.setProperty("value", parameterData.currentValue);

        this.flowNode.addPropertyEditListener("value", (value) => {
            nodeManager.nodeParameterChanged({
                id: id,
                data: value,
                binary: false
            });
        });
    }

    update(parameterData: ScalarNodeParameter) {
        this.flowNode.setProperty("value", parameterData.currentValue);
    }
}
