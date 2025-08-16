package room

import (
	"github.com/EliCDavis/polyform/generator/schema"
)

func DefaultWebScene() *schema.WebScene {
	return &schema.WebScene{
		RenderWireframe: false,
		AntiAlias:       true,
		XrEnabled:       false,
		Fog: schema.WebSceneFog{
			Color: "#A0A0A0",
			Near:  10,
			Far:   50,
		},
		Background: "#A0A0A0",
		Lighting:   "#FFFFFF",
		Ground:     "#CBCBCB",
	}
}
