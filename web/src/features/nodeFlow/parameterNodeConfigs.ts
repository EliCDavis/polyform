import type { FlowNodeConfig } from "@elicdavis/node-flow";

interface ColorScheme {
  Background: string;
  Title: string;
}

const ColorSchemes = {
  Green: {
    Background: "#233",
    Title: "#355",
  },
  Grey: {
    Background: "#333",
    Title: "#545454",
  },
};

const parameterPallete: ColorScheme = ColorSchemes.Grey;

const ParameterOutPortName = "Value";
const ParameterStyle = {
  title: { color: parameterPallete.Title },
  idle: { color: parameterPallete.Background },
  mouseOver: { color: parameterPallete.Background },
  grabbed: { color: parameterPallete.Background },
  selected: { color: parameterPallete.Background },
};

const IntParameter: FlowNodeConfig = {
  title: "Int Parameter",
  subTitle: "Int",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [{ name: ParameterOutPortName, type: "int" }],
  widgets: [{ type: "number", config: { property: "value" } }],
  style: ParameterStyle,
  metadata: {
    typeData: { type: "github.com/EliCDavis/polyform/generator/parameter.Value[int]" },
  },
};

const FloatParameter: FlowNodeConfig = {
  title: "Float64 Parameter",
  subTitle: "Float64",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [{ name: ParameterOutPortName, type: "float64" }],
  widgets: [{ type: "number", config: { property: "value" } }],
  style: ParameterStyle,
  metadata: {
    typeData: { type: "github.com/EliCDavis/polyform/generator/parameter.Value[float64]" },
  },
};

const AABBParameter: FlowNodeConfig = {
  title: "AABB Parameter",
  subTitle: "AABB",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [
    { name: ParameterOutPortName, type: "github.com/EliCDavis/polyform/math/geometry.AABB" },
  ],
  widgets: [
    { type: "text", config: { value: "min" } },
    { type: "number", config: { property: "min-x" } },
    { type: "number", config: { property: "min-y" } },
    { type: "number", config: { property: "min-z" } },
    { type: "text", config: { value: "max" } },
    { type: "number", config: { property: "max-x" } },
    { type: "number", config: { property: "max-y" } },
    { type: "number", config: { property: "max-z" } },
  ],
  style: ParameterStyle,
  metadata: {
    typeData: {
      type: "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/polyform/math/geometry.AABB]",
    },
  },
};

const ImageParameter: FlowNodeConfig = {
  title: "Image Parameter",
  subTitle: "Image",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [{ name: ParameterOutPortName, type: "image.Image" }],
  widgets: [{ type: "image", config: {} }],
  style: ParameterStyle,
  metadata: {
    typeData: { type: "github.com/EliCDavis/polyform/generator/parameter.Image" },
  },
};

const FileParameter: FlowNodeConfig = {
  title: "File Parameter",
  subTitle: "File",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [{ name: ParameterOutPortName, type: "[]uint8" }],
  style: ParameterStyle,
  metadata: {
    typeData: { type: "github.com/EliCDavis/polyform/generator/parameter.File" },
  },
};

const ColorParameter: FlowNodeConfig = {
  title: "Color Parameter",
  subTitle: "Color",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [{ name: ParameterOutPortName, type: "github.com/EliCDavis/polyform/drawing/coloring.Color" }],
  widgets: [{ type: "color", config: { property: "value" } }],
  style: ParameterStyle,
  metadata: {
    typeData: {
      type: "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/polyform/drawing/coloring.Color]",
    },
  },
};

const Vector3Parameter: FlowNodeConfig = {
  title: "Vector3 Parameter",
  subTitle: "Vector3",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [
    { name: ParameterOutPortName, type: "github.com/EliCDavis/vector/vector3.Vector[float64]" },
  ],
  widgets: [
    { type: "number", config: { property: "x" } },
    { type: "number", config: { property: "y" } },
    { type: "number", config: { property: "z" } },
  ],
  style: ParameterStyle,
  metadata: {
    typeData: {
      type: "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/vector/vector3.Vector[float64]]",
    },
  },
};

const Vector3ArrayParameter: FlowNodeConfig = {
  title: "Vector3 Array Parameter",
  subTitle: "Vector3 Array",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [
    { name: ParameterOutPortName, type: "[]github.com/EliCDavis/vector/vector3.Vector[float64]" },
  ],
  style: ParameterStyle,
  metadata: {
    typeData: {
      type: "github.com/EliCDavis/polyform/generator/parameter.Value[[]github.com/EliCDavis/vector/vector3.Vector[float64]]",
    },
  },
};

const Vector2Parameter: FlowNodeConfig = {
  title: "Vector2 Parameter",
  subTitle: "Vector2",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [
    { name: ParameterOutPortName, type: "github.com/EliCDavis/vector/vector2.Vector[float64]" },
  ],
  widgets: [
    { type: "number", config: { property: "x" } },
    { type: "number", config: { property: "y" } },
  ],
  style: ParameterStyle,
  metadata: {
    typeData: {
      type: "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/vector/vector2.Vector[float64]]",
    },
  },
};

const BoolParameter: FlowNodeConfig = {
  title: "Bool Parameter",
  subTitle: "Bool",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [{ name: ParameterOutPortName, type: "bool" }],
  widgets: [{ type: "toggle", config: { property: "value" } }],
  style: ParameterStyle,
  metadata: {
    typeData: { type: "github.com/EliCDavis/polyform/generator/parameter.Value[bool]" },
  },
};

