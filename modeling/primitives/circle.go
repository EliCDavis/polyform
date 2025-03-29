package primitives

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type CircleUVs struct {
	Center vector2.Float64
	Radius float64
}

type Circle struct {
	Sides  int
	Radius float64
	UVs    *CircleUVs
}

func (c Circle) ToMesh() modeling.Mesh {

	angleIncrement := (1.0 / float64(c.Sides)) * 2.0 * math.Pi
	vertices := make([]vector3.Float64, c.Sides+1)
	normals := make([]vector3.Float64, c.Sides+1)

	for sideIndex := 0; sideIndex < c.Sides; sideIndex++ {
		angle := angleIncrement * float64(sideIndex)
		vertices[sideIndex] = vector3.New(math.Cos(angle)*c.Radius, 0, math.Sin(angle)*c.Radius)
		// normals[sideIndex] = vector3.New(math.Cos(angle), .1, math.Sin(angle)).Normalized()
		normals[sideIndex] = vector3.New(0., 1., 0.)
	}

	topMiddleVert := c.Sides
	vertices[topMiddleVert] = vector3.Zero[float64]()
	normals[topMiddleVert] = vector3.New(0., 1., 0.)

	tris := make([]int, 0, c.Sides*3)
	for sideIndex := 1; sideIndex < c.Sides; sideIndex++ {
		topLeft := sideIndex - 1
		topRight := sideIndex
		tris = append(
			tris,

			topLeft,
			topMiddleVert,
			topRight,
		)
	}

	tris = append(
		tris,

		c.Sides-1,
		topMiddleVert,
		0,
	)

	meshV3Data := map[string][]vector3.Float64{
		modeling.PositionAttribute: vertices,
		modeling.NormalAttribute:   normals,
	}

	meshV2Data := map[string][]vector2.Float64{}

	if c.UVs != nil {
		uvs := make([]vector2.Float64, c.Sides+1)
		for sideIndex := 0; sideIndex < c.Sides; sideIndex++ {
			angle := angleIncrement * float64(sideIndex)
			uvs[sideIndex] = vector2.New(math.Cos(angle), math.Sin(angle)).Normalized()
		}
		meshV2Data[modeling.TexCoordAttribute] = uvs
	}

	return modeling.
		NewTriangleMesh(tris).
		SetFloat3Data(meshV3Data).
		SetFloat2Data(meshV2Data)
}

type CircleUVsNode = nodes.Struct[CircleUVsNodeData]

type CircleUVsNodeData struct {
	Center nodes.Output[vector2.Float64]
	Radius nodes.Output[float64]
}

func (c CircleUVsNodeData) Out() nodes.StructOutput[CircleUVs] {
	uvs := CircleUVs{
		Radius: 0.5,
		Center: vector2.New(0.5, 0.5),
	}

	if c.Center != nil {
		uvs.Center = c.Center.Value()
	}

	if c.Radius != nil {
		uvs.Radius = c.Radius.Value()
	}

	return nodes.NewStructOutput(uvs)
}

type CircleNode = nodes.Struct[CircleNodeData]

type CircleNodeData struct {
	Radius nodes.Output[float64]
	Sides  nodes.Output[int]
	UVs    nodes.Output[CircleUVs]
}

func (c CircleNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	circle := Circle{
		Radius: 0.5,
		Sides:  12,
	}

	if c.Sides != nil {
		circle.Sides = c.Sides.Value()
	}

	if c.Radius != nil {
		circle.Radius = c.Radius.Value()
	}

	if c.UVs != nil {
		uvs := c.UVs.Value()
		circle.UVs = &uvs
	}

	return nodes.NewStructOutput(circle.ToMesh())
}
