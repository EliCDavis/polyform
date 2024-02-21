package main

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type PlateNode struct {
	nodes.StructData[modeling.Mesh]

	Thickness  nodes.NodeOutput[float64]
	Radius     nodes.NodeOutput[float64]
	Resolution nodes.NodeOutput[int]
}

func (cn *PlateNode) Out() nodes.NodeOutput[modeling.Mesh] {
	return &nodes.StructNodeOutput[modeling.Mesh]{Definition: cn}
}

func (cn PlateNode) Process() (modeling.Mesh, error) {
	return primitives.Cylinder{
		Sides:  cn.Resolution.Data(),
		Height: cn.Thickness.Data(),
		Radius: cn.Radius.Data(),
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
	}.ToMesh(), nil
}
