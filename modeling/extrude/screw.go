package extrude

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type ScrewNode = nodes.Struct[modeling.Mesh, ScrewNodeData]

type ScrewNodeData struct {
	Line        nodes.NodeOutput[[]vector3.Float64]
	Segments    nodes.NodeOutput[int]
	Revolutions nodes.NodeOutput[float64]
	Distance    nodes.NodeOutput[float64]
	UVs         nodes.NodeOutput[primitives.StripUVs]
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

	segments := nodes.TryGetOutputValue(snd.Segments, 20)

	// 1 or 0 segments leaves us with an edge or nothing
	if segments < 2 {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	revolutions := nodes.TryGetOutputValue(snd.Revolutions, 1.)
	distance := nodes.TryGetOutputValue(snd.Distance, 0.)

	axis := vector3.Up[float64]()
	segmentInc := 1. / float64(segments-1)

	// Create Vertex positions ================================================
	verts := make([]vector3.Float64, 0, len(line)*segments)
	rotInc := math.Pi * 2 * revolutions * segmentInc
	posInc := axis.Scale(distance * segmentInc)
	for seg := 0; seg < segments; seg++ {
		q := quaternion.FromTheta(rotInc*float64(seg), axis)
		for _, v := range line {
			verts = append(verts, q.Rotate(v).Add(posInc.Scale(float64(seg))))
		}
	}

	// Create UVs =============================================================
	lineDist := vector3.Array[float64](line).Distance()
	uvs := make([]vector2.Float64, 0, len(line)*segments)
	var strip primitives.StripUVs

	if snd.UVs != nil {
		strip = snd.UVs.Value()
	} else {
		strip = primitives.StripUVs{
			Start: vector2.New(0, 0.5),
			End:   vector2.New(1, 0.5),
			Width: 1,
		}
	}

	uvDir := strip.Dir()
	uvSegInc := uvDir.Scale(segmentInc)
	uvPerpDirInc := strip.LeftToRight()
	for seg := 0; seg < segments; seg++ {

		centerPoint := strip.Start.Add(uvSegInc.Scale(float64(seg)))

		lineOffset := 0.
		for i, v := range line {
			if i > 0 {
				lineOffset += line[i-1].Distance(v) / lineDist
			}
			uvs = append(uvs, centerPoint.Add(strip.StartLeft().Add(uvPerpDirInc.Scale(lineOffset))))
		}
	}

	// Create Triangles =======================================================
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
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs), nil
}
