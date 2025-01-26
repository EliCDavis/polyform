package room

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator/schema"
)

func DefaultWebScene() *schema.WebScene {
	return &schema.WebScene{
		RenderWireframe: false,
		AntiAlias:       true,
		XrEnabled:       false,
		Fog: schema.WebSceneFog{
			Color: coloring.WebColor{R: 0xa0, G: 0xa0, B: 0xa0, A: 255},
			Near:  10,
			Far:   50,
		},
		Background: coloring.WebColor{R: 0xa0, G: 0xa0, B: 0xa0, A: 255},
		Lighting:   coloring.White(),
		Ground:     coloring.WebColor{R: 0xcb, G: 0xcb, B: 0xcb, A: 255},
	}
}
