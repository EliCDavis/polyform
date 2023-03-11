package gltf

import "fmt"

type AccessorComponentType int

const (
	AccessorComponentType_BYTE           AccessorComponentType = 5120
	AccessorComponentType_UNSIGNED_BYTE  AccessorComponentType = 5121
	AccessorComponentType_SHORT          AccessorComponentType = 5122
	AccessorComponentType_UNSIGNED_SHORT AccessorComponentType = 5123
	AccessorComponentType_UNSIGNED_INT   AccessorComponentType = 5125
	AccessorComponentType_FLOAT          AccessorComponentType = 5126
)

func (act AccessorComponentType) Size() int {
	switch act {
	case AccessorComponentType_BYTE:
		return 1

	case AccessorComponentType_UNSIGNED_BYTE:
		return 1

	case AccessorComponentType_SHORT:
		return 2

	case AccessorComponentType_UNSIGNED_SHORT:
		return 2

	case AccessorComponentType_UNSIGNED_INT:
		return 4

	case AccessorComponentType_FLOAT:
		return 4
	}

	panic(fmt.Errorf("unimplemented accessor component type: %d", act))
}

type AccessorType string

const (
	AccessorType_SCALAR AccessorType = "SCALAR"
	AccessorType_VEC2   AccessorType = "VEC2"
	AccessorType_VEC3   AccessorType = "VEC3"
	AccessorType_VEC4   AccessorType = "VEC4"
	AccessorType_MAT2   AccessorType = "MAT2"
	AccessorType_MAT3   AccessorType = "MAT3"
	AccessorType_MAT4   AccessorType = "MAT4"
)

// A typed view into a buffer view that contains raw binary data.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/accessor.schema.json
type Accessor struct {
	ChildOfRootProperty
	BufferView    *GltfId               `json:"bufferView,omitempty"` // The index of the buffer view. When undefined, the accessor **MUST** be initialized with zeros; `sparse` property or extensions **MAY** override zeros with actual values.
	ByteOffset    int                   `json:"byteOffset,omitempty"` // The offset relative to the start of the buffer view in bytes.  This **MUST** be a multiple of the size of the component datatype. This property **MUST NOT** be defined when `bufferView` is undefined.
	ComponentType AccessorComponentType `json:"componentType"`        // The datatype of the accessor's components.  UNSIGNED_INT type **MUST NOT** be used for any accessor that is not referenced by `mesh.primitive.indices`.
	Normalized    bool                  `json:"normalized,omitempty"` // Specifies whether integer data values are normalized (`true`) to [0, 1] (for unsigned types) or to [-1, 1] (for signed types) when they are accessed. This property **MUST NOT** be set to `true` for accessors with `FLOAT` or `UNSIGNED_INT` component type.
	Type          AccessorType          `json:"type"`                 // Specifies if the accessor's elements are scalars, vectors, or matrices.
	Count         int                   `json:"count"`                // The number of elements referenced by this accessor, not to be confused with the number of bytes or number of components.
	Max           []float64             `json:"max,omitempty"`        // Maximum value of each component in this accessor.  Array elements **MUST** be treated as having the same data type as accessor's `componentType`. Both `min` and `max` arrays have the same length.  The length is determined by the value of the `type` property; it can be 1, 2, 3, 4, 9, or 16.\n\n`normalized` property has no effect on array values: they always correspond to the actual values stored in the buffer. When the accessor is sparse, this property **MUST** contain maximum values of accessor data with sparse substitution applied.
	Min           []float64             `json:"min,omitempty"`        // Minimum value of each component in this accessor.  Array elements **MUST** be treated as having the same data type as accessor's `componentType`. Both `min` and `max` arrays have the same length.  The length is determined by the value of the `type` property; it can be 1, 2, 3, 4, 9, or 16.\n\n`normalized` property has no effect on array values: they always correspond to the actual values stored in the buffer. When the accessor is sparse, this property **MUST** contain minimum values of accessor data with sparse substitution applied.
	// Sparse        Sparse                `json:"sparse,omitempty,omitempty"` // Sparse storage of elements that deviate from their initialization value.
}
