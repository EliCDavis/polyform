class NodeBasicParameter {
    constructor(nodeManager, id, parameterData, guiFolder, guiFolderData) {
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
    }

    update(parameterData) {
        this.guiFolderData[this.id] = parameterData.currentValue;
    }
}

function BuildParameter(nodeManager, id, parameterData, guiFolder, guiFolderData) {
    switch (parameterData.type) {
        case "float64":
        case "float32":
        case "int":
        case "bool":
        case "coloring.WebColor":
            return new NodeBasicParameter(nodeManager, id, parameterData, guiFolder, guiFolderData);

        default:
            throw new Error("unimplemented type: " + parameterData.type)
    }
}

class PolyNode {
    constructor(nodeManager, id, nodeData, guiFolder, guiFolderData) {
        this.guiFolder = guiFolder;
        this.guiFolderData = guiFolderData;
        this.nodeManager = nodeManager;

        this.id = id;
        this.name = "";
        this.outputs = [];
        this.version = 0;
        this.dependencies = [];
        this.parameter = null;

        this.update(nodeData);
    }

    update(nodeData) {
        this.name = nodeData.name;
        this.outputs = nodeData.outputs;
        this.version = nodeData.version;
        this.dependencies = nodeData.dependencies;

        if (nodeData.parameter) {
            if (!this.parameter) {
                this.parameter = BuildParameter(this.nodeManager, this.id, nodeData.parameter, this.guiFolder, this.guiFolderData);
            } else {
                this.parameter.update(nodeData.parameter)
            }
        }
    }
}