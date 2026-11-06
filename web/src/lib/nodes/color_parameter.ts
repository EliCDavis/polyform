import { FlowNode } from "@elicdavis/node-flow";
import { NodeManager } from "../node_manager";
import { opaque, shownAlpha, withAlpha } from "../color_alpha";
import type { ScalarNodeParameter } from "../../types/parameter";

export class ColorParameterNodeController {

    flowNode: FlowNode;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: ScalarNodeParameter) {
        this.flowNode = flowNode;
        this.flowNode.setTitle(parameterData.name);
        this.update(parameterData);

        const edited = () => nodeManager.nodeParameterChanged({
            id: id,
            data: withAlpha(this.flowNode.getProperty("value"), this.flowNode.getProperty("alpha")),
            binary: false
        });
        this.flowNode.addPropertyEditListener("value", edited);
        this.flowNode.addPropertyEditListener("alpha", edited);
    }

    update(parameterData: ScalarNodeParameter) {
        const hex = String(parameterData.currentValue);
        this.flowNode.setProperty("value", opaque(hex));
        this.flowNode.setProperty("alpha", shownAlpha(hex));
    }
}
