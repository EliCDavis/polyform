import { NodeManager } from "../node_manager";
import { FlowNode } from "@elicdavis/node-flow";
import type { FileNodeParameter } from "../../types/parameter";

export class FileParameterNodeController {

    flowNode: FlowNode;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: FileNodeParameter) {
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

    update(parameterData: FileNodeParameter) {
        console.log("file parameter", parameterData)
    }

}