const StringParameter: FlowNodeConfig = {
  title: "String Parameter",
  subTitle: "String",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [{ name: ParameterOutPortName, type: "string" }],
  widgets: [{ type: "string", config: { property: "value" } }],
  style: ParameterStyle,
  metadata: {
    typeData: { type: "github.com/EliCDavis/polyform/generator/parameter.Value[string]" },
  },
};

const valueType = (t: string) => `github.com/EliCDavis/polyform/generator/parameter.Value[${t}]`;
const vector = (n: number, t: string) => `github.com/EliCDavis/vector/vector${n}.Vector[${t}]`;

const Vector4Parameter: FlowNodeConfig = {
  title: "Vector4 Parameter",
  subTitle: "Vector4",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [{ name: ParameterOutPortName, type: vector(4, "float64") }],
  widgets: [
    { type: "number", config: { property: "x" } },
    { type: "number", config: { property: "y" } },
    { type: "number", config: { property: "z" } },
    { type: "number", config: { property: "w" } },
  ],
  style: ParameterStyle,
  metadata: { typeData: { type: valueType(vector(4, "float64")) } },
};

// Item widgets are added by the controller once the current value is known.
function arrayParameter(title: string, itemType: string): FlowNodeConfig {
  return {
    title: `${title} Array Parameter`,
    subTitle: `${title} Array`,
    canEditTitle: true,
    canEditInfo: true,
    outputs: [{ name: ParameterOutPortName, type: `[]${itemType}` }],
    style: ParameterStyle,
    metadata: { typeData: { type: valueType(`[]${itemType}`) } },
  };
}

const quaternionType = "github.com/EliCDavis/polyform/math/quaternion.Quaternion";
const trsType = "github.com/EliCDavis/polyform/math/trs.TRS";

const eulerWidgets = [
  { type: "number", config: { property: "x" } },
  { type: "number", config: { property: "y" } },
  { type: "number", config: { property: "z" } },
];

const QuaternionParameter: FlowNodeConfig = {
  title: "Rotation Parameter",
  subTitle: "Rotation (degrees)",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [{ name: ParameterOutPortName, type: quaternionType }],
  widgets: eulerWidgets,
  style: ParameterStyle,
  metadata: { typeData: { type: valueType(quaternionType) } },
};

const TRSParameter: FlowNodeConfig = {
  title: "Transform Parameter",
  subTitle: "Transform",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [{ name: ParameterOutPortName, type: trsType }],
  widgets: [
    { type: "text", config: { value: "position" } },
    { type: "number", config: { property: "px" } },
    { type: "number", config: { property: "py" } },
    { type: "number", config: { property: "pz" } },
    { type: "text", config: { value: "rotation (degrees)" } },
    { type: "number", config: { property: "rx" } },
    { type: "number", config: { property: "ry" } },
    { type: "number", config: { property: "rz" } },
    { type: "text", config: { value: "scale" } },
    { type: "number", config: { property: "sx" } },
    { type: "number", config: { property: "sy" } },
    { type: "number", config: { property: "sz" } },
  ],
  style: ParameterStyle,
  metadata: { typeData: { type: valueType(trsType) } },
};

const colorType = "github.com/EliCDavis/polyform/drawing/coloring.Color";
const gradientType = `github.com/EliCDavis/polyform/drawing/coloring.Gradient[${colorType}]`;

// Stop widgets are added by the controller once the current value is known.
const ColorGradientParameter: FlowNodeConfig = {
  title: "Color Gradient Parameter",
  subTitle: "Color Gradient",
  canEditTitle: true,
  canEditInfo: true,
  outputs: [{ name: ParameterOutPortName, type: gradientType }],
  style: ParameterStyle,
  metadata: { typeData: { type: valueType(gradientType) } },
};

export const parameterNodeConfigs: Record<string, FlowNodeConfig> = {
  "Parameters/vector4.Vector[float64]": Vector4Parameter,
  "Parameters/[]coloring.Color": arrayParameter("Color", colorType),
  [`Parameters/coloring.Gradient[${colorType}]`]: ColorGradientParameter,
  "Parameters/quaternion.Quaternion": QuaternionParameter,
  "Parameters/trs.TRS": TRSParameter,
  "Parameters/[]trs.TRS": arrayParameter("Transform", trsType),
  "Parameters/[]float64": arrayParameter("Float64", "float64"),
  "Parameters/[]int": arrayParameter("Int", "int"),
  "Parameters/[]string": arrayParameter("String", "string"),
  "Parameters/[]vector2.Vector[float64]": arrayParameter("Vector2", vector(2, "float64")),
  "Parameters/[]vector2.Vector[int]": arrayParameter("Vector2 Int", vector(2, "int")),
  "Parameters/[]vector3.Vector[int]": arrayParameter("Vector3 Int", vector(3, "int")),
  "Parameters/bool": BoolParameter,
  "Parameters/int": IntParameter,
  "Parameters/float64": FloatParameter,
  "Parameters/coloring.Color": ColorParameter,
  "Parameters/string": StringParameter,
  "Parameters/geometry.AABB": AABBParameter,
  "Parameters/image.Image": ImageParameter,
  "Parameters/File": FileParameter,
  "Parameters/vector3.Vector[float64]": Vector3Parameter,
  "Parameters/[]vector3.Vector[float64]": Vector3ArrayParameter,
  "Parameters/vector2.Vector[float64]": Vector2Parameter,
};
