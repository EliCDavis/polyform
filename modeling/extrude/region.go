package extrude

import (
	"fmt"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// Region lays the region in the XY plane facing +Z and, given depth,
// extrudes it back to z=0 into a closed solid.
func Region(region triangulation.Region, depth float64) modeling.Mesh {
	if len(region.Triangles) == 0 {
		return modeling.EmptyMesh(modeling.TriangleTopology)
	}
	at := func(p vector2.Float64, z float64) vector3.Float64 {
		return vector3.New(p.X(), p.Y(), z)
	}

	positions := make([]vector3.Float64, 0, len(region.Points)*2)
	normals := make([]vector3.Float64, 0, len(region.Points)*2)
	indices := make([]int, 0, len(region.Triangles)*6)
	for _, p := range region.Points {
		positions = append(positions, at(p, depth))
		normals = append(normals, vector3.Forward[float64]())
	}
	for _, tri := range region.Triangles {
		indices = append(indices, tri[0], tri[1], tri[2])
	}

	if depth > 0 {
		back := len(positions)
		for _, p := range region.Points {
			positions = append(positions, at(p, 0))
			normals = append(normals, vector3.Backwards[float64]())
		}
		for _, tri := range region.Triangles {
			indices = append(indices, back+tri[0], back+tri[2], back+tri[1])
		}

		uses := map[[2]int]int{}
		for _, tri := range region.Triangles {
			for k := range tri {
				a, b := tri[k], tri[(k+1)%3]
				uses[[2]int{min(a, b), max(a, b)}]++
			}
		}
		for _, tri := range region.Triangles {
			for k := range tri {
				a, b := tri[k], tri[(k+1)%3]
				if uses[[2]int{min(a, b), max(a, b)}] != 1 {
					continue
				}
				along := region.Points[b].Sub(region.Points[a])
				outward := vector3.New(along.Y(), -along.X(), 0).Normalized()
				wall := len(positions)
				positions = append(positions,
					at(region.Points[a], 0), at(region.Points[b], 0),
					at(region.Points[b], depth), at(region.Points[a], depth),
				)
				normals = append(normals, outward, outward, outward, outward)
				indices = append(indices, wall, wall+1, wall+2, wall, wall+2, wall+3)
			}
		}
	}

	return modeling.NewTriangleMesh(indices).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: positions,
			modeling.NormalAttribute:   normals,
		})
}

type RegionNode struct {
	Outlines []nodes.LiftedPort[[]vector2.Float64] `description:"Closed 2D outlines to fill; overlapping ones merge. Each connection is one outline or a list."`
	Holes    []nodes.LiftedPort[[]vector2.Float64] `description:"Closed 2D outlines to cut out. Each connection is one outline or a list."`
	Depth    nodes.Output[float64]                 `description:"Thickness, extruded toward -Z. 0 (the default) gives a single face."`
}

func (RegionNode) Description() string {
	return "A flat solid: the area inside the outlines minus the holes, lying in the XY plane facing +Z."
}

func (RegionNode) Keywords() []string {
	return []string{"cutout", "polygon", "boolean"}
}

func (n RegionNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	depth := nodes.TryGetOutputValue(out, n.Depth, 0.)
	if depth < 0 {
		out.CaptureError(fmt.Errorf("depth can not be negative, got %g", depth))
		return
	}
	region, err := triangulation.FillDifference(liftedOutlines(out, n.Outlines), liftedOutlines(out, n.Holes))
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(Region(region, depth))
}

func liftedOutlines(recorder nodes.ExecutionRecorder, ports []nodes.LiftedPort[[]vector2.Float64]) []geometry.Shape {
	var outlines []geometry.Shape
	for _, port := range ports {
		switch connected := port.(type) {
		case nodes.Output[[][]vector2.Float64]:
			for _, outline := range nodes.GetOutputValue(recorder, connected) {
				outlines = append(outlines, outline)
			}
		case nodes.Output[[]vector2.Float64]:
			outlines = append(outlines, nodes.GetOutputValue(recorder, connected))
		}
	}
	return outlines
}
