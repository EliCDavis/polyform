package repeat

import (
	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Spline(curve curves.Spline, inbetween int) []trs.TRS {
	start := trs.New(
		curve.At(0),
		quaternion.RotationTo(vector3.Forward[float64](), curve.Dir(0)),
		vector3.One[float64](),
	)

	dist := curve.Length()
	end := trs.New(
		curve.At(dist),
		quaternion.RotationTo(vector3.Forward[float64](), curve.Dir(dist)),
		vector3.One[float64](),
	)

	result := make([]trs.TRS, 0, inbetween+2)
	result = append(result, start)
	result = append(result, SplineExlusive(curve, inbetween)...)
	result = append(result, end)

	return result
}

// Like line, but we don't include meshes on the start and end points. Only the
// inbetween points
func SplineExlusive(curve curves.Spline, inbetween int) []trs.TRS {

	inc := curve.Length() / float64(inbetween+1)

	transforms := make([]trs.TRS, inbetween)

	for i := 0; i < inbetween; i++ {
		dist := inc * float64(i+1)
		dir := curve.Dir(dist)

		transforms[i] = trs.New(
			curve.At(dist),
			quaternion.RotationTo(vector3.Forward[float64](), dir),
			vector3.One[float64](),
		)
	}

	return transforms
}

type SplineNode = nodes.Struct[SplineNodeData]

type SplineNodeData struct {
	Curve nodes.Output[curves.Spline]
	Times nodes.Output[int]
}

func (rnd SplineNodeData) Description() string {
	return "Creates an array of TRS matrices by sampling the curve"
}

func (r SplineNodeData) Out() nodes.StructOutput[[]trs.TRS] {
	if r.Curve == nil || r.Times == nil {
		return nodes.NewStructOutput[[]trs.TRS](nil)
	}

	times := r.Times.Value()
	if times <= 0 {
		return nodes.NewStructOutput[[]trs.TRS](nil)
	}

	curve := r.Curve.Value()
	if curve == nil {
		return nodes.NewStructOutput[[]trs.TRS](nil)
	}

	if times == 1 {
		return nodes.NewStructOutput(SplineExlusive(curve, 1))
	}

	if times == 2 {
		return nodes.NewStructOutput(Spline(curve, 0))
	}

	return nodes.NewStructOutput(Spline(curve, times-2))
}
