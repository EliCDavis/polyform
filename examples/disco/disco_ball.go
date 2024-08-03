package main

import (
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type DiscoBallNode = nodes.StructNode[[]gltf.PolyformModel, DiscoBallNodeData]

type DiscoBallNodeData struct {
	Color       nodes.NodeOutput[coloring.WebColor]
	Radius      nodes.NodeOutput[float64]
	PanelOffset nodes.NodeOutput[float64]
	Height      nodes.NodeOutput[float64]
	Rows        nodes.NodeOutput[int]
	Columns     nodes.NodeOutput[int]
}

func (cn DiscoBallNodeData) Process() ([]gltf.PolyformModel, error) {
	ballColor := cn.Color.Value()
	discoballRadius := cn.Radius.Value()

	discoball := primitives.
		UVSphereUnwelded(
			discoballRadius,
			cn.Rows.Value(),
			cn.Columns.Value(),
		).Transform(meshops.FlatNormalsTransformer{})

	discoBallHeight := vector3.Up[float64]().Scale(cn.Height.Value())
	discoNormals := discoball.Float3Attribute(modeling.NormalAttribute)

	panelOffset := cn.PanelOffset.Value()

	discoball = discoball.ModifyFloat3Attribute(
		modeling.PositionAttribute,
		func(i int, v vector3.Float64) vector3.Float64 {
			return v.Add(discoNormals.At(i).Scale(panelOffset))
		}).
		Append(primitives.UVSphere(
			discoballRadius+(panelOffset/2),
			cn.Rows.Value(),
			cn.Columns.Value(),
		)).
		Translate(discoBallHeight)

	discoballAttachment := // Base connecting ball to the rod
		primitives.
			Cylinder{
			Sides:  15,
			Height: 0.1,
			Radius: 0.2,
			UVs: &primitives.CylinderUVs{
				Top: &primitives.CircleUVs{
					Center: vector2.New(0.5, 0.5),
					Radius: 0.5,
				},
				Bottom: &primitives.CircleUVs{
					Center: vector2.New(0.5, 0.5),
					Radius: 0.5,
				},
				Side: &primitives.StripUVs{
					Start: vector2.New(0.5, 0.),
					End:   vector2.New(0.5, 1.),
					Width: 0.5,
				},
			},
		}.ToMesh().
			Translate(vector3.New(0., discoballRadius+panelOffset, 0.)).

			// Rod that the ball is hanging from
			Append(primitives.
				Cylinder{
				Sides:  4,
				Height: 3,
				Radius: 0.025,
				UVs: &primitives.CylinderUVs{
					Top: &primitives.CircleUVs{
						Center: vector2.New(0.5, 0.5),
						Radius: 0.5,
					},
					Bottom: &primitives.CircleUVs{
						Center: vector2.New(0.5, 0.5),
						Radius: 0.5,
					},
					Side: &primitives.StripUVs{
						Start: vector2.New(0.5, 0.),
						End:   vector2.New(0.5, 1.),
						Width: 0.5,
					},
				},
			}.ToMesh().
				Translate(vector3.New(0., discoballRadius+panelOffset+1.5, 0.)),
			).
			Translate(discoBallHeight)

	smooth := 0.

	gloss := 1.0
	return []gltf.PolyformModel{
		{
			Name: "Disco Ball",
			Mesh: discoball,
			Material: &gltf.PolyformMaterial{
				Name: "Disco Ball",
				// Extras: map[string]any{
				// 	"threejs-material": "phong",
				// },
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					RoughnessFactor: &smooth,
					BaseColorFactor: ballColor,
					MetallicRoughnessTexture: &gltf.PolyformTexture{
						URI: "metal.png",
					},
				},
				Extensions: []gltf.MaterialExtension{
					&gltf.PolyformPbrSpecularGlossiness{
						GlossinessFactor: &gloss,
						DiffuseFactor:    ballColor,
						SpecularFactor:   color.RGBA{R: 255, G: 255, B: 255, A: 255},
					},
				},
			},
		},
		{
			Name: "Disco Ball Attachment",
			Mesh: discoballAttachment,
			Material: &gltf.PolyformMaterial{
				PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
					RoughnessFactor: &smooth,
					BaseColorFactor: color.Black,
					MetallicRoughnessTexture: &gltf.PolyformTexture{
						URI: "metal.png",
					},
				},
			},
		},
	}, nil
}
