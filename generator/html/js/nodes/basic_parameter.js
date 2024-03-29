

export class NodeBasicParameter {
    constructor(app, nodeManager, id, parameterData) {
        this.id = id;


        let setting = null;

        // https://github.com/jagenjo/litegraph.js/tree/master/guides
        switch (parameterData.type) {
            case "float64":
                this.lightNode = LiteGraph.createNode("basic/const");
                break;
            case "float32":
                this.lightNode = LiteGraph.createNode("basic/const");
                break;
            case "int":
                this.lightNode = LiteGraph.createNode("basic/const");
                break;
            case "bool":
                this.lightNode = LiteGraph.createNode("basic/boolean");
                break;
            case "string":
                this.lightNode = LiteGraph.createNode("basic/string");
                break;
            default:
                console.log("unimplemented", parameterData.type)
            // case "coloring.WebColor":
            //     this.lightNode = LiteGraph.createNode("basic/const");
            //     break;
        }
        // this.lightNode.pos = [200, app.LightGraph._nodes.length * 100];
        this.lightNode.title = parameterData.name;
        app.LightGraph.add(this.lightNode);
        this.lightNode.outputs[0].type = parameterData.type;
        this.lightNode.setValue(parameterData.currentValue);
        this.lightNode.setSize(this.lightNode.computeSize());


        this.lightNode.onPropertyChanged = (property, value) => {
            if (property !== "value") {
                return;
            }
            nodeManager.nodeParameterChanged({ id: id, data: value });
        }
    }

    update(parameterData) {
        // this.guiFolderData[this.id] = parameterData.currentValue;
    }
}
