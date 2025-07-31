package repeat

import (
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Grid struct {
	Rows    int
	Columns int
	Width   float64
	Height  float64
}

func (g Grid) TRS() []trs.TRS {
	v2 := g.Vector2()
	output := make([]trs.TRS, len(v2))

	for i, v := range v2 {
		output[i] = trs.Position(vector3.New(v.X(), 0, v.Y()))
	}

	return output
}

func (g Grid) Vector3() []vector3.Float64 {
	v2 := g.Vector2()
	output := make([]vector3.Float64, len(v2))

	for i, v := range v2 {
		output[i] = vector3.New(v.X(), 0, v.Y())
	}

	return output
}

func (g Grid) Vector2() []vector2.Float64 {
	output := make([]vector2.Float64, g.Rows*g.Columns)

	left := -g.Width / 2
	bottom := -g.Height / 2
	bottomLeft := vector2.New(left, bottom)

	widthInc := g.Width / float64(g.Columns-1)
	if g.Columns <= 1 {
		widthInc = 0
		bottomLeft = bottomLeft.SetX(0)
	}

	heightInc := g.Height / float64(g.Rows-1)
	if g.Rows <= 1 {
		heightInc = 0
		bottomLeft = bottomLeft.SetY(0)
	}

	for y := range g.Rows {
		for x := range g.Columns {
			inc := vector2.New(
				widthInc*float64(x),
				heightInc*float64(y),
			)
			output[x+(g.Columns*y)] = bottomLeft.Add(inc)
		}
	}

	return output
}

type GridNode struct {
	Rows    nodes.Output[int]
	Columns nodes.Output[int]
	Width   nodes.Output[float64]
	Height  nodes.Output[float64]
}

func (g GridNode) grid(recorder nodes.ExecutionRecorder) Grid {
	return Grid{
		Rows:    max(nodes.TryGetOutputValue(recorder, g.Rows, 1), 0),
		Columns: max(nodes.TryGetOutputValue(recorder, g.Columns, 1), 0),
		Width:   max(nodes.TryGetOutputValue(recorder, g.Width, 1), 0),
		Height:  max(nodes.TryGetOutputValue(recorder, g.Height, 1), 0),
	}
}

func (g GridNode) TRS() nodes.StructOutput[[]trs.TRS] {
	out := nodes.StructOutput[[]trs.TRS]{}
	out.Set(g.grid(&out).TRS())
	return out
}

func (g GridNode) Vector2() nodes.StructOutput[[]vector2.Float64] {
	out := nodes.StructOutput[[]vector2.Float64]{}
	out.Set(g.grid(&out).Vector2())
	return out
}

func (g GridNode) Vector3() nodes.StructOutput[[]vector3.Float64] {
	out := nodes.StructOutput[[]vector3.Float64]{}
	out.Set(g.grid(&out).Vector3())
	return out
}
