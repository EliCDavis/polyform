package generator

import (
	"encoding/json"
	"io"
	"strings"
	"unicode"

	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/nodes"
)

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

func buildSwaggerDefinitionForProducer(producer nodes.NodeOutput[artifact.Artifact]) swagger.Definition {
	props := make(map[string]swagger.Property)
	params := recurseDependenciesType[SwaggerParameter](producer.Node())

	for _, param := range params {
		paramName := strings.Replace(param.DisplayName(), " ", "", -1)
		props[paramName] = param.SwaggerProperty()
	}

	return swagger.Definition{
		Type:       "object",
		Properties: props,
	}
}

func (a App) WriteSwagger(out io.Writer) error {
	jsonData, err := json.MarshalIndent(a.SwaggerSpec(), "", "    ")
	if err != nil {
		return err
	}

	_, err = out.Write(jsonData)
	return err
}

func recursivelyFindCommonSwaggerProperties(allDefs map[string]swagger.Definition, prop swagger.Property) {
	switch prop.Ref {
	case "#/definitions/AABB":
		allDefs[swagger.AABBDefinitionName] = swagger.AABBDefinition

	case "#/definitions/Vector2":
		allDefs[swagger.Vector2DefinitionName] = swagger.Vector2Definition

	case "#/definitions/Vector3":
		allDefs[swagger.Vector3DefinitionName] = swagger.Vector3Definition

	case "#/definitions/Vector4":
		allDefs[swagger.Vector4DefinitionName] = swagger.Vector4Definition
	}

	for _, p := range prop.Properties {
		if p.Type == swagger.ObjectPropertyType {
			recursivelyFindCommonSwaggerProperties(allDefs, p)
		}
	}

}

func (a App) SwaggerSpec() swagger.Spec {
	jsonApplication := "application/json"

	paths := make(map[string]swagger.Path)

	definitions := make(map[string]swagger.Definition)

	for path, producer := range a.Producers {
		definitionName := swaggerDefinitionNameFromProducerPath(path)

		paths["/producer/value/"+path] = swagger.Path{
			// Post required for bodys per HTTP spec.
			swagger.PostRequestMethod: swagger.RequestDefinition{
				// Summary:     "Test",
				// Description: "???",
				Produces: []string{
					// ???? How do we do this.
				},
				Consumes: []string{jsonApplication},
				Parameters: []swagger.Parameter{
					{
						In:       "body",
						Name:     "Request",
						Required: false,
						Schema: swagger.SchemaObject{
							Ref: swagger.DefinitionRefPath(definitionName),
						},
					},
				},
				Responses: map[int]swagger.Response{
					200: {
						Description: "Producer Payload",
					},
				},
			},
		}

		definitions[definitionName] = buildSwaggerDefinitionForProducer(producer)
	}

	for _, def := range definitions {
		for _, p := range def.Properties {
			recursivelyFindCommonSwaggerProperties(definitions, p)
		}
	}

	return swagger.Spec{
		Version: "2.0",
		Info: &swagger.Info{
			Title:       a.Name,
			Description: a.Description,
			Version:     a.Version,
		},
		Paths:       paths,
		Definitions: definitions,
		// Definitions: map[string]swagger.Definition{
		// swagger.Vector3DefinitionName: swagger.Vector3Definition,
		// swagger.AABBDefinitionName:    swagger.AABBDefinition,
		// },
	}
}
