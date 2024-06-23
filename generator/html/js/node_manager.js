import { PolyNode, camelCaseToWords } from "./nodes/node.js";


const ParameterNodeOutputPortName = "Out";
const ParameterNodeColor = "#233";
const ParameterNodeBackgroundColor = "#355";

/**
 * 
 * @param {string} dataType 
 * @returns {string}
 */
function ParameterNodeType(dataType) {
    return "github.com/EliCDavis/polyform/parameter.Value[" + dataType + "]";
}


/**
 * 
 * @param {string} subCategory 
 * @returns {string}
 */
function ParameterNamespace(subCategory) {
    return "parameter/" + subCategory
}


export class NodeManager {
    constructor(app) {
        this.app = app;
        this.guiFolderData = {};
        this.nodeIdToNode = new Map();
        this.nodeTypeToLitePath = new Map();
        this.subscribers = [];
        this.initializedNodeTypes = false
        this.registerSpecialParameterPolyformNodes();
    }

    onNodeCreateCallback(liteNode, nodeType) {
        if (this.app.ServerUpdatingNodeConnections) {
            return;
        }
        liteNode.setSize(liteNode.computeSize());
        this.app.RequestManager.createNode(nodeType, (resp) => {
            const isProducer = false;
            const nodeID = resp.nodeID
            const nodeData = resp.data;
            liteNode.nodeInstanceID = nodeID;
            this.nodeIdToNode.set(nodeID, new PolyNode(liteNode, this, nodeID, nodeData, this.app, isProducer));
        })
    }


