
export class Vector2ParameterNodeController {
    constructor(lightNode, nodeManager, id, parameterData, app) {
        this.nodeManager = nodeManager;
        this.id = id;
        this.updating = false;

        this.lightNode = lightNode;
        this.lightNode.setTitle(parameterData.name);

        const curVal = parameterData.currentValue;
        this.lightNode.setProperty("x", curVal.x);
        this.lightNode.setProperty("y", curVal.y);

        this.lightNode.addPropertyChangeListener("x", this.propertyChange.bind(this));
        this.lightNode.addPropertyChangeListener("y", this.propertyChange.bind(this));
       
    }

    propertyChange() {
        if (this.updating) {
            return
        }
        this.nodeManager.nodeParameterChanged({
            id: this.id,
            data: {
                x: this.lightNode.getProperty("x"),
                y: this.lightNode.getProperty("y"),
            },
            binary: false
        });
    }

    update(parameterData) {
        this.updating = true;
        const curVal = parameterData.currentValue;
        this.lightNode.setProperty("x", curVal.x);
        this.lightNode.setProperty("y", curVal.y);
        this.updating = false;
    }
}     