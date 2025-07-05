package graph

import (
	"encoding/json"
	"io"
	"unicode"

	"github.com/EliCDavis/polyform/formats/swagger"
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

	paths := make(map[string]swagger.Path)

	for _, path := range instance.ProducerNames() {
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

		// producer := instanceInstance.Producer(path)
	}

	requestDefinition := instance.SwaggerDefinition()
	definitions := map[string]swagger.Definition{
		"ManifestRequest": requestDefinition,
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
		// Definitions: map[string]swagger.Definition{
		// swagger.Vector3DefinitionName: swagger.Vector3Definition,
		// swagger.AABBDefinitionName:    swagger.AABBDefinition,
		// },
	}
}
