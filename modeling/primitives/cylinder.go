package primitives

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type TruncatedCone struct {
	Sides           int
	Height          float64
	TopRadius       float64
	BottomRadius    float64
	NoTop, NoBottom bool // Optionally turn off generation of top and/or bottom and turn the cylinder into pipe
	UVs             *CylinderUVs
}

type CylinderUVs struct {
	Top    *CircleUVs
	Bottom *CircleUVs
	Side   *StripUVs
}

func (c TruncatedCone) ToMesh() modeling.Mesh {
	halfHeight := c.Height / 2.

	angleIncrement := (1.0 / float64(c.Sides)) * 2.0 * math.Pi
	vertices := make([]vector3.Float64, (c.Sides*2)+2)
	normals := make([]vector3.Float64, (c.Sides*2)+2)
	for sideIndex := 0; sideIndex <= c.Sides; sideIndex++ {
		angle := angleIncrement * float64(sideIndex)
		vertices[sideIndex*2] = vector3.New(math.Cos(angle)*c.TopRadius, halfHeight, math.Sin(angle)*c.TopRadius)
		vertices[(sideIndex*2)+1] = vector3.New(math.Cos(angle)*c.BottomRadius, -halfHeight, math.Sin(angle)*c.BottomRadius)

		normals[sideIndex*2] = vector3.New(math.Cos(angle), .1, math.Sin(angle)).Normalized()
		normals[(sideIndex*2)+1] = vector3.New(math.Cos(angle), -.1, math.Sin(angle)).Normalized()
	}

	tris := make([]int, 0, c.Sides*2*3)
	for sideIndex := 1; sideIndex <= c.Sides; sideIndex++ {
		topLeft := (sideIndex - 1) * 2
		topRight := (sideIndex) * 2
		bottomLeft := topLeft + 1
		bottomRight := topRight + 1
		tris = append(
			tris,

			bottomLeft,
			topLeft,
			topRight,

			bottomLeft,
			topRight,
			bottomRight,
		)
	}

	top := Circle{
		Sides:  c.Sides,
		Radius: c.TopRadius,
	}
	bottom := Circle{
		Sides:  c.Sides,
		Radius: c.BottomRadius,
	}

	float2Data := make(map[string][]vector2.Float64)

	if c.UVs != nil {
		top.UVs = c.UVs.Top
		bottom.UVs = c.UVs.Bottom

		if c.UVs.Side != nil {
			uvs := make([]vector2.Float64, (c.Sides*2)+2)
			start := c.UVs.Side.Start
			dir := c.UVs.Side.End.Sub(start)

			perpendicular := dir.Perpendicular().Normalized().Scale(c.UVs.Side.Width / 2)
			for sideIndex := 0; sideIndex <= c.Sides; sideIndex++ {
				percent := float64(sideIndex) / float64(c.Sides)
				percentDir := start.Add(dir.Scale(percent))
				uvs[sideIndex*2] = percentDir.Add(perpendicular)
				uvs[(sideIndex*2)+1] = percentDir.Sub(perpendicular)
			}
			float2Data[modeling.TexCoordAttribute] = uvs
		}
	}

	cylinderMesh := modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: vertices,
			modeling.NormalAttribute:   normals,
		}).
		SetFloat2Data(float2Data)

	if !c.NoTop {
		cylinderMesh = cylinderMesh.Append(top.ToMesh().Translate(vector3.New(0, halfHeight, 0)))
	}
	if !c.NoBottom {
		cylinderMesh = cylinderMesh.Append(bottom.ToMesh().
			Transform(
				meshops.RotateAttribute3DTransformer{
					Attribute: modeling.PositionAttribute,
					Amount:    quaternion.FromTheta(math.Pi, vector3.New(1., 0., 0.)),
				},
				meshops.RotateAttribute3DTransformer{
					Attribute: modeling.NormalAttribute,
					Amount:    quaternion.FromTheta(math.Pi, vector3.New(1., 0., 0.)),
				},
			).
			Translate(vector3.New(0, -halfHeight, 0)),
		)
	}

	return cylinderMesh
}

type Cylinder struct {
	Sides           int
	Height          float64
	Radius          float64
	NoTop, NoBottom bool // Optionally turn off generation of top and/or bottom and turn the cylinder into pipe
	UVs             *CylinderUVs
}

func (c Cylinder) ToMesh() modeling.Mesh {
	return TruncatedCone{
		Sides:        c.Sides,
		Height:       c.Height,
		TopRadius:    c.Radius,
		BottomRadius: c.Radius,
		NoTop:        c.NoTop,
		NoBottom:     c.NoBottom,
		UVs:          c.UVs,
	}.ToMesh()
}

type CylinderUVsNode struct {
	Top    nodes.Output[CircleUVs]
	Bottom nodes.Output[CircleUVs]
	Side   nodes.Output[StripUVs]
}

func (n CylinderUVsNode) Out(out *nodes.StructOutput[CylinderUVs]) {
	out.Set(CylinderUVs{
		Top:    nodes.TryGetOutputReference(out, n.Top, nil),
		Bottom: nodes.TryGetOutputReference(out, n.Bottom, nil),
		Side:   nodes.TryGetOutputReference(out, n.Side, nil),
	})
}

type CylinderNode struct {
	Sides   nodes.Output[int]
	Height  nodes.Output[float64]
	Radius  nodes.Output[float64]
	Radius2 nodes.Output[float64]
	Top     nodes.Output[bool]
	Bottom  nodes.Output[bool]
	UVs     nodes.Output[CylinderUVs]
}

func (hnd CylinderNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	bottomRadius := nodes.TryGetOutputValue(out, hnd.Radius, 0.5)
	topRadius := nodes.TryGetOutputValue(out, hnd.Radius2, bottomRadius)

	cone := TruncatedCone{
		TopRadius:    topRadius,
		BottomRadius: bottomRadius,
		Height:       nodes.TryGetOutputValue(out, hnd.Height, 1),
		Sides:        max(nodes.TryGetOutputValue(out, hnd.Sides, 20), 3),
		NoTop:        !nodes.TryGetOutputValue(out, hnd.Top, true),
		NoBottom:     !nodes.TryGetOutputValue(out, hnd.Bottom, true),
		UVs:          nodes.TryGetOutputReference(out, hnd.UVs, nil),
	}
	out.Set(cone.ToMesh())
}
