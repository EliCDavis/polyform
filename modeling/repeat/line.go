package repeat

import (
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Line(start, end vector3.Float64, inbetween int) []trs.TRS {
	return append(
		LineExlusive(start, end, inbetween),
		trs.Position(start),
		trs.Position(end),
	)
}

// Like line, but we don't include transforms on the start and end points. Only
// the inbetween points
func LineExlusive(start, end vector3.Float64, inbetween int) []trs.TRS {

	dir := end.Sub(start)
	inc := dir.DivByConstant(float64(inbetween + 1))

	values := make([]trs.TRS, inbetween)

	for i := 0; i < inbetween; i++ {
		values[i] = trs.Position(start.Add(inc.Scale(float64(i + 1))))
	}

	return values
}

type LineNode = nodes.Struct[[]trs.TRS, LineNodeData]

type LineNodeData struct {
	Start nodes.NodeOutput[vector3.Float64]
	End   nodes.NodeOutput[vector3.Float64]
	Times nodes.NodeOutput[int]
}

func (r LineNodeData) Process() ([]trs.TRS, error) {
	if r.Start == nil || r.End == nil {
		return nil, nil
	}

	times := 0
	if r.Times != nil {
		times = r.Times.Value()
	}

	if times <= 0 {
		return nil, nil
	}

	start := r.Start.Value()
	end := r.End.Value()

	if times == 1 {
		LineExlusive(start, end, 1)
	}

	if times == 2 {
		Line(start, end, 0)
	}

	return Line(start, end, times-2), nil
}
