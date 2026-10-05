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

type ScrewNode struct {
	Line        nodes.Output[[]vector3.Float64] `description:"Profile swept around the Y axis. On the +X side, running up faces outward and running down faces the axis."`
	Segments    nodes.Output[int]               `description:"Steps the sweep is divided into. Defaults to 20."`
	Revolutions nodes.Output[float64]
	Distance    nodes.Output[float64]
	UVs         nodes.Output[primitives.StripUVs]
}

func (snd ScrewNode) Description() string {
	return "Sweeps a profile line around the Y axis. Distance 0 gives a closed round shape; non-zero screws it into a helix like a spring."
}

func (snd ScrewNode) Keywords() []string {
	return []string{"lathe", "revolve"}
}

func (snd ScrewNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if snd.Line == nil {
		return
	}
	line := withoutRepeats(nodes.GetOutputValue(out, snd.Line), false)
	segments := nodes.TryGetOutputValue(out, snd.Segments, 20)
	if len(line) < 2 || segments < 1 {
		return
	}

	turn := math.Pi * 2 * nodes.TryGetOutputValue(out, snd.Revolutions, 1.)
	climb := nodes.TryGetOutputValue(out, snd.Distance, 0.)
	strip := nodes.TryGetOutputValue(out, snd.UVs, primitives.StripUVs{
		Start: vector2.New(0, 0.5),
		End:   vector2.New(1, 0.5),
		Width: 1,
	})
	axis := vector3.Up[float64]()

	// The surface's normal is the way a point moves through the sweep
	// crossed with the way the profile runs through it.
	along := tangents(line)
	across := make([]float64, len(line))
	swing := make([]vector3.Float64, len(line))
	for i, p := range line {
		if i > 0 {
			across[i] = across[i-1] + p.Distance(line[i-1])
		}
		swing[i] = axis.Cross(p).Scale(turn).Add(axis.Scale(climb))
	}
	// A point on the axis moves nowhere, so it faces the way its neighbour does.
	for i := 1; i < len(swing); i++ {
		if swing[i].Length() == 0 {
			swing[i] = swing[i-1]
		}
	}
	for i := len(swing) - 2; i >= 0; i-- {
		if swing[i].Length() == 0 {
			swing[i] = swing[i+1]
		}
	}

	rings := segments + 1
	verts := make([]vector3.Float64, 0, len(line)*rings)
	normals := make([]vector3.Float64, 0, len(line)*rings)
	uvs := make([]vector2.Float64, 0, len(line)*rings)
	for ring := range rings {
		t := float64(ring) / float64(segments)
		q := quaternion.FromTheta(turn*t, axis)
		for i, p := range line {
			normal := swing[i].Cross(along[i])
			if normal.Length() == 0 {
				normal = axis
			}
			verts = append(verts, q.Rotate(p).Add(axis.Scale(climb*t)))
			normals = append(normals, q.Rotate(normal.Normalized()))
			uvs = append(uvs, strip.AtXY(vector2.New(across[i]/across[len(line)-1], t)))
		}
	}

	indices := make([]int, 0, (len(line)-1)*segments*6)
	for ring := 1; ring < rings; ring++ {
		for l := 1; l < len(line); l++ {
			bottomLeft := (l - 1) + ((ring - 1) * len(line))
			bottomRight := l + ((ring - 1) * len(line))
			topLeft := (l - 1) + (ring * len(line))
			topRight := l + (ring * len(line))

			indices = append(
				indices,
				bottomRight, bottomLeft, topLeft,
				topRight, bottomRight, topLeft,
			)
		}
	}

	out.Set(modeling.NewTriangleMesh(indices).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: verts,
			modeling.NormalAttribute:   normals,
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs))
}
