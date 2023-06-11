package main

import (
	"image/color"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/polyform/rendering/materials"
	"github.com/EliCDavis/vector/vector3"
)

func plumbob() modeling.Mesh {
	return primitives.
		UVSphere(1, 2, 8).
		Transform(
			meshops.ScaleAttribute3DTransformer{
				Amount: vector3.New(1., 2., 1.),
			},
			meshops.UnweldTransformer{},
			meshops.FlatNormalsTransformer{},
		).
		SetMaterial(modeling.Material{
			Name:              "Plumbob",
			DiffuseColor:      color.RGBA{0, 255, 0, 255},
			Transparency:      .1,
			SpecularHighlight: 50,
			SpecularColor:     color.RGBA{0, 255, 0, 255},
		})
}

func render() {
	jewelColor := vector3.New(0., 0.9, 0.4)
	jewelMat := materials.NewDielectricWithColor(1.5, jewelColor)

	scene := []rendering.Hittable{
		rendering.NewMesh(
			plumbob(),
			jewelMat,
		),
		rendering.NewMesh(
			plumbob().
				Scale(vector3.Fill(0.9)).
				FlipTriWinding(),
			jewelMat,
		),
	}

	origin := vector3.New(2.1, 0.5, 2.1)
	lookat := vector3.Zero[float64]()
	camera := rendering.NewDefaultCamera(1, origin, lookat, 0, 0)

	err := rendering.RenderToFile(50, 200, 500, scene, camera, "tmp/plumbob/preview.png", nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	err := obj.Save(
		"tmp/plumbob/plumbob.obj",
		plumbob(),
	)
	if err != nil {
		panic(err)
	}
	render()
}
