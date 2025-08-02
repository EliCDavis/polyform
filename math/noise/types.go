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

	refutil.RegisterType[nodes.Struct[Perlin1DNode]](factory)
	refutil.RegisterType[nodes.Struct[Perlin2DNode]](factory)
	refutil.RegisterType[nodes.Struct[Perlin3DNode]](factory)

	generator.RegisterTypes(factory)
}

type Perlin1DNode struct {
	Time      nodes.Output[[]float64]
	Shift     nodes.Output[float64]
	Amplitude nodes.Output[float64]
	Frequency nodes.Output[float64]
}

func (cn Perlin1DNode) Out(out *nodes.StructOutput[[]float64]) {
	if cn.Time == nil {
		return
	}

	scale := nodes.TryGetOutputValue(out, cn.Amplitude, 1.)
	frequency := nodes.TryGetOutputValue(out, cn.Frequency, 1.)
	shift := nodes.TryGetOutputValue(out, cn.Shift, 0.)
	times := nodes.GetOutputValue(out, cn.Time)

	values := make([]float64, len(times))
	for i, t := range times {
		values[i] = Perlin1D((t*frequency)+shift) * scale
	}
	out.Set(values)
}

type Perlin2DNode struct {
	Time      nodes.Output[[]vector2.Float64]
	Amplitude nodes.Output[float64]
	Frequency nodes.Output[vector2.Float64]
	Shift     nodes.Output[vector2.Float64]
}

func (cn Perlin2DNode) Out(out *nodes.StructOutput[[]float64]) {
	if cn.Time == nil {
		return
	}

	times := nodes.GetOutputValue(out, cn.Time)
	scale := nodes.TryGetOutputValue(out, cn.Amplitude, 1.)
	frequency := nodes.TryGetOutputValue(out, cn.Frequency, vector2.One[float64]())
	shift := nodes.TryGetOutputValue(out, cn.Shift, vector2.Zero[float64]())

	values := make([]float64, len(times))
	for i, t := range times {
		values[i] = Perlin2D(t.MultByVector(frequency).Add(shift)) * scale
	}
	out.Set(values)
}

type Perlin3DNode struct {
	Time      nodes.Output[[]vector3.Float64]
	Amplitude nodes.Output[float64]
	Frequency nodes.Output[vector3.Float64]
	Shift     nodes.Output[vector3.Float64]
}

func (cn Perlin3DNode) Out(out *nodes.StructOutput[[]float64]) {
	if cn.Time == nil {
		return
	}

	scale := nodes.TryGetOutputValue(out, cn.Amplitude, 1.)

	times := nodes.GetOutputValue(out, cn.Time)
	frequency := nodes.TryGetOutputValue(out, cn.Frequency, vector3.One[float64]())
	shift := nodes.TryGetOutputValue(out, cn.Shift, vector3.Zero[float64]())
	values := make([]float64, len(times))
	for i, t := range times {
		values[i] = Perlin3D(t.MultByVector(frequency).Add(shift)) * scale
	}
	out.Set(values)
}
