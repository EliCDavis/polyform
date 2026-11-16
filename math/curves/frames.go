package curves

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

// SplinePoints samples a spline evenly, as close to spacing apart as a whole
// number of steps allows. A spline that ends where it began comes back
// closed, without repeating its first point.
func SplinePoints(spline Spline, spacing float64) (points []vector3.Float64, closed bool) {
	if spline == nil {
		return nil, false
	}
	length := spline.Length()
	steps := 1
	if spacing > 0 && length > 0 {
		steps = max(1, int(math.Round(length/spacing)))
	}
	step := length / float64(steps)
	points = make([]vector3.Float64, steps+1)
	for i := range points {
		points[i] = spline.At(step * float64(i))
	}
	if last := len(points) - 1; last > 1 && points[last].Distance(points[0]) <= step*1e-6 {
		return points[:last], true
	}
	return points, false
}

type SplineFramesNode struct {
	Spline  nodes.LiftedPort[Spline]
	Spacing nodes.LiftedPort[float64] `description:"Distance between frames, rounded to fit the spline evenly." default:"0.1"`
}

func (SplineFramesNode) Description() string {
	return "Samples TRSs along a spline."
}

func (SplineFramesNode) AlongDescription() string {
	return "Each frame's fraction of the way along, 0 to 1."
}

func (SplineFramesNode) Keywords() []string {
	return []string{"orient", "sample"}
}

func splineFrames(spline Spline, spacing float64) ([]geometry.PathFrame, bool) {
	points, closed := SplinePoints(spline, spacing)
	return geometry.PathFrames(points, closed), closed
}

func (n SplineFramesNode) Frames(out *nodes.Lifted[[]trs.TRS]) {
	nodes.Zip2(out, n.Spline, n.Spacing, func(spline Spline, spacing float64) []trs.TRS {
		frames, _ := splineFrames(spline, spacing)
		return geometry.FrameTRSs(frames)
	})
}

func (n SplineFramesNode) Along(out *nodes.Lifted[[]float64]) {
	nodes.Zip2(out, n.Spline, n.Spacing, func(spline Spline, spacing float64) []float64 {
		return geometry.Along(splineFrames(spline, spacing))
	})
}

func (n SplineFramesNode) Closed(out *nodes.Lifted[bool]) {
	nodes.Zip2(out, n.Spline, n.Spacing, func(spline Spline, spacing float64) bool {
		_, closed := SplinePoints(spline, spacing)
		return closed
	})
}

func (SplineFramesNode) ClosedDescription() string {
	return "Whether the spline ends where it began."
}
