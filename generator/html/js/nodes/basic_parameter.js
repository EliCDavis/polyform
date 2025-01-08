export class BasicParameterNodeController {
    constructor(lightNode, nodeManager, id, parameterData) {
        this.id = id;
        this.lightNode = lightNode;
        this.lightNode.setTitle(parameterData.name);
        this.lightNode.setProperty("value", parameterData.currentValue);
        this.updating = false;

        this.lightNode.addPropertyChangeListener("value", (oldVal, newVal) => {
            if (this.updating) {
                return;
            }
            nodeManager.nodeParameterChanged({ id: id, data: newVal });
        });
    }

    update(parameterData) {
        this.updating = true;
        this.lightNode.setProperty("value", parameterData.currentValue);
        this.updating = false;
    }
}
