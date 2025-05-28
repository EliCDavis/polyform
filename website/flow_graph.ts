import { FlowNodeConfig, NodeFlowGraph, Publisher } from "@elicdavis/node-flow";

const nodeCanvas = document.getElementById("light-canvas")


const Green = {
    Background: "#233",
    Title: "#355"
}

const ParameterNodeBackgroundColor = "#333";
const ParameterOutPortName = "Value";
const ParameterStyle = {
    title: {
        color: "#545454"
    },
    idle: {
        color: ParameterNodeBackgroundColor,
    },
    mouseOver: {
        color: ParameterNodeBackgroundColor,
    },
    grabbed: {
        color: ParameterNodeBackgroundColor,
    },
    selected: {
        color: ParameterNodeBackgroundColor,
    }
}

const IntParameter: FlowNodeConfig = {
    title: "Int Parameter",
    subTitle: "Int",
    canEditTitle: true,
    canEditInfo: true,
    outputs: [
        {
            name: ParameterOutPortName,
            type: "int"
        }
    ],
    widgets: [
        {
            type: "number",
            config: {
                property: "value"
            }
        }
    ],
    style: ParameterStyle,
    metadata: {
        typeData: {
            type: "github.com/EliCDavis/polyform/generator/parameter.Value[int]"
        }
    }
}

const FloatParameter: FlowNodeConfig = {
    title: "Float64 Parameter",
    subTitle: "Float64",
    canEditTitle: true,
    canEditInfo: true,
    outputs: [
        {
            name: ParameterOutPortName,
            type: "float64"
        }
    ],
    widgets: [
        {
            type: "number",
            config: {
                property: "value"
            }
        }
    ],
    style: ParameterStyle,
    metadata: {
        typeData: {
            type: "github.com/EliCDavis/polyform/generator/parameter.Value[float64]"
        }
    }
};

const AABBParameter: FlowNodeConfig = {
    title: "AABB Parameter",
    subTitle: "AABB",
    canEditTitle: true,
    canEditInfo: true,
    outputs: [
        {
            name: ParameterOutPortName,
            type: "github.com/EliCDavis/polyform/math/geometry.AABB"
        }
    ],
    widgets: [
        {
            type: "text",
            config: {
                value: "min"
            }
        },
        {
            type: "number",
            config: {
                property: "min-x"
            }
        },
        {
            type: "number",
            config: {
                property: "min-y"
            }
        },
        {
            type: "number",
            config: {
                property: "min-z"
            }
        },
        {
            type: "text",
            config: {
                value: "max"
            }
        },
        {
            type: "number",
            config: {
                property: "max-x"
            }
        },
        {
            type: "number",
            config: {
                property: "max-y"
            }
        },
        {
            type: "number",
            config: {
                property: "max-z"
            }
        }
    ],
    style: ParameterStyle,
    metadata: {
        typeData: {
            type: "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/polyform/math/geometry.AABB]"
        }
    }
};

const ImageParameter: FlowNodeConfig = {
    title: "Image Parameter",
    subTitle: "Image",
    canEditTitle: true,
    canEditInfo: true,
    outputs: [
        {
            name: ParameterOutPortName,
            type: "image.Image"
        }
    ],
    widgets: [
        {
            type: "image",
            config: {
            }
        }
    ],
    style: ParameterStyle,
    metadata: {
        typeData: {
            type: "github.com/EliCDavis/polyform/generator/parameter.Image"
        }
    }
};

const FileParameter: FlowNodeConfig = {
    title: "File Parameter",
    subTitle: "File",
    canEditTitle: true,
    canEditInfo: true,
    outputs: [
        {
            name: ParameterOutPortName,
            type: "[]uint8"
        }
    ],
    style: ParameterStyle,
    metadata: {
        typeData: {
            type: "github.com/EliCDavis/polyform/generator/parameter.File"
        }
    }
};

const ColorParamter: FlowNodeConfig = {
    title: "Color Parameter",
    subTitle: "Color",
    canEditTitle: true,
    canEditInfo: true,
    outputs: [
        {
            name: ParameterOutPortName,
            type: "github.com/EliCDavis/polyform/drawing/coloring.WebColor"
        }
    ],
    widgets: [
        {
            type: "color",
            config: {
                property: "value"
            }
        }
    ],
    style: ParameterStyle,
    metadata: {
        typeData: {
            type: "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/polyform/drawing/coloring.WebColor]"
        }
    }
};

