

export class NodeBasicParameter {
    constructor(app, nodeManager, id, parameterData, guiFolder, guiFolderData) {
        this.id = id;
        this.guiFolder = guiFolder;
        this.guiFolderData = guiFolderData;

        guiFolderData[id] = parameterData.currentValue;

        let setting = null;

        if (parameterData.type === "coloring.WebColor") {
            setting = guiFolder.addColor(guiFolderData, id)
        } else {
            setting = guiFolder.add(guiFolderData, id)
        }

        setting = setting.name(parameterData.name);

        if (parameterData.type === "int") {
            setting = setting.step(1)
        }

        this.setting = setting
            .listen()
            .onChange((newData) => {
                nodeManager.nodeParameterChanged({
                    id: id,
                    data: newData
                });
            });

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
        console.log(this.lightNode)
        this.lightNode.outputs[0].type = parameterData.type;
        this.lightNode.setValue(parameterData.currentValue);

        this.lightNode.onPropertyChanged = (property, value) => {
            console.log(property, value);
            if (property !== "value") {
                return;
            }
            nodeManager.nodeParameterChanged({ id: id, data: value });
        }
    }

    update(parameterData) {
        this.guiFolderData[this.id] = parameterData.currentValue;
    }
}
