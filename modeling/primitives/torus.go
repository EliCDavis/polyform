package primitives

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Torus struct {
	MajorRadius     float64
	MinorRadius     float64
	MajorResolution int
	MinorResolution int
	UVs             *StripUVs
}

func (c Torus) ToMesh() modeling.Mesh {

	majorAngleIncrement := (1.0 / float64(c.MajorResolution)) * 2.0 * math.Pi
	minorAngleIncrement := (1.0 / float64(c.MinorResolution)) * 2.0 * math.Pi

	verts := make([]vector3.Float64, 0, (c.MinorResolution+1)*(c.MajorResolution+1))
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
		majorUVIncrement := 1.0 / float64(c.MajorResolution+1)
		minorUVIncrement := 1.0 / float64(c.MinorResolution+1)

		uvs := make([]vector2.Float64, 0, c.MinorResolution*c.MajorResolution)
		for majorI := range c.MajorResolution + 1 {
			majorAngle := float64(majorI) * majorUVIncrement

			for minorI := range c.MinorResolution + 1 {
				minorAngle := float64(minorI) * minorUVIncrement
				uvs = append(uvs, vector2.New(majorAngle, minorAngle))
			}
		}

		result = result.SetFloat2Attribute(modeling.TexCoordAttribute, c.UVs.AtXYs(uvs))
	}

	return result
}

type TorusNode struct {
	MajorRadius     nodes.Output[float64]
	MinorRadius     nodes.Output[float64]
	MajorResolution nodes.Output[int]
	MinorResolution nodes.Output[int]
	UVs             nodes.Output[StripUVs]
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
