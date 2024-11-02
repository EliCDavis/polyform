
export class ImageParameterNodeController {

    constructor(lightNode, nodeManager, id, parameterData, app) {
        this.lightNode = lightNode;
        this.id = id;
        this.app = app;
        this.lightNode.setTitle(parameterData.name);

        this.lightNode.addFileDropListener((file) => {
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
            this.lightNode.getWidget(0).SetBlob(file);
        });
    }

    update(parameterData) {
        this.app.RequestManager.getParameterValue(this.id, (response) => {
            this.lightNode.getWidget(0).SetBlob(response);
        });
    }
}