package extrude

import (
	"errors"
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type AreaAlongPath struct {
	Profile sketch.Area
	Frames  []trs.TRS
	Closed  bool

	// Outline and path corners turning at least this many degrees keep a
	// hard edge. Nil smooths anything under defaultSmoothingAngle.
	SmoothingAngle *float64

	// Each loop of the profile runs across the strip and the path along it.
	// Nil writes no texture coordinates.
	UVs *primitives.StripUVs
}

func (a AreaAlongPath) smoothing() float64 {
	if a.SmoothingAngle == nil {
		return defaultSmoothingAngle * math.Pi / 180
	}
	return *a.SmoothingAngle * math.Pi / 180
}

func (a AreaAlongPath) Extrude() (modeling.Mesh, error) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	if len(a.Profile.Polygons) == 0 {
		return empty, nil
	}
	if len(a.Frames) < 2 {
		return empty, fmt.Errorf("path needs at least 2 frames, got %d", len(a.Frames))
	}
	if a.Profile.Size() == 0 {
		return empty, errors.New("profile encloses no area")
	}

	frames := framesFromTRS(a.Frames, a.Closed)
	var uv func(int, float64, float64) vector2.Float64
	if a.UVs != nil {
		uv = func(_ int, across, along float64) vector2.Float64 {
			return a.UVs.AtXY(vector2.New(across, along))
		}
	}

	loops := a.Profile.Loops()
	mesh := empty
	var outline []vector2.Float64
	for _, loop := range loops {
		shape := withoutRepeats(loop, true)
		if len(shape) < 3 {
			continue
		}
		outline = append(outline, shape...)
		mesh = mesh.Append(sweep{
			profile:   shape,
			frames:    frames,
			closed:    a.Closed,
			smoothing: a.smoothing(),
			uv:        uv,
		}.shell())
	}
	if a.Closed {
		return mesh, nil
	}

	region, err := triangulation.Fill(loops...)
	if err != nil {
		return empty, err
	}
	capUV := Outline{UVs: a.UVs}.capUVs(outline)
	ends := sweep{frames: frames}
	return mesh.
		Append(ends.cap(region, frames[0], false, capUV)).
		Append(ends.cap(region, frames[len(frames)-1], true, capUV)), nil
}

type AreaAlongPathNode struct {
	Profile        nodes.LiftedPort[sketch.Area]     `description:"Looking along the path, x is to the right and y is up."`
	Frames         nodes.LiftedPort[[]trs.TRS]       `description:"Where each cross-section sits, faces and how big it is."`
	Closed         nodes.Output[bool]                `description:"Joins the last frame to the first; the ends stay uncapped."`
	SmoothingAngle nodes.Output[float64]             `description:"Degrees a profile or path corner turns to keep a hard edge." default:"45"`
	UVs            nodes.Output[primitives.StripUVs] `description:"Around each loop runs across the strip, the path along it."`
}

func (AreaAlongPathNode) Description() string {
	return "Sweeps an area through a list of transforms, each placing one cross-section with its position, rotation and scale."
}

func (AreaAlongPathNode) Keywords() []string {
	return []string{"sweep", "tube"}
}

func (n AreaAlongPathNode) Out(out *nodes.Lifted[modeling.Mesh]) {
	smoothing := nodes.GetOutputValue(out, n.SmoothingAngle)
	strip := nodes.TryGetOutputValue(out, n.UVs, primitives.UnitStripUVs())
	closed := nodes.GetOutputValue(out, n.Closed)
	nodes.Zip2(out, n.Profile, n.Frames, func(profile sketch.Area, frames []trs.TRS) modeling.Mesh {
		if n.Frames == nil {
			return modeling.EmptyMesh(modeling.TriangleTopology)
		}
		mesh, err := AreaAlongPath{
			Profile:        profile,
			Frames:         frames,
			Closed:         closed,
			SmoothingAngle: &smoothing,
			UVs:            &strip,
		}.Extrude()
		out.CaptureError(err)
		return mesh
	})
}
