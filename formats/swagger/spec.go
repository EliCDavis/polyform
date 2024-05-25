package swagger

import "errors"

func DefinitionRefPath(r string) string {
	if r == "" {
		panic(errors.New("can not build definition reference from an empty string"))
	}
	return "#/definitions/" + r
}

type SchemaObject struct {
	Ref string `json:"$ref,omitempty"`
}

type RequestMethod string

const (
	GetRequestMethod  RequestMethod = "get"
	PostRequestMethod RequestMethod = "post"
)

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}

type Response struct {
	Description string `json:"description,omitempty"`
	// This is a lookup into the responses section, not definitions
	Ref    string `json:"$ref,omitempty"`
	Schema any    `json:"schema,omitempty"`
}

type RequestDefinition struct {
	Summary     string           `json:"summary"`
	Description string           `json:"description"`
	Produces    []string         `json:"produces"`
	Consumes    []string         `json:"consumes"`
	Responses   map[int]Response `json:"responses"`
	Parameters  []Parameter      `json:"parameters"`
}

type Path map[RequestMethod]RequestDefinition

type Definition struct {
	// Type is probably always "object"
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
	Example    any                 `json:"example,omitempty"`
}

type Spec struct {
	Version     string                `json:"swagger"` // Must be 2.0
	Info        *Info                 `json:"info,omitempty"`
	Paths       map[string]Path       `json:"paths"`
	Definitions map[string]Definition `json:"definitions,omitempty"`
	Responses   map[string]Response   `json:"responses,omitempty"`
}
