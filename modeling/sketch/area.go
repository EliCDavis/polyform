package sketch

import (
	"math"
	"slices"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type Polygon struct {
	Outline []vector2.Float64   `json:"outline"`
	Holes   [][]vector2.Float64 `json:"holes,omitempty"`
}

// Area is a filled region of the plane. Outlines run counter-clockwise,
// holes clockwise, and no two polygons overlap.
type Area struct {
	Polygons []Polygon `json:"polygons"`
}

func (a Area) Loops() []geometry.Shape {
	var loops []geometry.Shape
	for _, polygon := range a.Polygons {
		loops = append(loops, polygon.Outline)
		for _, hole := range polygon.Holes {
			loops = append(loops, hole)
		}
	}
	return loops
}

func (a Area) Contains(p vector2.Float64) bool {
	return slices.ContainsFunc(a.Polygons, func(polygon Polygon) bool {
		return geometry.Shape(polygon.Outline).IsInside(p) &&
			!slices.ContainsFunc(polygon.Holes, func(hole []vector2.Float64) bool { return geometry.Shape(hole).IsInside(p) })
	})
}

// Size is the area enclosed: every outline minus its holes.
func (a Area) Size() float64 {
	total := 0.
	for _, loop := range a.Loops() {
		total += loop.SignedArea()
	}
	return total
}

func (a Area) Transform(translation vector2.Float64, degrees float64, scale vector2.Float64) Area {
	mirrored := scale.X()*scale.Y() < 0
	move := func(loop []vector2.Float64) []vector2.Float64 {
		moved := Transform(loop, translation, degrees, scale)
		if mirrored {
			slices.Reverse(moved)
		}
		return moved
	}

	moved := Area{Polygons: make([]Polygon, len(a.Polygons))}
	for i, polygon := range a.Polygons {
		moved.Polygons[i].Outline = move(polygon.Outline)
		for _, hole := range polygon.Holes {
			moved.Polygons[i].Holes = append(moved.Polygons[i].Holes, move(hole))
		}
	}
	return moved
}

func areaOf(region triangulation.Region) Area {
	var outlines, holes []geometry.Shape
	for _, loop := range region.Boundaries() {
		if loop.SignedArea() > 0 {
			outlines = append(outlines, loop)
		} else {
			holes = append(holes, loop)
		}
	}

	area := Area{Polygons: make([]Polygon, len(outlines))}
	for i, outline := range outlines {
		area.Polygons[i].Outline = outline
	}
	for _, hole := range holes {
		if i := smallestEnclosing(outlines, hole); i >= 0 {
			area.Polygons[i].Holes = append(area.Polygons[i].Holes, hole)
		}
	}
	return area
}

func smallestEnclosing(outlines []geometry.Shape, hole geometry.Shape) int {
	best := -1
	for i, outline := range outlines {
		if outline.SignedArea() <= math.Abs(hole.SignedArea()) {
			continue
		}
		inside := 0
		for _, p := range hole {
			if outline.IsInside(p) {
				inside++
			}
		}
		if 2*inside <= len(hole) {
			continue
		}
		if best < 0 || outline.SignedArea() < outlines[best].SignedArea() {
			best = i
		}
	}
	return best
}

func Fill(outlines, holes []geometry.Shape) (Area, error) {
	region, err := triangulation.FillDifference(outlines, holes)
	if err != nil {
		return Area{}, err
	}
	return areaOf(region), nil
}

func combine(areas []Area, keep func(inside func(i int) bool) bool) (Area, error) {
	groups := make([][]geometry.Shape, len(areas))
	for i, area := range areas {
		groups[i] = area.Loops()
	}
	region, err := triangulation.FillWinding(groups, func(w []int) bool {
		return keep(func(i int) bool { return w[i] != 0 })
	})
	if err != nil {
		return Area{}, err
	}
	return areaOf(region), nil
}

func Union(areas ...Area) (Area, error) {
	return combine(areas, func(inside func(int) bool) bool {
		for i := range areas {
			if inside(i) {
				return true
			}
		}
		return false
	})
}

func Intersect(areas ...Area) (Area, error) {
	if len(areas) == 0 {
		return Area{}, nil
	}
	return combine(areas, func(inside func(int) bool) bool {
		for i := range areas {
			if !inside(i) {
				return false
			}
		}
		return true
	})
}

func Subtract(base Area, remove ...Area) (Area, error) {
	return combine(append([]Area{base}, remove...), func(inside func(int) bool) bool {
		if !inside(0) {
			return false
		}
		for i := range remove {
			if inside(i + 1) {
				return false
			}
		}
		return true
	})
}

type FillNode struct {
	Outlines []nodes.Output[[]vector2.Float64] `description:"Closed outlines to fill, one per connection; overlapping ones merge."`
	Holes    []nodes.Output[[]vector2.Float64] `description:"Closed outlines to cut out, one per connection."`
}

func (FillNode) Description() string {
	return "An area: everything inside any outline and outside every hole."
}

func (n FillNode) Out(out *nodes.StructOutput[Area]) {
	area, err := Fill(toShapes(nodes.GetOutputValues(out, n.Outlines)), toShapes(nodes.GetOutputValues(out, n.Holes)))
	out.CaptureError(err)
	out.Set(area)
}

func toShapes(outlines [][]vector2.Float64) []geometry.Shape {
	shapes := make([]geometry.Shape, len(outlines))
	for i, outline := range outlines {
		shapes[i] = outline
	}
	return shapes
}

type UnionNode struct {
	Areas []nodes.Output[Area] `description:"Areas to merge, one per connection."`
}

func (UnionNode) Description() string {
	return "Merges areas into one, covering everything any of them covers."
}

func (n UnionNode) Out(out *nodes.StructOutput[Area]) {
	area, err := Union(nodes.GetOutputValues(out, n.Areas)...)
	out.CaptureError(err)
	out.Set(area)
}

type IntersectionNode struct {
	Areas []nodes.Output[Area] `description:"Areas to overlap, one per connection."`
}

func (IntersectionNode) Description() string {
	return "The part of the plane every area covers."
}

func (n IntersectionNode) Out(out *nodes.StructOutput[Area]) {
	area, err := Intersect(nodes.GetOutputValues(out, n.Areas)...)
	out.CaptureError(err)
	out.Set(area)
}

type SubtractNode struct {
	Base   nodes.Output[Area]   `description:"The area to cut from."`
	Remove []nodes.Output[Area] `description:"Areas to cut away, one per connection."`
}

func (SubtractNode) Description() string {
	return "Cuts areas out of a base area: holes, notches, the area around other shapes."
}

func (SubtractNode) Keywords() []string {
	return []string{"difference", "cut", "hole", "boolean"}
}

func (n SubtractNode) Out(out *nodes.StructOutput[Area]) {
	if n.Base == nil {
		out.Set(Area{})
		return
	}
	area, err := Subtract(nodes.GetOutputValue(out, n.Base), nodes.GetOutputValues(out, n.Remove)...)
	out.CaptureError(err)
	out.Set(area)
}

type TransformAreaNode struct {
	Area        nodes.LiftedPort[Area]            `description:"The area to move."`
	Translation nodes.LiftedPort[vector2.Float64] `description:"Applied last. Defaults to no move."`
	Rotation    nodes.LiftedPort[float64]         `description:"Degrees counter-clockwise about the origin. Defaults to 0."`
	Scale       nodes.LiftedPort[vector2.Float64] `description:"Applied first, about the origin. A negative axis mirrors it. Defaults to 1 by 1."`
}

func (TransformAreaNode) Description() string {
	return "Scales, rotates about the origin, then moves an area. Lists give one each."
}

func (TransformAreaNode) Keywords() []string {
	return []string{"rotate", "translate", "move", "scale", "place", "mirror"}
}

func (n TransformAreaNode) Out(out *nodes.Lifted[Area]) {
	nodes.Zip4(
		out,
		n.Area,
		nodes.LiftedOr(n.Translation, vector2.Zero[float64]()),
		n.Rotation,
		nodes.LiftedOr(n.Scale, vector2.One[float64]()),
		Area.Transform,
	)
}
