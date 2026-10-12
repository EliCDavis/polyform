package repeat

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type LineBySpacing struct {
	Line       geometry.Line3D
	MaxSpacing float64
}

func (l LineBySpacing) TRS() []trs.TRS {
	length := l.Line.Length()
	if l.MaxSpacing <= 0 || length == 0 {
		return nil
	}

	// The tolerance keeps 0.9/0.3 (3.0000000000000004) from adding a point.
	count := int(math.Ceil(length/l.MaxSpacing-1e-9)) - 1
	if count <= 0 {
		return nil
	}

	rotation := quaternion.RotationTo(vector3.Forward[float64](), l.Line.Direction())
	transforms := make([]trs.TRS, 0, count)
	for i := range count {
		t := float64(i+1) / float64(count+1)
		transforms = append(transforms, trs.New(l.Line.AtTime(t), rotation, vector3.One[float64]()))
	}
	return transforms
}

type LineBySpacingNode struct {
	Line       nodes.LiftedPort[geometry.Line3D]
	MaxSpacing nodes.LiftedPort[float64] `description:"Largest allowed gap between neighbouring transforms. Gaps are equal and never larger."`
}

func (LineBySpacingNode) Description() string {
	return "Transforms evenly spaced strictly between a line's ends, facing along the line."
}

func (r LineBySpacingNode) Out(out *nodes.Lifted[[]trs.TRS]) {
	nodes.Zip2(out, r.Line, r.MaxSpacing,
		func(line geometry.Line3D, maxSpacing float64) []trs.TRS {
			if maxSpacing <= 0 {
				out.CaptureError(fmt.Errorf("max spacing %v is not positive", maxSpacing))
				return nil
			}
			return LineBySpacing{Line: line, MaxSpacing: maxSpacing}.TRS()
		},
	)
}
