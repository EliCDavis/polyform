package room

type WebSceneFog struct {
	Color string  `json:"color"`
	Near  float64 `json:"near"`
	Far   float64 `json:"far"`
}

type WebScene struct {
	RenderWireframe bool        `json:"renderWireframe"`
	Fog             WebSceneFog `json:"fog"`
	Background      string      `json:"background"`
	Lighting        string      `json:"lighting"`
	Ground          string      `json:"ground"`
}

func DefaultWebScene() *WebScene {
	return &WebScene{
		RenderWireframe: false,
		Fog: WebSceneFog{
			Color: "#a0a0a0", //color.RGBA{R: 0xa0, G: 0xa0, B: 0xa0},
			Near:  10,
			Far:   50,
		},
		Background: "#a0a0a0", // color.RGBA{R: 0xa0, G: 0xa0, B: 0xa0},
		Lighting:   "#FFFFFF", //color.RGBA{R: 255, G: 255, B: 255},
		Ground:     "#cbcbcb", //color.RGBA{R: 0xcb, G: 0xcb, B: 0xcb},
	}
}
