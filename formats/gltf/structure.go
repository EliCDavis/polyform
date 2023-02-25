package gltf

type GltfId = int

// JSON object with extension-specific objects.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/extension.schema.json
type Extension struct {
	Properties           map[string]any `json:"properties"`
	AdditionalProperties map[string]any `json:"additionalProperties"`
}

// Although `extras` **MAY** have any type, it is common for applications to
// store and access custom data as key/value pairs. Therefore, `extras`
// **SHOULD** be a JSON object rather than a primitive value for best portability.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/extras.schema.json
type Extra = map[string]any

// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/glTFProperty.schema.json
type Property struct {
	Extensions map[string]Extension `json:"extensions,omitempty"`
	Extras     Extra                `json:"extras,omitempty"`
}

// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/glTFChildOfRootProperty.schema.json
type ChildOfRootProperty struct {
	Property
	// The user-defined name of this object.  This is not necessarily unique, e.g.,
	// an accessor and a buffer could have the same name, or two accessors could
	// even have the same name.
	Name string `json:"name,omitempty"`
}

// The root nodes of a scene.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/scene.schema.json
type Scene struct {
	ChildOfRootProperty
	Nodes []GltfId `json:"nodes"` // The indices of each root node.
}

// A buffer points to binary geometry, animation, or skins.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/buffer.schema.json
type Buffer struct {
	ChildOfRootProperty
	ByteLength int    `json:"byteLength"`    // The length of the buffer in bytes.
	URI        string `json:"uri,omitempty"` // The URI (or IRI) of the buffer.  Relative paths are relative to the current glTF asset.  Instead of referencing an external file, this field **MAY** contain a `data:`-URI.
}

type BufferViewTarget int

const (
	ARRAY_BUFFER         BufferViewTarget = 34962
	ELEMENT_ARRAY_BUFFER BufferViewTarget = 34963
)

// A view into a buffer generally representing a subset of the buffer.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/bufferView.schema.json
type BufferView struct {
	ChildOfRootProperty
	Buffer     int              `json:"buffer"`               // The index of the buffer
	ByteOffset int              `json:"byteOffset,omitempty"` // The offset into the buffer in bytes.
	ByteLength int              `json:"byteLength"`           // The length of the bufferView in bytes.
	ByteStride *int             `json:"byteStride,omitempty"` // The stride, in bytes, between vertex attributes.  When this is not defined, data is tightly packed. When two or more accessors use the same buffer view, this field **MUST** be defined.
	Target     BufferViewTarget `json:"target,omitempty"`     // The hint representing the intended GPU buffer type to use with this buffer view.
}

// A node in the node hierarchy.  When the node contains `skin`, all
// `mesh.primitives` **MUST** contain `JOINTS_0` and `WEIGHTS_0` attributes.
// A node **MAY** have either a `matrix` or any combination of
// `translation`/`rotation`/`scale` (TRS) properties. TRS properties are
// converted to matrices and postmultiplied in the `T * R * S` order to compose
// the transformation matrix; first the scale is applied to the vertices, then
// the rotation, and then the translation. If none are provided, the transform
// is the identity. When a node is targeted for animation (referenced by an
// animation.channel.target), `matrix` **MUST NOT** be present.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/node.schema.json
type Node struct {
	ChildOfRootProperty
	Camera      *GltfId      `json:"camera,omitempty"`      // The index of the camera referenced by this node.
	Children    []GltfId     `json:"children,omitempty"`    // The indices of this node's children.
	Skin        *GltfId      `json:"skin,omitempty"`        // The index of the skin referenced by this node. When a skin is referenced by a node within a scene, all joints used by the skin **MUST** belong to the same scene. When defined, `mesh` **MUST** also be defined.
	Matrix      *[16]float64 `json:"matrix,omitempty"`      // A floating-point 4x4 transformation matrix stored in column-major order.
	Mesh        *GltfId      `json:"mesh,omitempty"`        // The index of the mesh in this node.
	Rotation    *[4]float64  `json:"rotation,omitempty"`    // The node's unit quaternion rotation in the order (x, y, z, w), where w is the scalar.
	Scale       *[3]float64  `json:"scale,omitempty"`       // The node's non-uniform scale, given as the scaling factors along the x, y, and z axes.
	Translation *[3]float64  `json:"translation,omitempty"` // The node's translation along the x, y, and z axes.
	Weights     []float64    `json:"weights,omitempty"`     // The weights of the instantiated morph target. The number of array elements **MUST** match the number of morph targets of the referenced mesh. When defined, `mesh` **MUST** also be defined.
}

