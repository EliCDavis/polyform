package noise

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/sample"
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
	refutil.RegisterType[nodes.Struct[Perlin3DFieldNode]](factory)

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

type Perlin3DFieldNode struct {
	Amplitude nodes.Output[float64]         `description:"Scales the noise's output range. Defaults to 1."`
	Frequency nodes.Output[vector3.Float64] `description:"Scales input position before sampling. Defaults to (1,1,1)."`
	Shift     nodes.Output[vector3.Float64] `description:"Offset added to position before sampling. Defaults to (0,0,0)."`
}

func (cn Perlin3DFieldNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	amplitude := nodes.TryGetOutputValue(out, cn.Amplitude, 1.)
	frequency := nodes.TryGetOutputValue(out, cn.Frequency, vector3.One[float64]())
	shift := nodes.TryGetOutputValue(out, cn.Shift, vector3.Zero[float64]())

	out.Set(func(v vector3.Float64) float64 {
		return Perlin3D(v.MultByVector(frequency).Add(shift)) * amplitude
	})
}
