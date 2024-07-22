package extrude

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type ScrewNode = nodes.StructNode[modeling.Mesh, ScrewNodeData]

type ScrewNodeData struct {
	Line        nodes.NodeOutput[[]vector3.Float64]
	Segments    nodes.NodeOutput[int]
	Revolutions nodes.NodeOutput[float64]
	Distance    nodes.NodeOutput[float64]
}

func (snd ScrewNodeData) Process() (modeling.Mesh, error) {
	if snd.Line == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}
	line := snd.Line.Value()

	// Can't create a mesh with a single point
	if len(line) < 2 {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	segments := 20
	if snd.Segments != nil {
		segments = snd.Segments.Value()
	}

	// 1 or 0 segments leaves us with an edge or nothing
	if segments < 2 {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	revolutions := 1.
	if snd.Revolutions != nil {
		revolutions = snd.Revolutions.Value()
	}

	distance := 0.
	if snd.Distance != nil {
		distance = snd.Distance.Value()
	}

	axis := vector3.Up[float64]()

	verts := make([]vector3.Float64, 0, len(line)*segments)
	inc := 1. / float64(segments-1)
	rotInc := math.Pi * 2 * revolutions * inc
	posInc := axis.Scale(distance * inc)
	for seg := 0; seg < segments; seg++ {
		q := quaternion.FromTheta(rotInc*float64(seg), axis)
		for _, v := range line {
			verts = append(verts, q.Rotate(v).Add(posInc.Scale(float64(seg))))
		}
	}

	indices := make([]int, 0, len(line)*segments*2)

	for seg := 1; seg < segments; seg++ {
		for l := 1; l < len(line); l++ {
			bottomLeft := (l - 1) + ((seg - 1) * len(line))
			bottomRight := l + ((seg - 1) * len(line))

			topLeft := (l - 1) + (seg * len(line))
			topRight := l + (seg * len(line))

			indices = append(
				indices,
				// topLeft, bottomLeft, bottomRight,
				// topLeft, bottomRight, topRight,
				bottomRight, bottomLeft, topLeft,
				topRight, bottomRight, topLeft,
			)
		}
	}

	return modeling.NewTriangleMesh(indices).
		SetFloat3Attribute(modeling.PositionAttribute, verts), nil
}