type ImageMimeType string

const (
	ImageMimeType_JPEG ImageMimeType = "image/jpeg"
	ImageMimeType_PNG  ImageMimeType = "image/png"
)

// Image data used to create a texture. Image **MAY** be referenced by an URI (or IRI) or a buffer view index.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/image.schema.json
type Image struct {
	ChildOfRootProperty
	MimeType   ImageMimeType `json:"mimeType,omitempty"`   // The image's media type. This field **MUST** be defined when `bufferView` is defined.
	URI        string        `json:"uri,omitempty"`        // The URI (or IRI) of the image.  Relative paths are relative to the current glTF asset.  Instead of referencing an external file, this field **MAY** contain a `data:`-URI. This field **MUST NOT** be defined when `bufferView` is defined.
	BufferView GltfId        `json:"bufferView,omitempty"` // The index of the bufferView that contains the image. This field **MUST NOT** be defined when `uri` is defined.
}

// "A texture and its sampler.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/texture.schema.json
type Texture struct {
	ChildOfRootProperty
	Sampler GltfId `json:"sampler,omitempty"`
	Source  GltfId `json:"source,omitempty"`
}

// Joints and matrices defining a skin.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/skin.schema.json
type Skin struct {
	ChildOfRootProperty
	InverseBindMatrices GltfId   `json:"inverseBindMatrices,omitempty"` // The index of the accessor containing the floating-point 4x4 inverse-bind matrices. Its `accessor.count` property **MUST** be greater than or equal to the number of elements of the `joints` array. When undefined, each matrix is a 4x4 identity matrix.
	Skeleton            GltfId   `json:"skeleton,omitempty"`            // The index of the node used as a skeleton root. The node **MUST** be the closest common root of the joints hierarchy or a direct or indirect parent node of the closest common root.
	Joints              []GltfId `json:"joints"`                        // Indices of skeleton nodes, used as joints in this skin.
}

type Camera struct{}

// The root object for a glTF asset.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/glTF.schema.json
type Gltf struct {
	Property

	ExtensionsUsed     []string `json:"extensionsUsed,omitempty"`     // Names of glTF extensions used in this asset.
	ExtensionsRequired []string `json:"extensionsRequired,omitempty"` // Names of glTF extensions required to properly load this asset.

	Accessors   []Accessor   `json:"accessors,omitempty"`   // An array of accessors.  An accessor is a typed view into a bufferView.
	Animations  []Animation  `json:"animations,omitempty"`  // An array of keyframe animations.
	Asset       Asset        `json:"asset"`                 // Metadata about the glTF asset.
	Buffers     []Buffer     `json:"buffers,omitempty"`     // An array of buffers.  A buffer points to binary geometry, animation, or skins.
	BufferViews []BufferView `json:"bufferViews,omitempty"` // An array of bufferViews.  A bufferView is a view into a buffer generally representing a subset of the buffer
	Cameras     []Camera     `json:"cameras,omitempty"`     // An array of cameras.  A camera defines a projection matrix.
	Images      []Image      `json:"images,omitempty"`      // An array of images.  An image defines data used to create a texture.
	Materials   []Material   `json:"materials,omitempty"`   // An array of materials.  A material defines the appearance of a primitive.
	Meshes      []Mesh       `json:"meshes,omitempty"`      // An array of meshes.  A mesh is a set of primitives to be rendered
	Nodes       []Node       `json:"nodes,omitempty"`       // An array of nodes.
	Samplers    []Sampler    `json:"samplers,omitempty"`    // An array of samplers.  A sampler contains properties for texture filtering and wrapping modes.
	Scene       int          `json:"scene,omitempty"`       // The index of the default scene.  This property **MUST NOT** be defined, when `scenes` is undefined.
	Scenes      []Scene      `json:"scenes,omitempty"`      // An array of scenes.
	Skins       []Skin       `json:"skins,omitempty"`       // An array of skins.  A skin is defined by joints and matrices.
	Textures    []Texture    `json:"textures,omitempty"`    // An array of textures
}
