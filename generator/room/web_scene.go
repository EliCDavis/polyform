package room

import (
	"github.com/EliCDavis/polyform/generator/persistence"
)

func DefaultWebScene() *persistence.WebScene {
	return &persistence.WebScene{
		RenderWireframe: false,
		AntiAlias:       true,
		XrEnabled:       false,
		Fog: persistence.WebSceneFog{
			Color: persistence.WebColor{R: 0xA0, B: 0xA0, G: 0xA0},
			Near:  10,
			Far:   50,
		},
		Background: persistence.WebColor{R: 0xA0, B: 0xA0, G: 0xA0},
		Lighting:   persistence.WebColor{R: 0xFF, B: 0xFF, G: 0xFF},
		Ground:     persistence.WebColor{R: 0xCB, B: 0xCB, G: 0xCB},
	}
}
