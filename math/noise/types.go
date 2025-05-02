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
	Time      nodes.Output[[]float64]
	Amplitude nodes.Output[float64]
	Frequency nodes.Output[float64]
}

func (cn Perlin1DNodeData) Out() nodes.StructOutput[[]float64] {
	if cn.Time == nil {
		return nodes.NewStructOutput[[]float64](nil)
	}

	scale := nodes.TryGetOutputValue(cn.Amplitude, 1.)
	frequency := nodes.TryGetOutputValue(cn.Frequency, 1.)
	times := cn.Time.Value()
	values := make([]float64, len(times))
	for i, t := range times {
		values[i] = Perlin1D(t*frequency) * scale
	}

	return nodes.NewStructOutput(values)
}

type Perlin2DNode = nodes.Struct[Perlin2DNodeData]

type Perlin2DNodeData struct {
	Time      nodes.Output[[]vector2.Float64]
	Amplitude nodes.Output[float64]
	Frequency nodes.Output[vector2.Float64]
}

func (cn Perlin2DNodeData) Out() nodes.StructOutput[[]float64] {
	if cn.Time == nil {
		return nodes.NewStructOutput[[]float64](nil)
	}

	scale := nodes.TryGetOutputValue(cn.Amplitude, 1.)
	times := cn.Time.Value()
	frequency := nodes.TryGetOutputValue(cn.Frequency, vector2.One[float64]())
	values := make([]float64, len(times))
	for i, t := range times {
		values[i] = Perlin2D(t.MultByVector(frequency)) * scale
	}

	return nodes.NewStructOutput(values)
}

type Perlin3DNode = nodes.Struct[Perlin3DNodeData]

type Perlin3DNodeData struct {
	Time      nodes.Output[[]vector3.Float64]
	Amplitude nodes.Output[float64]
	Frequency nodes.Output[vector3.Float64]
}

func (cn Perlin3DNodeData) Out() nodes.StructOutput[[]float64] {
	if cn.Time == nil {
		return nodes.NewStructOutput[[]float64](nil)
	}

	scale := nodes.TryGetOutputValue(cn.Amplitude, 1.)

	times := cn.Time.Value()
	frequency := nodes.TryGetOutputValue(cn.Frequency, vector3.One[float64]())
	values := make([]float64, len(times))
	for i, t := range times {
		values[i] = Perlin3D(t.MultByVector(frequency)) * scale
	}

	return nodes.NewStructOutput(values)
}
