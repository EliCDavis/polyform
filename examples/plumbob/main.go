package main

import (
	"image/color"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/vector"
)

func main() {
	err := obj.Save(
		"tmp/plumbob/plumbob.obj",
		primitives.
			UVSphere(1, 2, 8).
			Scale(vector.Vector3Zero(), vector.NewVector3(1, 2, 1)).
			Unweld().
			CalculateFlatNormals().
			SetMaterial(modeling.Material{
				Name:              "Plumbob",
				DiffuseColor:      color.RGBA{0, 255, 0, 255},
				Transparency:      .1,
				SpecularHighlight: 50,
				SpecularColor:     color.RGBA{0, 255, 0, 255},
			}),
	)

	if err != nil {
		panic(err)
	}
}
