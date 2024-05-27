package swagger

type PropertyType string

const (
	ObjectPropertyType  PropertyType = "object"
	StringPropertyType  PropertyType = "string"
	IntegerPropertyType PropertyType = "integer"
	NumberPropertyType  PropertyType = "number"
	BooleanPropertyType PropertyType = "boolean"
	ArrayPropertyType   PropertyType = "array"
)

type PropertyFormat string

const (
	// Integer ================================================================

	Int32PropertyFormat PropertyFormat = "int32"
	Int64PropertyFormat PropertyFormat = "int64"

	// Number =================================================================

	FloatPropertyFormat  PropertyFormat = "float"
	DoublePropertyFormat PropertyFormat = "double"

	// String =================================================================

	// Base64-encoded characters, for example, U3dhZ2dlciByb2Nrcw==
	BytePropertyFormat PropertyFormat = "byte"

	// Binary data, used to describe files
	BinaryPropertyFormat PropertyFormat = "binary"

	// Full-date notation as defined by RFC 3339, section 5.6, for example,
	// 2017-07-21
	DatePropertyFormat PropertyFormat = "date"

	// The date-time notation as defined by RFC 3339, section 5.6, for example,
	// 2017-07-21T17:32:28Z
	DateTimePropertyFormat PropertyFormat = "date-time"

	// A hint to UIs to mask the input
	PasswordPropertyFormat PropertyFormat = "password"
)

type Property struct {
	Type PropertyType `json:"type,omitempty"`

	// An optional format modifier serves as a hint at the contents and format
	// of the string
	Format PropertyFormat `json:"format,omitempty"`

	Example     any `json:"example,omitempty"`
	Ref         any `json:"$ref,omitempty"`
	Description any `json:"description,omitempty"`
	Items       any `json:"items,omitempty"`

	// If type is object
	Properties []Property `json:"properties,omitempty"`
}
