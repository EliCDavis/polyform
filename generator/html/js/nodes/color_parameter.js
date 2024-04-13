
function BuildColorParameterNode(app) {
    const node = LiteGraph.createNode("polyform/color");
    console.log(node)
    app.LightGraph.add(node);
    return node;
}

export class ColorParameter {

    constructor(nodeManager, id, parameterData, app) {
        this.lightNode = BuildColorParameterNode(app);
        this.lightNode.title = parameterData.name;

        this.lightNode.widgets[0].value = parameterData.currentValue;
        this.lightNode.widgets[0].mouse = (event, pos, node) => {
            if (event.type !== "mouseup") {
                return;
            }
            app.ColorSelector.show(this.lightNode.widgets[0].value, (newColor) => {
                this.lightNode.widgets[0].value = newColor;
                app.LightGraph.setDirtyCanvas(true);
                nodeManager.nodeParameterChanged({
                    id: id,
                    data: newColor
                });
            })
            console.log("clicked!", event)
        }
    }

    update(parameterData) {
        this.lightNode.widgets[0].value = parameterData.currentValue;
    }

}     