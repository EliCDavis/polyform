import { NodeManager } from "../node_manager";
import { FlowNode } from "@elicdavis/node-flow";

export class FileParameterNodeController {

    flowNode: FlowNode;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData) {
        this.flowNode = flowNode;
        this.flowNode.setTitle(parameterData.name);

        this.flowNode.addFileDropListener((file) => {
            var reader = new FileReader();
            reader.onload = (evt) => {
                console.log(evt.target.result)
                nodeManager.nodeParameterChanged({
                    id: id,
                    data: evt.target.result,
                    binary: true
                });
            }
            reader.readAsArrayBuffer(file);
        });
    }

    update(parameterData) {
        console.log("file parameter", parameterData)
    }

}