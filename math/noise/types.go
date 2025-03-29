package noise

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[Perlin1DNode](factory)
	refutil.RegisterType[Perlin2DNode](factory)
	refutil.RegisterType[Perlin3DNode](factory)

	generator.RegisterTypes(factory)
}

type Perlin1DNode = nodes.Struct[Perlin1DNodeData]

type Perlin1DNodeData struct {
	Time nodes.Output[[]float64]
}

func (cn Perlin1DNodeData) Out() nodes.StructOutput[[]float64] {
	if cn.Time == nil {
		return nodes.NewStructOutput[[]float64](nil)
	}

	times := cn.Time.Value()
	values := make([]float64, len(times))
	for i, t := range times {
		values[i] = Perlin1D(t)
	}

	return nodes.NewStructOutput(values)
}

type Perlin2DNode = nodes.Struct[Perlin2DNodeData]

type Perlin2DNodeData struct {
	Time nodes.Output[[]vector2.Float64]
}

func (cn Perlin2DNodeData) Out() nodes.StructOutput[[]float64] {
	if cn.Time == nil {
		return nodes.NewStructOutput[[]float64](nil)
	}

	times := cn.Time.Value()
	values := make([]float64, len(times))
	for i, t := range times {
		values[i] = Perlin2D(t)
	}

	return nodes.NewStructOutput(values)
}

type Perlin3DNode = nodes.Struct[Perlin3DNodeData]

type Perlin3DNodeData struct {
	Time nodes.Output[[]vector3.Float64]
}

func (cn Perlin3DNodeData) Out() nodes.StructOutput[[]float64] {
	if cn.Time == nil {
		return nodes.NewStructOutput[[]float64](nil)
	}

	times := cn.Time.Value()
	values := make([]float64, len(times))
	for i, t := range times {
		values[i] = Perlin3D(t)
	}

	return nodes.NewStructOutput(values)
}
