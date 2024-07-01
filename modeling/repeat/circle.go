package repeat

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type CirclePointsNode = nodes.StructNode[[]vector3.Float64, CirclePointsNodeData]

type CirclePointsNodeData struct {
	Count  nodes.NodeOutput[int]
	Radius nodes.NodeOutput[float64]
}

func (cpnd CirclePointsNodeData) Process() ([]vector3.Float64, error) {
	count := 0
	radius := 1.

	if cpnd.Count != nil {
		count = cpnd.Count.Value()
	}

	if cpnd.Radius != nil {
		radius = cpnd.Radius.Value()
	}

	return CirclePoints(count, radius), nil
}

func CirclePoints(count int, radius float64) []vector3.Float64 {
	angleIncrement := (1.0 / float64(count)) * 2.0 * math.Pi
	final := make([]vector3.Float64, count)

	for i := 0; i < count; i++ {
		angle := angleIncrement * float64(i)
		final[i] = vector3.New(math.Cos(angle)*radius, 0, math.Sin(angle)*radius)
	}

	return final
}

func Circle(in modeling.Mesh, times int, radius float64) modeling.Mesh {
	angleIncrement := (1.0 / float64(times)) * 2.0 * math.Pi

	final := modeling.EmptyMesh(in.Topology())

	for i := 0; i < times; i++ {
		angle := angleIncrement * float64(i)

		pos := vector3.New(math.Cos(angle), 0, math.Sin(angle)).Scale(radius)
		rot := quaternion.FromTheta(angle-(math.Pi/2), vector3.Down[float64]())

		final = final.Append(
			in.Rotate(rot).
				Transform(
					meshops.TranslateAttribute3DTransformer{
						Amount: pos,
					},
				),
		)
	}

	return final
}

type CircleNode = nodes.StructNode[modeling.Mesh, CircleNodeData]

type CircleNodeData struct {
	Mesh   nodes.NodeOutput[modeling.Mesh]
	Radius nodes.NodeOutput[float64]
	Times  nodes.NodeOutput[int]
}

func (r CircleNodeData) Process() (modeling.Mesh, error) {
	return Circle(r.Mesh.Value(), r.Times.Value(), r.Radius.Value()), nil
}
