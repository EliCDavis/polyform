package extrude

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func alongX(v vector3.Float64) vector3.Float64 { return vector3.New(v.Z(), v.X(), v.Y()) }

func alongY(v vector3.Float64) vector3.Float64 { return vector3.New(v.X(), v.Z(), v.Y()) }

func alongZ(v vector3.Float64) vector3.Float64 { return v }

// The solid spans x = 0 to depth, which may be negative; a depth of 0 is one face facing +X.
// The region's X runs along Y and its Y along Z.
func RegionX(region triangulation.Region, depth float64) modeling.Mesh {
	return extrudeRegion(region, depth, alongX)
}

// The solid spans y = 0 to depth, which may be negative; a depth of 0 is one face facing +Y.
// The region's X runs along X and its Y along Z.
func RegionY(region triangulation.Region, depth float64) modeling.Mesh {
	return extrudeRegion(region, depth, alongY)
}

// The solid spans z = 0 to depth, which may be negative; a depth of 0 is one face facing +Z.
func RegionZ(region triangulation.Region, depth float64) modeling.Mesh {
	return extrudeRegion(region, depth, alongZ)
}

func extrudeRegion(region triangulation.Region, depth float64, place func(vector3.Float64) vector3.Float64) modeling.Mesh {
	if len(region.Triangles) == 0 {
		return modeling.EmptyMesh(modeling.TriangleTopology)
	}
	at := func(i int, w float64) vector3.Float64 {
		p := region.Points[i]
		return place(vector3.New(p.X(), p.Y(), w))
	}
	bottom, top := min(0, depth), max(0, depth)

	positions := make([]vector3.Float64, 0, len(region.Points)*2)
	normals := make([]vector3.Float64, 0, len(region.Points)*2)
	indices := make([]int, 0, len(region.Triangles)*6)
	front := place(vector3.Forward[float64]())
	for i := range region.Points {
		positions = append(positions, at(i, top))
		normals = append(normals, front)
	}
	for _, tri := range region.Triangles {
		indices = append(indices, tri[0], tri[1], tri[2])
	}

	if top > bottom {
		back := len(positions)
		backward := place(vector3.Backwards[float64]())
		for i := range region.Points {
			positions = append(positions, at(i, bottom))
			normals = append(normals, backward)
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
				outward := place(vector3.New(along.Y(), -along.X(), 0).Normalized())
				wall := len(positions)
				positions = append(positions, at(a, bottom), at(b, bottom), at(b, top), at(a, top))
				normals = append(normals, outward, outward, outward, outward)
				indices = append(indices, wall, wall+1, wall+2, wall, wall+2, wall+3)
			}
		}
	}

	mirrored := place(vector3.Right[float64]()).Cross(place(vector3.Up[float64]())).Dot(place(vector3.Forward[float64]())) < 0
	if mirrored {
		for i := 0; i < len(indices); i += 3 {
			indices[i+1], indices[i+2] = indices[i+2], indices[i+1]
		}
	}

	return modeling.NewTriangleMesh(indices).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: positions,
			modeling.NormalAttribute:   normals,
		})
}

func AreaX(area sketch.Area, depth float64) (modeling.Mesh, error) {
	return extrudeArea(area, depth, RegionX)
}

func AreaY(area sketch.Area, depth float64) (modeling.Mesh, error) {
	return extrudeArea(area, depth, RegionY)
}

func AreaZ(area sketch.Area, depth float64) (modeling.Mesh, error) {
	return extrudeArea(area, depth, RegionZ)
}

func extrudeArea(area sketch.Area, depth float64, extrude func(triangulation.Region, float64) modeling.Mesh) (modeling.Mesh, error) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	if len(area.Polygons) == 0 {
		return empty, nil
	}
	region, err := triangulation.Fill(area.Loops()...)
	if err != nil {
		return empty, err
	}
	return extrude(region, depth), nil
}

type AreaNode struct {
	Area  nodes.LiftedPort[sketch.Area] `description:"The area to extrude."`
	Depth nodes.LiftedPort[float64]     `description:"Thickness: the solid spans 0 to Depth along the output's axis; negative goes the other way. 0 (the default) gives a single face."`
}

func (AreaNode) Description() string {
	return "A flat solid: an area given a thickness."
}

func (AreaNode) Keywords() []string {
	return []string{"slab"}
}

func (n AreaNode) out(out *nodes.Lifted[modeling.Mesh], extrude func(sketch.Area, float64) (modeling.Mesh, error)) {
	nodes.Zip2(out, n.Area, n.Depth, func(area sketch.Area, depth float64) modeling.Mesh {
		mesh, err := extrude(area, depth)
		out.CaptureError(err)
		return mesh
	})
}

func (n AreaNode) X(out *nodes.Lifted[modeling.Mesh]) {
	n.out(out, AreaX)
}

func (AreaNode) XDescription() string {
	return "The area in the YZ plane facing +X: its X along Y, its Y along Z."
}

func (n AreaNode) Y(out *nodes.Lifted[modeling.Mesh]) {
	n.out(out, AreaY)
}

func (AreaNode) YDescription() string {
	return "The area in the XZ plane facing +Y: its X along X, its Y along Z."
}

func (n AreaNode) Z(out *nodes.Lifted[modeling.Mesh]) {
	n.out(out, AreaZ)
}

func (AreaNode) ZDescription() string {
	return "The area in the XY plane facing +Z."
}