const Vector3Parameter: FlowNodeConfig = {
    title: "Vector3 Parameter",
    subTitle: "Vector3",
    canEditTitle: true,
    canEditInfo: true,
    outputs: [
        {
            name: ParameterOutPortName,
            type: "github.com/EliCDavis/vector/vector3.Vector[float64]"
        }
    ],
    widgets: [
        {
            type: "number",
            config: {
                property: "x"
            }
        },
        {
            type: "number",
            config: {
                property: "y"
            }
        },
        {
            type: "number",
            config: {
                property: "z"
            }
        }
    ],
    style: ParameterStyle,
    metadata: {
        typeData: {
            type: "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/vector/vector3.Vector[float64]]"
        }
    }
};

const Vector3ArrayParameter: FlowNodeConfig = {
    title: "Vector3 Array Parameter",
    subTitle: "Vector3 Array",
    canEditTitle: true,
    canEditInfo: true,
    outputs: [
        {
            name: ParameterOutPortName,
            type: "[]github.com/EliCDavis/vector/vector3.Vector[float64]"
        }
    ],
    style: ParameterStyle,
    metadata: {
        typeData: {
            type: "github.com/EliCDavis/polyform/generator/parameter.Value[[]github.com/EliCDavis/vector/vector3.Vector[float64]]"
        }
    }
};

const Vector2Parameter: FlowNodeConfig = {
    title: "Vector2 Parameter",
    subTitle: "Vector2",
    canEditTitle: true,
    canEditInfo: true,
    outputs: [
        {
            name: ParameterOutPortName,
            type: "github.com/EliCDavis/vector/vector2.Vector[float64]"
        }
    ],
    widgets: [
        {
            type: "number",
            config: {
                property: "x"
            }
        },
        {
            type: "number",
            config: {
                property: "y"
            }
        }
    ],
    style: ParameterStyle,
    metadata: {
        typeData: {
            type: "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/vector/vector2.Vector[float64]]"
        }
    }
};

const BoolParameters: FlowNodeConfig = {
    title: "Bool Parameter",
    subTitle: "Bool",
    canEditTitle: true,
    canEditInfo: true,
    outputs: [
        {
            name: ParameterOutPortName,
            type: "bool"
        }
    ],
    widgets: [
        {
            type: "toggle",
            config: {
                property: "value"
            }
        },
    ],
    style: ParameterStyle,
    metadata: {
        typeData: {
            type: "github.com/EliCDavis/polyform/generator/parameter.Value[bool]"
        }
    }
};
const StringParameter: FlowNodeConfig = {
    title: "String Parameter",
    subTitle: "String",
    canEditTitle: true,
    canEditInfo: true,
    outputs: [
        {
            name: ParameterOutPortName,
            type: "string"
        }
    ],
    widgets: [
        {
            type: "string",
            config: {
                property: "value"
            }
        },
    ],
    style: ParameterStyle,
    metadata: {
        typeData: {
            type: "github.com/EliCDavis/polyform/generator/parameter.Value[string]"
        }
    }
}

interface FlowGraphInit {
    PolyformNodesPublisher: Publisher,
    NodeFlowGraph: NodeFlowGraph
}

export function CreateNodeFlowGraph(): FlowGraphInit {
    const publisher = new Publisher({
        name: "Polyform",
        version: "1.0.0",
        nodes: {
            "Parameters/bool": BoolParameters,
            "Parameters/int": IntParameter,
            "Parameters/float64": FloatParameter,
            "Parameters/coloring.WebColor": ColorParamter,
            "Parameters/string": StringParameter,
            "Parameters/geometry.AABB": AABBParameter,
            "Parameters/image.Image": ImageParameter,
            "Parameters/File": FileParameter,
            "Parameters/vector3.Vector[float64]": Vector3Parameter,
            "Parameters/[]vector3.Vector[float64]": Vector3ArrayParameter,
            "Parameters/vector2.Vector[float64]": Vector2Parameter,
        },
    });

    const nodeFlowGraph = new NodeFlowGraph(nodeCanvas as HTMLCanvasElement, {});
    nodeFlowGraph.addPublisher("polyform", publisher);
    return {
        NodeFlowGraph: nodeFlowGraph,
        PolyformNodesPublisher: publisher
    }
}



