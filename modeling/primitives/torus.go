package primitives

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type TorusUVs struct {
	MajorOffset float64
	MinorOffset float64
	Strip       StripUVs
}

type Torus struct {
	MajorRadius     float64
	MinorRadius     float64
	MajorResolution int
	MinorResolution int
	UVs             *TorusUVs
}

func (c Torus) ToMesh() modeling.Mesh {

	majorAngleIncrement := (1.0 / float64(c.MajorResolution)) * 2.0 * math.Pi
	minorAngleIncrement := (1.0 / float64(c.MinorResolution)) * 2.0 * math.Pi

	vertCount := (c.MinorResolution + 1) * (c.MajorResolution + 1)
	verts := make([]vector3.Float64, 0, vertCount)
	for majorI := range c.MajorResolution + 1 {
		majorAngle := float64(majorI) * majorAngleIncrement
		majorPoint := vector3.New(math.Cos(majorAngle)*c.MajorRadius, 0, math.Sin(majorAngle)*c.MajorRadius)

		for minorI := range c.MinorResolution + 1 {
			minorAngle := float64(minorI) * minorAngleIncrement
			minorPoint := vector3.New(
				(math.Cos(minorAngle)*c.MinorRadius)*math.Cos(majorAngle),
				math.Sin(minorAngle)*c.MinorRadius,
				(math.Cos(minorAngle)*c.MinorRadius)*math.Sin(majorAngle),
			)

			verts = append(verts, majorPoint.Add(minorPoint))
		}
	}

	indices := make([]int, 0, c.MinorResolution*c.MajorResolution*2)
	for row := 1; row <= c.MajorResolution+1; row++ {
		currentRow := (row * c.MinorResolution) //% len(verts)
		previousRow := (row - 1) * c.MinorResolution

		for i := 0; i < c.MinorResolution; i++ {
			p1 := (i + 1) // % c.MinorResolution
			indices = append(indices,
				previousRow+i, currentRow+p1, currentRow+i,
				previousRow+p1, currentRow+p1, previousRow+i,
			)
		}
	}

	result := modeling.NewTriangleMesh(indices).SetFloat3Attribute(modeling.PositionAttribute, verts)

	if c.UVs != nil {
		majorUVIncrement := 1.0 / float64(c.MajorResolution)
		minorUVIncrement := 1.0 / float64(c.MinorResolution)

		uvs := make([]vector2.Float64, 0, vertCount)
		for majorI := range c.MajorResolution + 1 {
			majorAngle := float64(majorI) * majorUVIncrement

			for minorI := range c.MinorResolution + 1 {
				minorAngle := float64(minorI) * minorUVIncrement
				uvs = append(uvs, vector2.New(
					majorAngle+c.UVs.MajorOffset,
					minorAngle+c.UVs.MinorOffset,
				))
			}
		}

		result = result.SetFloat2Attribute(modeling.TexCoordAttribute, c.UVs.Strip.AtXYs(uvs))
	}

	return result
}

type TorusUVNode struct {
	MajorOffset nodes.Output[float64]
	MinorOffset nodes.Output[float64]
	Strip       nodes.Output[StripUVs]
}

func (c TorusUVNode) Out(out *nodes.StructOutput[TorusUVs]) {
	out.Set(TorusUVs{
		MinorOffset: nodes.TryGetOutputValue(out, c.MinorOffset, 0),
		MajorOffset: nodes.TryGetOutputValue(out, c.MajorOffset, 0),
		Strip: nodes.TryGetOutputValue(out, c.Strip, StripUVs{
			Start: vector2.New(0.5, 0.),
			End:   vector2.New(0.5, 1.),
			Width: 1,
		}),
	})
}

type TorusNode struct {
	MajorRadius     nodes.Output[float64]
	MinorRadius     nodes.Output[float64]
	MajorResolution nodes.Output[int]
	MinorResolution nodes.Output[int]
	UVs             nodes.Output[TorusUVs]
}

func (c TorusNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	circle := Torus{
		MajorRadius:     nodes.TryGetOutputValue(out, c.MajorRadius, .5),
		MinorRadius:     nodes.TryGetOutputValue(out, c.MinorRadius, .25),
		MajorResolution: nodes.TryGetOutputValue(out, c.MajorResolution, 20),
		MinorResolution: nodes.TryGetOutputValue(out, c.MinorResolution, 20),
		UVs:             nodes.TryGetOutputReference(out, c.UVs, nil),
	}
	out.Set(circle.ToMesh())
}
