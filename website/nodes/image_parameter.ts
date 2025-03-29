import { NodeManager } from "../node_manager";
import { FlowNode, ImageWidget } from "@elicdavis/node-flow";
import { RequestManager } from "../requests";

export class ImageParameterNodeController {

    id: string;

    flowNode: FlowNode;

    requestManager: RequestManager;

    nodeManager: NodeManager;

    constructor(lightNode: FlowNode, nodeManager: NodeManager, requestManager: RequestManager, id: string, parameterData) {
        this.flowNode = lightNode;
        this.id = id;
        this.requestManager = requestManager;
        this.flowNode.setTitle(parameterData.name);

        this.flowNode.addFileDropListener((file) => {
            const reader = new FileReader();
            reader.onload = (evt) => {
                console.log(evt.target.result)
                nodeManager.nodeParameterChanged({
                    id: id,
                    data: evt.target.result,
                    binary: true
                });
            }
            reader.readAsArrayBuffer(file);
            (this.flowNode.getWidget(0) as ImageWidget).SetBlob(file);
        });
    }

    update(parameterData) {
        this.requestManager.getParameterValue(this.id, (response) => {
            (this.flowNode.getWidget(0) as ImageWidget).SetBlob(response);
        });
    }
}