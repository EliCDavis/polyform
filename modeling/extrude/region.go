package extrude

import (
	"fmt"

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

// Area lays the area facing +Z at z = depth and, given depth, closes it into
// a solid down to z = 0.
func Area(area sketch.Area, depth float64) (modeling.Mesh, error) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	if depth < 0 {
		return empty, fmt.Errorf("depth can not be negative, got %g", depth)
	}
	if len(area.Polygons) == 0 {
		return empty, nil
	}
	region, err := triangulation.Fill(area.Loops()...)
	if err != nil {
		return empty, err
	}
	return Region(region, depth), nil
}

type AreaNode struct {
	Area  nodes.LiftedPort[sketch.Area] `description:"The area to extrude."`
	Depth nodes.LiftedPort[float64]     `description:"Thickness: the solid spans z = 0 to Depth, face at z = Depth. 0 (the default) gives a single face."`
}

func (AreaNode) Description() string {
	return "A flat solid: an area lying in the XY plane facing +Z, given a thickness. A list of areas gives one solid each."
}

func (AreaNode) Keywords() []string {
	return []string{"slab"}
}

func (n AreaNode) Out(out *nodes.Lifted[modeling.Mesh]) {
	nodes.Zip2(out, n.Area, n.Depth, func(area sketch.Area, depth float64) modeling.Mesh {
		mesh, err := Area(area, depth)
		out.CaptureError(err)
		return mesh
	})
}
