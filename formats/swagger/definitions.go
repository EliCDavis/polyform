package swagger

const Int2DefinitionName = "Int2"
const Int3DefinitionName = "Int3"
const Int4DefinitionName = "Int4"
const Float2DefinitionName = "Float2"
const Float3DefinitionName = "Float3"
const Float4DefinitionName = "Float4"
const AABBDefinitionName = "AABB"
const TRSDefinitionName = "TRS"
const ColorGradientDefinitionName = "ColorGradient"

var floatComponent = Property{
	Type:    NumberPropertyType,
	Format:  DoublePropertyFormat,
	Example: "1.0",
}

var intComponent = Property{
	Type:    IntegerPropertyType,
	Example: "1",
}

var Float2Definition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"x": floatComponent,
		"y": floatComponent,
	},
}

var Float3Definition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"x": floatComponent,
		"y": floatComponent,
		"z": floatComponent,
	},
}

var Float4Definition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"x": floatComponent,
		"y": floatComponent,
		"z": floatComponent,
		"w": floatComponent,
	},
}

var AABBDefinition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"min": {
			Ref: DefinitionRefPath(Float3DefinitionName),
		},
		"max": {
			Ref: DefinitionRefPath(Float3DefinitionName),
		},
	},
}

var TRSDefinition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"position": {
			Ref: DefinitionRefPath(Float3DefinitionName),
		},
		"rotation": {
			Ref: DefinitionRefPath(Float4DefinitionName),
		},
		"scale": {
			Ref: DefinitionRefPath(Float3DefinitionName),
		},
	},
}

var ColorGradientDefinition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"keys": {
			Type: ArrayPropertyType,
			Items: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"time":  map[string]any{"type": NumberPropertyType, "format": DoublePropertyFormat},
					"value": map[string]any{"type": StringPropertyType, "format": "color"},
				},
			},
		},
	},
}

var Int2Definition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"x": intComponent,
		"y": intComponent,
	},
}

var Int3Definition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"x": intComponent,
		"y": intComponent,
		"z": intComponent,
	},
}

var Int4Definition = Definition{
	Type: "object",
	Properties: map[string]Property{
		"x": intComponent,
		"y": intComponent,
		"z": intComponent,
		"w": intComponent,
	},
}
