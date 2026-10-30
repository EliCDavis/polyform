package extrude

import (
	"maps"
	"slices"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func alongX(v vector3.Float64) vector3.Float64 { return vector3.New(v.Z(), v.X(), v.Y()) }

func alongY(v vector3.Float64) vector3.Float64 { return vector3.New(v.X(), v.Z(), v.Y()) }

func alongZ(v vector3.Float64) vector3.Float64 { return v }

func RegionX(region triangulation.Region, depth float64, uvs *primitives.StripUVs) modeling.Mesh {
	return extrudeRegion(region, depth, alongX, uvs)
}

func RegionY(region triangulation.Region, depth float64, uvs *primitives.StripUVs) modeling.Mesh {
	return extrudeRegion(region, depth, alongY, uvs)
}

func RegionZ(region triangulation.Region, depth float64, uvs *primitives.StripUVs) modeling.Mesh {
	return extrudeRegion(region, depth, alongZ, uvs)
}

func extrudeRegion(region triangulation.Region, depth float64, place func(vector3.Float64) vector3.Float64, uvs *primitives.StripUVs) modeling.Mesh {
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
	var texcoords []vector2.Float64
	indices := make([]int, 0, len(region.Triangles)*6)
	front := place(vector3.Forward[float64]())
	for i := range region.Points {
		positions = append(positions, at(i, top))
		normals = append(normals, front)
		if uvs != nil {
			texcoords = append(texcoords, uvs.AtXY(region.Points[i]))
		}
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
			if uvs != nil {
				texcoords = append(texcoords, uvs.AtXY(region.Points[i]))
			}
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
		boundary := func(a, b int) bool { return uses[[2]int{min(a, b), max(a, b)}] == 1 }
		across := perimeterDistances(region, boundary)
		for _, tri := range region.Triangles {
			for k := range tri {
				a, b := tri[k], tri[(k+1)%3]
				if !boundary(a, b) {
					continue
				}
				along := region.Points[b].Sub(region.Points[a])
				outward := place(vector3.New(along.Y(), -along.X(), 0).Normalized())
				wall := len(positions)
				positions = append(positions, at(a, bottom), at(b, bottom), at(b, top), at(a, top))
				normals = append(normals, outward, outward, outward, outward)
				indices = append(indices, wall, wall+1, wall+2, wall, wall+2, wall+3)
				if uvs != nil {
					from, to := across[a], across[a]+along.Length()
					texcoords = append(texcoords,
						uvs.AtXY(vector2.New(from, bottom)), uvs.AtXY(vector2.New(to, bottom)),
						uvs.AtXY(vector2.New(to, top)), uvs.AtXY(vector2.New(from, top)))
				}
			}
		}
	}

	mirrored := place(vector3.Right[float64]()).Cross(place(vector3.Up[float64]())).Dot(place(vector3.Forward[float64]())) < 0
	if mirrored {
		for i := 0; i < len(indices); i += 3 {
			indices[i+1], indices[i+2] = indices[i+2], indices[i+1]
		}
	}

	mesh := modeling.NewTriangleMesh(indices).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: positions,
			modeling.NormalAttribute:   normals,
		})
	if uvs == nil {
		return mesh
	}
	return mesh.SetFloat2Attribute(modeling.TexCoordAttribute, texcoords)
}

// How far around its loop each boundary point is, following the triangles' winding.
func perimeterDistances(region triangulation.Region, boundary func(a, b int) bool) map[int]float64 {
	next := map[int]int{}
	for _, tri := range region.Triangles {
		for k := range tri {
			if a, b := tri[k], tri[(k+1)%3]; boundary(a, b) {
				next[a] = b
			}
		}
	}
	across := map[int]float64{}
	for _, start := range slices.Sorted(maps.Keys(next)) {
		if _, walked := across[start]; walked {
			continue
		}
		distance := 0.
		for a := start; ; {
			across[a] = distance
			b := next[a]
			distance += region.Points[b].Sub(region.Points[a]).Length()
			if a = b; a == start {
				break
			}
			if _, walked := across[a]; walked {
				break
			}
		}
	}
	return across
}

func AreaX(area sketch.Area, depth float64, uvs *primitives.StripUVs) (modeling.Mesh, error) {
	return extrudeArea(area, depth, uvs, RegionX)
}

func AreaY(area sketch.Area, depth float64, uvs *primitives.StripUVs) (modeling.Mesh, error) {
	return extrudeArea(area, depth, uvs, RegionY)
}

func AreaZ(area sketch.Area, depth float64, uvs *primitives.StripUVs) (modeling.Mesh, error) {
	return extrudeArea(area, depth, uvs, RegionZ)
}

func extrudeArea(area sketch.Area, depth float64, uvs *primitives.StripUVs, extrude func(triangulation.Region, float64, *primitives.StripUVs) modeling.Mesh) (modeling.Mesh, error) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	if len(area.Polygons) == 0 {
		return empty, nil
	}
	region, err := triangulation.Fill(area.Loops()...)
	if err != nil {
		return empty, err
	}
	return extrude(region, depth, uvs), nil
}

type AreaNode struct {
	Area  nodes.LiftedPort[sketch.Area]     `description:"The area to extrude."`
	Depth nodes.LiftedPort[float64]         `description:"Thickness: the solid spans 0 to Depth along the output's axis; negative goes the other way. 0 (the default) gives a single face."`
	UVs   nodes.Output[primitives.StripUVs] `description:"Lays each wall loop out as a strip in world units: the distance around the loop runs across it, the depth along it. A cap takes the area's own X and Y. Defaults to the unit strip, so a texture repeats once per unit."`
}

func (AreaNode) Description() string {
	return "A flat solid: an area given a thickness."
}

func (AreaNode) Keywords() []string {
	return []string{"slab"}
}

func (n AreaNode) out(out *nodes.Lifted[modeling.Mesh], extrude func(sketch.Area, float64, *primitives.StripUVs) (modeling.Mesh, error)) {
	strip := nodes.TryGetOutputValue(out, n.UVs, primitives.UnitStripUVs())
	nodes.Zip2(out, n.Area, n.Depth, func(area sketch.Area, depth float64) modeling.Mesh {
		mesh, err := extrude(area, depth, &strip)
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
