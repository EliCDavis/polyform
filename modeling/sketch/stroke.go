package sketch

import (
	"fmt"
	"math"
	"slices"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

func Stroke(path []vector2.Float64, width float64, segments int, closed bool) (Area, error) {
	if width <= 0 {
		return Area{}, fmt.Errorf("width must be positive, got %g", width)
	}
	if len(path) == 0 {
		return Area{}, nil
	}
	segments = max(segments, 1)
	if closed && len(path) > 2 {
		path = append(slices.Clone(path), path[0])
	}

	capsules := make([]geometry.Shape, 0, len(path))
	for i := 0; i+1 < len(path); i++ {
		capsules = append(capsules, capsule(path[i], path[i+1], width/2, segments))
	}
	if len(path) == 1 {
		capsules = append(capsules, capsule(path[0], path[0], width/2, segments))
	}

	region, err := triangulation.FillDifference(capsules, nil)
	if err != nil {
		return Area{}, err
	}
	return areaOf(region), nil
}

func capsule(a, b vector2.Float64, radius float64, segments int) geometry.Shape {
	along := b.Sub(a)
	start := math.Atan2(along.Y(), along.X()) + math.Pi/2
	shape := make(geometry.Shape, 0, 2*(segments+1))
	for k, end := range []vector2.Float64{a, b} {
		for j := 0; j <= segments; j++ {
			angle := start + math.Pi*(float64(k)+float64(j)/float64(segments))
			shape = append(shape, end.Add(vector2.New(math.Cos(angle), math.Sin(angle)).Scale(radius)))
		}
	}
	return shape
}

type StrokeNode struct {
	Path     nodes.LiftedPort[[]vector2.Float64] `description:"Points to run through, in order."`
	Width    nodes.LiftedPort[float64]           `description:"Full width of the stroke. Defaults to 1."`
	Segments nodes.Output[int]                   `description:"Straight segments per half circle at the ends and bends. Defaults to 8."`
	Closed   nodes.Output[bool]                  `description:"Joins the last point back to the first, so the stroke is a ring. Defaults to false."`
}

func (StrokeNode) Description() string {
	return "The area a path covers drawn at a width, with round ends and bends: a strip along a path. A gap the path encloses stays open. A list of paths gives one area each."
}

func (StrokeNode) Keywords() []string {
	return []string{"polyline", "thicken"}
}

func (n StrokeNode) Out(out *nodes.Lifted[Area]) {
	segments := max(nodes.TryGetOutputValue(out, n.Segments, 8), 1)
	closed := nodes.TryGetOutputValue(out, n.Closed, false)
	nodes.Zip2(out, n.Path, nodes.LiftedOr(n.Width, 1.), func(path []vector2.Float64, width float64) Area {
		area, err := Stroke(path, width, segments, closed)
		out.CaptureError(err)
		return area
	})
}
