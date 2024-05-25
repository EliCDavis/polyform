package swagger

type ParameterLocation string

const (
	PathParameterLocation   ParameterLocation = "path"
	QueryParameterLocation  ParameterLocation = "query"
	HeaderParameterLocation ParameterLocation = "header"
	BodyParameterLocation   ParameterLocation = "body"
	FormParameterLocation   ParameterLocation = "formData"
)

type Parameter struct {
	// Required. The location of the parameter. Possible values are "query",
	// "header", "path", "formData" or "body".
	In ParameterLocation `json:"in"`

	// Required. The name of the parameter. Parameter names are case sensitive.
	//  * If in is "path", the name field MUST correspond to the associated
	//    path segment from the path field in the Paths Object. See Path
	//    Templating for further information.
	//  * For all other cases, the name corresponds to the parameter name used
	//    based on the in property
	Name string `json:"name,omitempty"`

	// A brief description of the parameter. This could contain examples of
	// use. GFM syntax can be used for rich text representation.
	Description string `json:"description,omitempty"`

	// Determines whether this parameter is mandatory. If the parameter is in
	// "path", this property is required and its value MUST be true. Otherwise,
	// the property MAY be included and its default value is false.
	Required bool `json:"required,omitempty"`

	// If in is "body": Required. The schema defining the type used for the
	// body parameter.
	Schema any `json:"schema,omitempty"`

	// If in is any value other than "body"
	// Required. The type of the parameter. Since the parameter is not located
	// at the request body, it is limited to simple types (that is, not an
	// object). The value MUST be one of "string", "number", "integer",
	// "boolean", "array" or "file". If type is "file", the consumes MUST be
	// either "multipart/form-data", " application/x-www-form-urlencoded" or
	// both and the parameter MUST be in "formData".
	Type string `json:"type,omitempty"`
}