    registerSpecialParameterPolyformNodes() {
        const nm = this;
        function ImageParameterNode() {
            const nodeDataType = "image.Image";
            this.addOutput(ParameterNodeOutputPortName, nodeDataType);
            this.title = "Image";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;

            const H = LiteGraph.NODE_WIDGET_HEIGHT;

            const imgWidget = this.addWidget("image", "Image", true, { property: "surname" }); //this will modify the node.properties
            this.imgWidget = imgWidget;
            const margin = 15;
            this.imgWidget.draw = (ctx, node, widget_width, y, H) => {
                if (!imgWidget.image) {
                    return;
                }

                const adjustedWidth = widget_width - margin * 2
                ctx.drawImage(
                    imgWidget.image,
                    margin,
                    y,
                    adjustedWidth,
                    (adjustedWidth / imgWidget.image.width) * imgWidget.image.height
                );
            }

            this.imgWidget.computeSize = (width) => {
                if (!!imgWidget.image) {
                    const adjustedWidth = width - margin * 2
                    const newH = (adjustedWidth / imgWidget.image.width) * imgWidget.image.height;
                    return [width, newH]
                }
                return [width, 0];
            }

            nm.onNodeCreateCallback(this, "github.com/EliCDavis/polyform/generator.ImageParameterNode");
        }

        function ColorParameterNode() {
            const nodeDataType = "github.com/EliCDavis/polyform/drawing/coloring.WebColor";
            this.addOutput(ParameterNodeOutputPortName, nodeDataType);
            this.title = "Color";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;


            const imgWidget = this.addWidget("color", "Color", "#00FF00", {}); //this will modify the node.properties
            this.imgWidget = imgWidget;
            const margin = 15;
            this.imgWidget.draw = (ctx, node, widget_width, y, H) => {
                const adjustedWidth = widget_width - margin * 2
                ctx.beginPath(); // Start a new path
                ctx.rect(margin, y, adjustedWidth, H); // Add a rectangle to the current path
                ctx.fillStyle = this.imgWidget.value;
                ctx.fill(); // Render the path
            }

            // this.imgWidget.mouse = (event, pos, node) => {
            //     if (event.type == LiteGraph.pointerevents_method + "down") {
            //         w.value = !w.value;
            //         setTimeout(function () {
            //             inner_value_change(w, w.value);
            //         }, 20);
            //     }
            // }
            nm.onNodeCreateCallback(this, ParameterNodeType(nodeDataType));
        }

        function FileParameterNode() {
            const nodeDataType = "[]uint8";
            // const nodeDataType = "github.com/EliCDavis/polyform/generator.FileParameterNode";
            this.addOutput(ParameterNodeOutputPortName, nodeDataType);
            this.title = "File";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;
            // nm.onNodeCreateCallback(this, ParameterNodeType(nodeDataType));
            nm.onNodeCreateCallback(this, "github.com/EliCDavis/polyform/generator.FileParameterNode");
        }

        function Vector3ParameterNode() {
            const nodeDataType = "github.com/EliCDavis/vector/vector3.Vector[float64]";
            this.addOutput(ParameterNodeOutputPortName, nodeDataType);
            this.title = "Vector3";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;
            nm.onNodeCreateCallback(this, ParameterNodeType(nodeDataType));
        }

        function Vector3ArrayParameterNode() {
            const nodeDataType = "[]github.com/EliCDavis/vector/vector3.Vector[float64]";
            this.addOutput(ParameterNodeOutputPortName, nodeDataType);
            this.title = "Vector3 Array";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;
            nm.onNodeCreateCallback(this, ParameterNodeType(nodeDataType));
        }


        function AABBParameterNode() {
            const nodeDataType = "github.com/EliCDavis/polyform/math/geometry.AABB";
            this.addOutput(ParameterNodeOutputPortName, nodeDataType);
            this.title = "AABB";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;
            nm.onNodeCreateCallback(this, ParameterNodeType(nodeDataType));
        }

        function Float64Parameter() {
            this.addOutput("value", "float64");
            this.addProperty("value", 1.0);
            this.widget = this.addWidget("number", "value", 1, "value");
            this.widgets_up = true;
            this.size = [180, 30];
            this.title = "Const Float64";
            this.desc = "Constant Float64";
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;

            this.onExecute = () => {
                this.setOutputData(0, parseFloat(this.properties["value"]));
            };

            this.getTitle = () => {
                if (this.flags.collapsed) {
                    return this.properties.value;
                }
                return this.title;
            };

            this.setValue = (v) => {
                this.setProperty("value", v);
            }

            this.onDrawBackground = function (ctx) {
                this.outputs[0].label = this.properties["value"].toFixed(3);
            };

            nm.onNodeCreateCallback(this, ParameterNodeType("float64"));
        }

        function StringParameter() {
            this.addOutput("string", "string");
            this.addProperty("value", "");
            this.widget = this.addWidget("text", "value", "", "value");  //link to property value
            this.widgets_up = true;
            this.size = [180, 30];
            this.color = ParameterNodeColor;
            this.bgcolor = ParameterNodeBackgroundColor;

            console.log("bitch im alive!!!")

            this.title = "Const String";
            this.desc = "Constant string";


            this.getTitle = () => {
                if (this.flags.collapsed) {
                    return this.properties.value;
                }
                return this.title;
            };

            this.onExecute = function () {
                this.setOutputData(0, this.properties["value"]);
            };

            this.setValue = (v) => {
                this.setProperty("value", v);
            }

            this.onDropFile = function (file) {
                var that = this;
                var reader = new FileReader();
                reader.onload = function (e) {
                    that.setProperty("value", e.target.result);
                }
                reader.readAsText(file);
            }

            nm.onNodeCreateCallback(this, ParameterNodeType("string"));
        }


        LiteGraph.registerNodeType(ParameterNamespace("string"), StringParameter);
        LiteGraph.registerNodeType(ParameterNamespace("float64"), Float64Parameter);
        LiteGraph.registerNodeType(ParameterNamespace("aabb"), AABBParameterNode);
        LiteGraph.registerNodeType(ParameterNamespace("vector3"), Vector3ParameterNode);
        LiteGraph.registerNodeType(ParameterNamespace("vector3[]"), Vector3ArrayParameterNode);
        LiteGraph.registerNodeType(ParameterNamespace("color"), ColorParameterNode);
        LiteGraph.registerNodeType(ParameterNamespace("image"), ImageParameterNode);
        LiteGraph.registerNodeType(ParameterNamespace("file"), FileParameterNode);
    }

    sortNodesByName(nodesToSort) {
        let sortable = [];
        for (var nodeID in nodesToSort) {
            sortable.push([nodeID, nodesToSort[nodeID]]);
        }

        sortable.sort((a, b) => {
            var textA = a[1].name.toUpperCase();
            var textB = b[1].name.toUpperCase();
            return (textA < textB) ? -1 : (textA > textB) ? 1 : 0;
        });
        return sortable;
    }

    updateNodeConnections(nodes) {
        for (let node of nodes) {
            const nodeID = node[0];
            const nodeData = node[1];
            const source = this.nodeIdToNode.get(nodeID);

            for (let i = 0; i < nodeData.dependencies.length; i++) {
                const dep = nodeData.dependencies[i];
                const target = this.nodeIdToNode.get(dep.dependencyID);

                let sourceInput = -1;
                for (let sourceInputIndex = 0; sourceInputIndex < source.liteNode.inputs.length; sourceInputIndex++) {
                    if (source.liteNode.inputs[sourceInputIndex].name === dep.name) {
                        sourceInput = sourceInputIndex;
                    }
                }

                // TODO: This only works for nodes with one output
                target.liteNode.connect(0, source.liteNode, sourceInput)
                // source.lightNode.connect(i, target.lightNode, 0);
            }
        }
    }

