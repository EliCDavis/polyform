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
	Time      nodes.LiftedPort[float64] `description:"The values to sample the noise at."`
	Shift     nodes.LiftedPort[float64] `description:"Offset added before sampling, for a different slice of the same noise. Defaults to 0."`
	Amplitude nodes.LiftedPort[float64] `description:"Scales the noise's output range. Defaults to 1."`
	Frequency nodes.LiftedPort[float64] `description:"Scales the input before sampling, so higher means finer noise. Defaults to 1."`
}

func (cn Perlin1DNode) Description() string {
	return "Perlin noise sampled along one dimension."
}

func (cn Perlin1DNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip4(
		out,
		nodes.LiftedOr(cn.Time, 0.),
		nodes.LiftedOr(cn.Frequency, 1.),
		nodes.LiftedOr(cn.Shift, 0.),
		nodes.LiftedOr(cn.Amplitude, 1.),
		func(time, frequency, shift, amplitude float64) float64 {
			return Perlin1D((time*frequency)+shift) * amplitude
		},
	)
}

type Perlin2DNode struct {
	Time      nodes.LiftedPort[vector2.Float64] `description:"The points to sample the noise at, despite the name - pass positions here."`
	Amplitude nodes.LiftedPort[float64]         `description:"Scales the noise's output range. Defaults to 1."`
	Frequency nodes.LiftedPort[vector2.Float64] `description:"Scales each point before sampling, so higher means finer noise. Defaults to (1,1)."`
	Shift     nodes.LiftedPort[vector2.Float64] `description:"Offset added to each point before sampling, for a different slice of the same noise. Defaults to (0,0)."`
}

func (cn Perlin2DNode) Description() string {
	return "Perlin noise sampled at 2D points."
}

func (cn Perlin2DNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip4(
		out,
		nodes.LiftedOr(cn.Time, vector2.Zero[float64]()),
		nodes.LiftedOr(cn.Frequency, vector2.One[float64]()),
		nodes.LiftedOr(cn.Shift, vector2.Zero[float64]()),
		nodes.LiftedOr(cn.Amplitude, 1.),
		func(time, frequency, shift vector2.Float64, amplitude float64) float64 {
			return Perlin2D(time.MultByVector(frequency).Add(shift)) * amplitude
		},
	)
}

type Perlin3DNode struct {
	Time      nodes.LiftedPort[vector3.Float64] `description:"The points to sample the noise at, despite the name - pass positions here."`
	Amplitude nodes.LiftedPort[float64]         `description:"Scales the noise's output range. Defaults to 1."`
	Frequency nodes.LiftedPort[vector3.Float64] `description:"Scales each point before sampling, so higher means finer noise. Defaults to (1,1,1)."`
	Shift     nodes.LiftedPort[vector3.Float64] `description:"Offset added to each point before sampling, for a different slice of the same noise. Defaults to (0,0,0)."`
}

func (cn Perlin3DNode) Description() string {
	return "Perlin noise sampled at 3D points."
}

func (cn Perlin3DNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip4(
		out,
		nodes.LiftedOr(cn.Time, vector3.Zero[float64]()),
		nodes.LiftedOr(cn.Frequency, vector3.One[float64]()),
		nodes.LiftedOr(cn.Shift, vector3.Zero[float64]()),
		nodes.LiftedOr(cn.Amplitude, 1.),
		func(time, frequency, shift vector3.Float64, amplitude float64) float64 {
			return Perlin3D(time.MultByVector(frequency).Add(shift)) * amplitude
		},
	)
}

type Perlin3DFieldNode struct {
	Amplitude nodes.Output[float64]         `description:"Scales the noise's output range. Defaults to 1."`
	Frequency nodes.Output[vector3.Float64] `description:"Scales input position before sampling. Defaults to (1,1,1)."`
	Shift     nodes.Output[vector3.Float64] `description:"Offset added to position before sampling. Defaults to (0,0,0)."`
}

func (cn Perlin3DFieldNode) Description() string {
	return "Perlin noise as a field that can be sampled anywhere in 3D space."
}

func (cn Perlin3DFieldNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	amplitude := nodes.TryGetOutputValue(out, cn.Amplitude, 1.)
	frequency := nodes.TryGetOutputValue(out, cn.Frequency, vector3.One[float64]())
	shift := nodes.TryGetOutputValue(out, cn.Shift, vector3.Zero[float64]())

	out.Set(func(v vector3.Float64) float64 {
		return Perlin3D(v.MultByVector(frequency).Add(shift)) * amplitude
	})
}
