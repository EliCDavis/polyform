package schema

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
