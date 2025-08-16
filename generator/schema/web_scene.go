package schema

type WebSceneFog struct {
	Color string  `json:"color"`
	Near  float32 `json:"near"`
	Far   float32 `json:"far"`
}

type WebScene struct {
	RenderWireframe bool        `json:"renderWireframe"`
	AntiAlias       bool        `json:"antiAlias"`
	XrEnabled       bool        `json:"xrEnabled"`
	Fog             WebSceneFog `json:"fog"`
	Background      string      `json:"background"`
	Lighting        string      `json:"lighting"`
	Ground          string      `json:"ground"`
}
