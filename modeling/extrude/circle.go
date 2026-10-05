package extrude

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func tube(sides int, path []vector3.Float64, radii []float64, closed bool, uv func(point int, across, along float64) vector2.Float64) (modeling.Mesh, error) {
	if len(path) < 2 {
		return modeling.EmptyMesh(modeling.TriangleTopology), fmt.Errorf("path needs at least 2 points, got %d", len(path))
	}
	if sides < 3 {
		return modeling.EmptyMesh(modeling.TriangleTopology), fmt.Errorf("a tube needs at least 3 sides, got %d", sides)
	}

	profile := make([]vector2.Float64, sides)
	for i := range profile {
		angle := math.Pi * 2 * float64(i) / float64(sides)
		profile[i] = vector2.New(math.Cos(angle), math.Sin(angle))
	}

	return sweep{
		profile:   profile,
		frames:    pathFrames(path, closed),
		closed:    closed,
		scales:    radii,
		smoothing: allSmooth,
		uv:        uv,
	}.shell(), nil
}

// Polygon sweeps a regular polygon through the points, each Thickness its
// radius there. Points may repeat to step the radius in place.
func Polygon(sides int, points []ExtrusionPoint) modeling.Mesh {
	path := make([]vector3.Float64, len(points))
	radii := make([]float64, len(points))
	textured := len(points) > 1
	for i, p := range points {
		path[i] = p.Point
		radii[i] = p.Thickness
		textured = textured && p.UV != nil
	}

	var uv func(point int, across, along float64) vector2.Float64
	if textured {
		uv = func(point int, across, _ float64) vector2.Float64 {
			at := points[point].UV
			dir := points[max(point, 1)].UV.Point.Sub(points[max(point-1, 0)].UV.Point)
			if dir.Length() == 0 {
				return at.Point
			}
			return at.Point.Add(dir.Perpendicular().Normalized().Scale(at.Thickness * (across - 0.5)))
		}
	}

	mesh, err := tube(sides, path, radii, false, uv)
	if err != nil {
		panic(err)
	}
	return mesh
}

func radiiAlong(count int, radius float64, radii []float64) ([]float64, error) {
	if len(radii) == count {
		return radii, nil
	}
	if len(radii) != 0 {
		return nil, fmt.Errorf("got %d radii for %d path points", len(radii), count)
	}
	constant := make([]float64, count)
	for i := range constant {
		constant[i] = radius
	}
	return constant, nil
}

type Circle struct {
	Resolution int
	Radius     float64
	Radii      []float64
	ClosePath  bool
	Path       []vector3.Float64
}

func (c Circle) Extrude() (modeling.Mesh, error) {
	radii, err := radiiAlong(len(c.Path), c.Radius, c.Radii)
	if err != nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), err
	}
	return tube(c.Resolution, c.Path, radii, c.ClosePath, nil)
}

type CircleAlongSpline struct {
	CircleResolution int
	Radius           float64
	Radii            []float64
	ClosePath        bool

	Spline           curves.Spline
	SplineResolution int
	UVs              *primitives.StripUVs
}

func (c CircleAlongSpline) Extrude() (modeling.Mesh, error) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	if c.SplineResolution < 2 {
		return empty, fmt.Errorf("spline resolution needs to be at least 2, got %d", c.SplineResolution)
	}
	radii, err := radiiAlong(c.SplineResolution, c.Radius, c.Radii)
	if err != nil {
		return empty, err
	}

	path := make([]vector3.Float64, c.SplineResolution)
	step := c.Spline.Length() / float64(c.SplineResolution-1)
	for i := range path {
		path[i] = c.Spline.At(step * float64(i))
	}

	// A spline that is itself a loop ends where it began, and closing the
	// path over that would put two rings on the same spot.
	if last := len(path) - 1; c.ClosePath && last > 1 && path[last].Distance(path[0]) < step*1e-6 {
		path, radii = path[:last], radii[:last]
	}

	var uv func(point int, across, along float64) vector2.Float64
	if c.UVs != nil {
		uv = func(_ int, across, along float64) vector2.Float64 {
			return c.UVs.AtXY(vector2.New(across, along))
		}
	}
	return tube(c.CircleResolution, path, radii, c.ClosePath, uv)
}

type CircleNode struct {
	Closed     nodes.Output[bool]
	Resolution nodes.Output[int]
	Radius     nodes.Output[float64]
	Radii      nodes.Output[[]float64]
	Path       nodes.Output[[]vector3.Float64]
}

func (pnd CircleNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if pnd.Path == nil {
		return
	}

	mesh, err := Circle{
		Radius:     nodes.TryGetOutputValue(out, pnd.Radius, 1.0),
		Resolution: max(3, nodes.TryGetOutputValue(out, pnd.Resolution, 3)),
		ClosePath:  nodes.TryGetOutputValue(out, pnd.Closed, false),
		Path:       nodes.GetOutputValue(out, pnd.Path),
		Radii:      nodes.TryGetOutputValue(out, pnd.Radii, nil),
	}.Extrude()
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(mesh)
}

type CircleAlongSplineNode struct {
	Closed           nodes.Output[bool]
	CircleResolution nodes.Output[int]
	Radius           nodes.Output[float64]
	Radii            nodes.Output[[]float64]
	Spline           nodes.Output[curves.Spline]
	SplineResolution nodes.Output[int]
	UVs              nodes.Output[primitives.StripUVs]
}

func (pnd CircleAlongSplineNode) Description() string {
	return "Sweeps a circle along a spline, making a tube. Radii varies the thickness along its length."
}

func (pnd CircleAlongSplineNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if pnd.Spline == nil {
		return
	}
	spline := nodes.GetOutputValue(out, pnd.Spline)
	if spline == nil {
		return
	}

	mesh, err := CircleAlongSpline{
		Radius:           nodes.TryGetOutputValue(out, pnd.Radius, 1.0),
		CircleResolution: max(3, nodes.TryGetOutputValue(out, pnd.CircleResolution, 3)),
		ClosePath:        nodes.TryGetOutputValue(out, pnd.Closed, false),
		Spline:           spline,
		SplineResolution: max(3, nodes.TryGetOutputValue(out, pnd.SplineResolution, 3)),
		Radii:            nodes.TryGetOutputValue(out, pnd.Radii, nil),
		UVs:              nodes.TryGetOutputReference(out, pnd.UVs, nil),
	}.Extrude()
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(mesh)
}
