import { NodeManager } from "../node_manager";
import { FlowNode } from "@elicdavis/node-flow";
import type { ScalarNodeParameter } from "../../types/parameter";

export class BasicParameterNodeController {

    flowNode: FlowNode;

    updating: boolean;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: ScalarNodeParameter) {
        this.flowNode = flowNode;
        this.flowNode.setTitle(parameterData.name);
        this.flowNode.setProperty("value", parameterData.currentValue);
        this.updating = false;

        this.flowNode.addPropertyChangeListener("value", (oldVal, newVal) => {
            if (this.updating) {
                return;
            }
            nodeManager.nodeParameterChanged({
                id: id,
                data: newVal,
                binary: false
            });
        });
    }

    update(parameterData: ScalarNodeParameter) {
        this.updating = true;
        this.flowNode.setProperty("value", parameterData.currentValue);
        this.updating = false;
    }
}
