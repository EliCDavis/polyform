package extrude

import (
	"fmt"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// Region lays the region facing +Z at z = depth and, given depth, closes
// it into a solid down to z = 0.
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
	Outlines  []nodes.Output[[]vector2.Float64] `description:"Closed 2D outlines to fill, one per connection; overlapping ones merge."`
	Areas     []nodes.Output[sketch.Area]       `description:"Areas to fill, one per connection; they merge with the outlines."`
	Holes     []nodes.Output[[]vector2.Float64] `description:"Closed 2D outlines to cut out of everything, one per connection."`
	HoleAreas []nodes.Output[sketch.Area]       `description:"Areas to cut out of everything, one per connection."`
	Depth     nodes.Output[float64]             `description:"Thickness: the solid spans z = 0 to Depth, face at z = Depth. 0 (the default) gives a single face."`
}

func (RegionNode) Description() string {
	return "A flat solid: the outlines and areas minus the holes and hole areas, lying in the XY plane facing +Z."
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
	region, err := fillRegion(
		shapes(out, n.Outlines), nodes.GetOutputValues(out, n.Areas),
		shapes(out, n.Holes), nodes.GetOutputValues(out, n.HoleAreas),
	)
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(Region(region, depth))
}

func fillRegion(outlines []geometry.Shape, areas []sketch.Area, holes []geometry.Shape, holeAreas []sketch.Area) (triangulation.Region, error) {
	if len(areas) == 0 && len(holeAreas) == 0 {
		return triangulation.FillDifference(outlines, holes)
	}

	loopsOf := func(areas []sketch.Area) []geometry.Shape {
		var loops []geometry.Shape
		for _, area := range areas {
			loops = append(loops, area.Loops()...)
		}
		return loops
	}
	groups := [][]geometry.Shape{
		triangulation.CounterClockwise(outlines), loopsOf(areas),
		triangulation.CounterClockwise(holes), loopsOf(holeAreas),
	}
	return triangulation.FillWinding(groups, func(w []int) bool {
		return (w[0] != 0 || w[1] != 0) && w[2] == 0 && w[3] == 0
	})
}

func shapes(recorder nodes.ExecutionRecorder, ports []nodes.Output[[]vector2.Float64]) []geometry.Shape {
	outlines := nodes.GetOutputValues(recorder, ports)
	converted := make([]geometry.Shape, len(outlines))
	for i, outline := range outlines {
		converted[i] = outline
	}
	return converted
}
