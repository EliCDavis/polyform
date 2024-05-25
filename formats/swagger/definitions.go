package swagger

const Vector2DefinitionName = "Vector2"
const Vector3DefinitionName = "Vector3"
const Vector4DefinitionName = "Vector4"
const AABBDefinitionName = "AABB"

var vectorComponent = Property{
	Type:    NumberPropertyType,
	Format:  DoublePropertyFormat,
	Example: "1.0",
}

var Vector2Definition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"x": vectorComponent,
		"y": vectorComponent,
	},
}

var Vector3Definition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"x": vectorComponent,
		"y": vectorComponent,
		"z": vectorComponent,
	},
}

var Vector4Definition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"x": vectorComponent,
		"y": vectorComponent,
		"z": vectorComponent,
		"w": vectorComponent,
	},
}

var AABBDefinition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"min": {
			Ref: DefinitionRefPath(Vector3DefinitionName),
		},
		"max": {
			Ref: DefinitionRefPath(Vector3DefinitionName),
		},
	},
}
