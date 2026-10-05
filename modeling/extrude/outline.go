package extrude

import (
	"errors"
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Outline struct {
	Shape  []vector2.Float64
	Path   []vector3.Float64
	Closed bool

	// Outline and path corners turning at least this many degrees keep a
	// hard edge. Nil smooths anything under defaultSmoothingAngle.
	SmoothingAngle *float64

	// Lays the sweep out as a strip: across is the outline's perimeter,
	// along is the path. Nil writes no texture coordinates at all.
	UVs *primitives.StripUVs
}

const defaultSmoothingAngle = 45.

func (o Outline) smoothing() float64 {
	if o.SmoothingAngle == nil {
		return defaultSmoothingAngle * math.Pi / 180
	}
	return *o.SmoothingAngle * math.Pi / 180
}

func (o Outline) Extrude() (modeling.Mesh, error) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)

	shape := withoutRepeats(o.Shape, true)
	if len(shape) < 3 {
		return empty, fmt.Errorf("outline needs at least 3 distinct points, got %d", len(shape))
	}
	path := withoutRepeats(o.Path, o.Closed)
	if len(path) < 2 {
		return empty, fmt.Errorf("path needs at least 2 distinct points, got %d", len(path))
	}
	if geometry.Shape(shape).SignedArea() == 0 {
		return empty, errors.New("outline encloses no area")
	}

	s := sweep{
		profile:   counterClockwise(shape),
		frames:    pathFrames(path, o.Closed),
		closed:    o.Closed,
		smoothing: o.smoothing(),
	}
	if o.UVs != nil {
		s.uv = func(_ int, across, along float64) vector2.Float64 {
			return o.UVs.AtXY(vector2.New(across, along))
		}
	}

	shell := s.shell()
	if o.Closed {
		return shell, nil
	}

	region, err := triangulation.Fill(s.profile)
	if err != nil {
		return empty, err
	}
	capUV := o.capUVs(s.profile)
	return shell.
		Append(s.cap(region, s.frames[0], false, capUV)).
		Append(s.cap(region, s.frames[len(s.frames)-1], true, capUV)), nil
}

// The strip has no room for a cap, so each one takes the whole of it, the
// profile's bounding box stretched over the strip's quad.
func (o Outline) capUVs(shape []vector2.Float64) func(vector2.Float64) vector2.Float64 {
	if o.UVs == nil {
		return nil
	}
	min, max := shape[0], shape[0]
	for _, p := range shape {
		min = vector2.New(math.Min(min.X(), p.X()), math.Min(min.Y(), p.Y()))
		max = vector2.New(math.Max(max.X(), p.X()), math.Max(max.Y(), p.Y()))
	}
	size := max.Sub(min)
	return func(p vector2.Float64) vector2.Float64 {
		unit := p.Sub(min)
		if size.X() > 0 {
			unit = vector2.New(unit.X()/size.X(), unit.Y())
		}
		if size.Y() > 0 {
			unit = vector2.New(unit.X(), unit.Y()/size.Y())
		}
		return o.UVs.AtXY(unit)
	}
}

type OutlineNode struct {
	Outline        nodes.Output[[]vector2.Float64]   `description:"Closed 2D outline to sweep. Winding does not matter. Looking along the path, x is to the right and y is up wherever the path runs level."`
	Path           nodes.Output[[]vector3.Float64]   `description:"Points to sweep through. Sharp turns are mitred, keeping the sweep's width."`
	Closed         nodes.Output[bool]                `description:"Joins the last path point back to the first and leaves the ends uncapped."`
	SmoothingAngle nodes.Output[float64]             `description:"Corners in the outline or the path turning at least this many degrees keep a hard edge. Defaults to 45."`
	UVs            nodes.Output[primitives.StripUVs] `description:"Lays the sweep out as a strip: the outline's perimeter runs across it, the path along it. Each cap takes the whole strip separately. Unconnected, the mesh carries no texture coordinates at all."`
}

func (node OutlineNode) Description() string {
	return "Sweeps a 2D outline along a path, capping each end with a triangulation of the outline."
}

func (node OutlineNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))

	if node.Outline == nil || node.Path == nil {
		return
	}

	smoothing := nodes.TryGetOutputValue(out, node.SmoothingAngle, defaultSmoothingAngle)
	var uvs *primitives.StripUVs
	if node.UVs != nil {
		strip := nodes.GetOutputValue(out, node.UVs)
		uvs = &strip
	}
	mesh, err := Outline{
		Shape:          nodes.GetOutputValue(out, node.Outline),
		Path:           nodes.GetOutputValue(out, node.Path),
		Closed:         nodes.TryGetOutputValue(out, node.Closed, false),
		SmoothingAngle: &smoothing,
		UVs:            uvs,
	}.Extrude()
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(mesh)
}
