class NodeBasicParameter {
    constructor(nodeManager, id, parameterData, guiFolder, guiFolderData) {
        guiFolderData[id] = parameterData.currentValue;

        console.log(guiFolderData)

        this.setting = guiFolder.add(guiFolderData, id)
            .name(parameterData.name)
            .listen()
            .onChange((newData) => {
                nodeManager.nodeParameterChanged({
                    id: id,
                    data: newData
                });
            });
    }

    update(parameterData) {

    }
}

function BuildParameter(nodeManager, id, parameterData, guiFolder, guiFolderData) {
    switch (parameterData.type.toLowerCase()) {
        case "float64":
        case "float32":
        case "int":
        case "bool":
        case "color":
            return new NodeBasicParameter(nodeManager, id, parameterData, guiFolder, guiFolderData);

        default:
            throw new Error("unimplemented type: " + param.type)
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