    buildCustomNodeType(typeData) {
        const nm = this;
        function CustomNodeFunc() {
            for (var inputName in typeData.inputs) {
                this.addInput(inputName, typeData.inputs[inputName].type);
            }

            if (typeData.outputs) {
                typeData.outputs.forEach((o) => {
                    this.addOutput(o.name, o.type);
                })
            }

            // if (producers.includes(nodeData.name)) {
            //     this.color = "#232";
            //     this.bgcolor = "#353";
            //     this.addWidget("button", "Download", null, () => {
            //         console.log("presed");
            //         saveFileToDisk("/producer/" + typeData.displayName, typeData.displayName);
            //     })
            // }
            this.title = camelCaseToWords(typeData.displayName);

            // this.onNodeCreated = () => {
            //     if (nm.app.ServerUpdatingNodeConnections) {
            //         return;
            //     }
            //     nm.app.RequestManager.createNode(typeData.type)
            //     console.log("node created: ", typeData.type)
            // }
            nm.onNodeCreateCallback(this, typeData.type);
        }

        Object.defineProperty(CustomNodeFunc, "name", { value: typeData.displayName });

        const category = typeData.path + "/" + typeData.displayName;
        LiteGraph.registerNodeType(category, CustomNodeFunc);
        this.nodeTypeToLitePath.set(typeData.type, category);
    }

    newLiteNode(nodeData) {
        const isParameter = !!nodeData.parameter;

        // Not a parameter, just create a node that adhere's to the server's 
        // reflection.
        if (!isParameter) {
            const nodeIdentifier = this.nodeTypeToLitePath.get(nodeData.type)
            return LiteGraph.createNode(nodeIdentifier);
        }

        const parameterType = nodeData.parameter.type;
        switch (parameterType) {
            case "float64":
                return LiteGraph.createNode(ParameterNamespace("float64"));

            case "string":
                return LiteGraph.createNode(ParameterNamespace("string"));

            case "coloring.WebColor":
                return LiteGraph.createNode(ParameterNamespace("color"));

            case "vector3.Vector[float64]":
            case "vector3.Vector[float32]":
                return LiteGraph.createNode(ParameterNamespace("vector3"));

            case "[]vector3.Vector[float64]":
            case "[]vector3.Vector[float32]":
                return LiteGraph.createNode(ParameterNamespace("vector3[]"));

            case "geometry.AABB":
                return LiteGraph.createNode(ParameterNamespace("aabb"));

            case "image.Image":
                return LiteGraph.createNode(ParameterNamespace("image"));

            case "[]uint8":
                return LiteGraph.createNode(ParameterNamespace("file"));

            default:
                throw new Error("unimplemented parameter type: " + parameterType);

        }
    }

    updateNodes(newSchema) {
        // Only need to do this once, since types are set at compile time. If
        // that ever changes, god.
        if (this.initializedNodeTypes === false) {
            this.initializedNodeTypes = true;
            newSchema.types.forEach(type => {
                // We should have custom nodes already defined for parameters
                if (type.parameter) {
                    return;
                }
                this.buildCustomNodeType(type)
            })
        }

        const sortedNodes = this.sortNodesByName(newSchema.nodes);

        this.app.ServerUpdatingNodeConnections = true;

        let nodeAdded = false;
        for (let node of sortedNodes) {
            const nodeID = node[0];
            const nodeData = node[1];

            if (this.nodeIdToNode.has(nodeID)) {
                const nodeToUpdate = this.nodeIdToNode.get(nodeID);
                nodeToUpdate.update(nodeData);
            } else {
                let isProducer = false;
                for (const [key, value] of Object.entries(newSchema.producers)) {
                    if (value.nodeID === nodeID) {
                        isProducer = true;
                    }
                }

                const liteNode = this.newLiteNode(nodeData);
                this.app.LightGraph.add(liteNode);
                liteNode.setSize(liteNode.computeSize());
                liteNode.nodeInstanceID = nodeID;

                this.nodeIdToNode.set(nodeID, new PolyNode(liteNode, this, nodeID, nodeData, this.app, isProducer));
                nodeAdded = true;
            }
        }

        this.updateNodeConnections(sortedNodes);

        if (nodeAdded) {
            lgraphInstance.arrange();
        }
        this.app.ServerUpdatingNodeConnections = false;
    }

    subscribeToParameterChange(subscriber) {
        this.subscribers.push(subscriber)
    }

    nodeParameterChanged(change) {
        this.subscribers.forEach((e) => e(change))
    }
}