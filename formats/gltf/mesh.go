package gltf

// A set of primitives to be rendered.  Its global transform is defined by a node that references it.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/mesh.schema.json
type Mesh struct {
	ChildOfRootProperty
	Primitives []Primitive `json:"primitives"`
	Weights    []float64   `json:"weights,omitempty"`
}

type PrimitiveMode int

const (
	PrimitiveMode_POINTS         PrimitiveMode = 0
	PrimitiveMode_LINES          PrimitiveMode = 1
	PrimitiveMode_LINE_LOOP      PrimitiveMode = 2
	PrimitiveMode_LINE_STRIP     PrimitiveMode = 3
	PrimitiveMode_TRIANGLES      PrimitiveMode = 4
	PrimitiveMode_TRIANGLE_STRIP PrimitiveMode = 5
	PrimitiveMode_TRIANGLE_FAN   PrimitiveMode = 6
)

// https://registry.khronos.org/glTF/specs/2.0/glTF-2.0.pdf
// Page 39
const (
	POSITION   = "POSITION"
	NORMAL     = "NORMAL"
	TANGENT    = "TANGENT"
	TEXCOORD_0 = "TEXCOORD_0"
	COLOR_0    = "COLOR_0"
	JOINTS_0   = "JOINTS_0"
	WEIGHTS_0  = "WEIGHTS_0"
)

// Geometry to be rendered with the given material.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/mesh.primitive.schema.json
type Primitive struct {
	Property
	Attributes map[string]GltfId `json:"attributes"`         // A plain JSON object, where each key corresponds to a mesh attribute semantic and each value is the index of the accessor containing attribute's data.
	Indices    *GltfId           `json:"indices,omitempty"`  // The index of the accessor that contains the vertex indices.  When this is undefined, the primitive defines non-indexed geometry.  When defined, the accessor **MUST** have `SCALAR` type and an unsigned integer component type.
	Material   *GltfId           `json:"material,omitempty"` // The index of the material to apply to this primitive when rendering.
	Targets    []GltfId          `json:"targets,omitempty"`  // A plain JSON object specifying attributes displacements in a morph target, where each key corresponds to one of the three supported attribute semantic (`POSITION`, `NORMAL`, or `TANGENT`) and each value is the index of the accessor containing the attribute displacements' data.
	Mode       *PrimitiveMode    `json:"mode,omitempty"`     // The topology type of primitives to render.
}
