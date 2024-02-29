package main

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type TableNode struct {
	nodes.StructData[gltf.PolyformModel]

	Color      nodes.NodeOutput[coloring.WebColor]
	Radius     nodes.NodeOutput[float64]
	Height     nodes.NodeOutput[float64]
	Thickness  nodes.NodeOutput[float64]
	Resolution nodes.NodeOutput[int]
}

func (tn *TableNode) Out() nodes.NodeOutput[gltf.PolyformModel] {
	return &nodes.StructNodeOutput[gltf.PolyformModel]{Definition: tn}
}

func (tn TableNode) Process() (gltf.PolyformModel, error) {
	tableHeight := tn.Height.Data()
	return gltf.PolyformModel{
		Name: "Table",
		Mesh: primitives.Cylinder{
			Sides:  tn.Resolution.Data(),
			Height: tn.Thickness.Data(),
			Radius: tn.Radius.Data(),
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
		}.
			ToMesh().
			Translate(vector3.New(0., tableHeight, 0.)).
			Append(primitives.Cylinder{
				Sides:  tn.Resolution.Data(),
				Height: tableHeight,
				Radius: tn.Radius.Data() / 8,
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
			}.ToMesh().Translate(vector3.New(0., tableHeight/2, 0.))),

		Material: &gltf.PolyformMaterial{
			PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
				BaseColorFactor: tn.Color.Data(),
				RoughnessFactor: 1,
				MetallicRoughnessTexture: &gltf.PolyformTexture{
					URI: "rough.png",
				},
			},
		},
	}, nil
}