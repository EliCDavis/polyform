package room

import "github.com/EliCDavis/polyform/drawing/coloring"

type WebSceneFog struct {
	Color coloring.WebColor `json:"color"`
	Near  float32           `json:"near"`
	Far   float32           `json:"far"`
}

type WebScene struct {
	RenderWireframe bool              `json:"renderWireframe"`
	AntiAlias       bool              `json:"antiAlias"`
	XrEnabled       bool              `json:"xrEnabled"`
	Fog             WebSceneFog       `json:"fog"`
	Background      coloring.WebColor `json:"background"`
	Lighting        coloring.WebColor `json:"lighting"`
	Ground          coloring.WebColor `json:"ground"`
}

func DefaultWebScene() *WebScene {
	return &WebScene{
		RenderWireframe: false,
		AntiAlias:       true,
		XrEnabled:       false,
		Fog: WebSceneFog{
			Color: coloring.WebColor{R: 0xa0, G: 0xa0, B: 0xa0, A: 255},
			Near:  10,
			Far:   50,
		},
		Background: coloring.WebColor{R: 0xa0, G: 0xa0, B: 0xa0, A: 255},
		Lighting:   coloring.White(),
		Ground:     coloring.WebColor{R: 0xcb, G: 0xcb, B: 0xcb, A: 255},
	}
}
