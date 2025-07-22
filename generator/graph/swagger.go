package graph

import (
	"encoding/json"
	"fmt"
	"io"
	"unicode"

	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/nodes"
)

func findAllRefernecesObjects(allDefs map[string]swagger.Definition, def swagger.Definition) {
	for _, p := range def.Properties {
		recursivelyFindCommonSwaggerProperties(allDefs, p)
	}
}

func recursivelyFindCommonSwaggerProperties(allDefs map[string]swagger.Definition, prop swagger.Property) {
	switch prop.Ref {
	case "#/definitions/AABB":
		allDefs[swagger.AABBDefinitionName] = swagger.AABBDefinition

	case "#/definitions/Float2":
		allDefs[swagger.Float2DefinitionName] = swagger.Float2Definition

	case "#/definitions/Float3":
		allDefs[swagger.Float3DefinitionName] = swagger.Float3Definition

	case "#/definitions/Float4":
		allDefs[swagger.Float4DefinitionName] = swagger.Float4Definition

	case "#/definitions/Int2":
		allDefs[swagger.Int2DefinitionName] = swagger.Int2Definition

	case "#/definitions/Int3":
		allDefs[swagger.Int3DefinitionName] = swagger.Int3Definition

	case "#/definitions/Int4":
		allDefs[swagger.Int4DefinitionName] = swagger.Int4Definition
	}

	for _, p := range prop.Properties {
		if p.Type == swagger.ObjectPropertyType {
			recursivelyFindCommonSwaggerProperties(allDefs, p)
		}
	}

}

func swaggerDefinitionNameFromProducerPath(producerPath string) string {
	var output []rune //create an output slice
	isWord := true
	for _, val := range producerPath {
		if isWord && unicode.IsLetter(val) { //check if character is a letter convert the first character to upper case
			output = append(output, unicode.ToUpper(val))
			isWord = false
		} else if unicode.IsLetter(val) {
			output = append(output, unicode.ToLower(val))
		} else {
			isWord = true
		}
	}
	return string(output) + "Request"
}

func WriteSwagger(instance *Instance, out io.Writer) error {
	jsonData, err := json.MarshalIndent(SwaggerSpec(instance), "", "    ")
	if err != nil {
		return err
	}

	_, err = out.Write(jsonData)
	return err
}

func SwaggerSpec(instance *Instance) swagger.Spec {
	jsonApplication := "application/json"
	profileRequestObject := "VariableProfile"
	createManifestResponse := "CreateManifestResponse"

	paths := make(map[string]swagger.Path)

	err := ForeachManifestNodeOutput(instance, func(nodeId string, node nodes.Node, output nodes.Output[manifest.Manifest]) error {
		description := ""
		if describable, ok := node.(nodes.Describable); ok {
			description = describable.Description()
		}

		nodeName := nodeId
		if name, ok := instance.IsPortNamed(node, output.Name()); ok {
			nodeName = name
		}

		paths[fmt.Sprintf("/manifest/%s/%s", nodeName, output.Name())] = swagger.Path{
			swagger.PostRequestMethod: swagger.RequestDefinition{
				Description: description,
				Produces:    []string{jsonApplication},
				Consumes:    []string{jsonApplication},
				Parameters: []swagger.Parameter{
					{
						In:       "body",
						Name:     "Request",
						Required: false,
						Schema: swagger.SchemaObject{
							Ref: swagger.DefinitionRefPath(profileRequestObject),
						},
					},
				},
				Responses: map[int]swagger.Response{
					200: {
						Description: "Success",
						Schema: &swagger.Property{
							Ref: fmt.Sprintf("#/definitions/%s", createManifestResponse),
						},
					},
				},
			},
		}
		return nil
	})

	if err != nil {
		panic(fmt.Errorf("failed iterating over node manifest ports: %w", err))
	}

	paths["/manifest/{id}/{entry}"] = swagger.Path{
		swagger.GetRequestMethod: swagger.RequestDefinition{
			Parameters: []swagger.Parameter{
				{
					In:          swagger.PathParameterLocation,
					Name:        "id",
					Required:    true,
					Type:        "string",
					Description: "ID of the produced manifest to query",
				},
				{
					In:          swagger.PathParameterLocation,
					Name:        "entry",
					Required:    true,
					Type:        "string",
					Description: "Entry in the produced manifest to retrieve",
				},
			},
			Produces: []string{"application/octet-stream"},
			Responses: map[int]swagger.Response{
				200: {
					Description: "Success",
					Schema: &swagger.Property{
						Type: "file",
					},
				},
			},
		},
	}

	paths["/profile"] = swagger.Path{
		swagger.GetRequestMethod: swagger.RequestDefinition{
			Produces: []string{jsonApplication},
			Responses: map[int]swagger.Response{
				200: {
					Description: "Success",
					Schema: &swagger.Definition{
						Type:                 "object",
						AdditionalProperties: true,
					},
				},
			},
		},
	}

	requestDefinition := instance.SwaggerDefinition()
	definitions := map[string]swagger.Definition{
		profileRequestObject: requestDefinition,
		"Metadata": {
			Type:                 "object",
			AdditionalProperties: true,
		},
		"Entry": {
			Type: "object",
			Properties: map[string]swagger.Property{
				"metadata": {
					Ref: "#/definitions/Metadata",
				},
			},
		},
		"Entries": {
			Type: "object",
			AdditionalProperties: &swagger.Property{
				Ref: "#/definitions/Entry",
			},
		},
		"Manifest": {
			Type: "object",
			Properties: map[string]swagger.Property{
				"main":    {Type: swagger.StringPropertyType},
				"entries": {Ref: "#/definitions/Entries"},
			},
		},
		"CreateManifestResponse": {
			Type: "object",
			Properties: map[string]swagger.Property{
				"id":       {Type: swagger.StringPropertyType},
				"manifest": {Ref: "#/definitions/Manifest"},
			},
		},
	}

	findAllRefernecesObjects(definitions, requestDefinition)

	return swagger.Spec{
		Version: "2.0",
		Info: &swagger.Info{
			Title:       instance.GetName(),
			Description: instance.GetDescription(),
			Version:     instance.GetVersion(),
		},
		Paths:       paths,
		Definitions: definitions,
	}
}
