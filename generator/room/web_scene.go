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
			Color: schema.WebColor{R: 0xA0, B: 0xA0, G: 0xA0},
			Near:  10,
			Far:   50,
		},
		Background: schema.WebColor{R: 0xA0, B: 0xA0, G: 0xA0},
		Lighting:   schema.WebColor{R: 0xFF, B: 0xFF, G: 0xFF},
		Ground:     schema.WebColor{R: 0xCB, B: 0xCB, G: 0xCB},
	}
}
