package noise

import (
	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/math/chance"
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

	registry.Register(factory)
}

type Perlin1DNode struct {
	Time      nodes.LiftedPort[float64] `description:"The values to sample the noise at."`
	Shift     nodes.LiftedPort[float64] `description:"Offset added before sampling, for a different slice of the same noise."`
	Amplitude nodes.LiftedPort[float64] `description:"Scales the noise's output range." default:"1"`
	Frequency nodes.LiftedPort[float64] `description:"Scales the input before sampling, so higher means finer noise." default:"1"`
	Seed      nodes.LiftedPort[int]     `description:"Picks a different stretch of noise. The same seed always gives the same noise."`
}

func (cn Perlin1DNode) Description() string {
	return "Perlin noise sampled along one dimension."
}

func (cn Perlin1DNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip5(
		out,
		cn.Time,
		cn.Frequency,
		cn.Shift,
		cn.Amplitude,
		nodes.LiftedOr(cn.Seed, 0),
		func(time, frequency, shift, amplitude float64, seed int) float64 {
			return Perlin1D((time*frequency)+shift+seedShift(seed)) * amplitude
		},
	)
}

// The permutation table repeats every 256 units, so a shift inside that span
// reaches every distinct stretch of the noise.
func seedShift(seed int) float64 {
	return chance.Unit(seed, 0) * 256
}

func seedShift2(seed int) vector2.Float64 {
	next := chance.Roll(seed, 0)
	return vector2.New(next(), next()).Scale(256)
}

func seedShift3(seed int) vector3.Float64 {
	next := chance.Roll(seed, 0)
	return vector3.New(next(), next(), next()).Scale(256)
}

type Perlin2DNode struct {
	Time      nodes.LiftedPort[vector2.Float64] `description:"The points to sample the noise at, despite the name - pass positions here."`
	Amplitude nodes.LiftedPort[float64]         `description:"Scales the noise's output range." default:"1"`
	Frequency nodes.LiftedPort[vector2.Float64] `description:"Scales each point before sampling, so higher means finer noise." default:"1,1"`
	Shift     nodes.LiftedPort[vector2.Float64] `description:"Offset added to each point before sampling, for a different slice of the same noise."`
	Seed      nodes.LiftedPort[int]             `description:"Picks a different patch of noise. The same seed always gives the same noise."`
}

func (cn Perlin2DNode) Description() string {
	return "Perlin noise sampled at 2D points."
}

func (cn Perlin2DNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip5(
		out,
		cn.Time,
		cn.Frequency,
		cn.Shift,
		cn.Amplitude,
		nodes.LiftedOr(cn.Seed, 0),
		func(time, frequency, shift vector2.Float64, amplitude float64, seed int) float64 {
			return Perlin2D(time.MultByVector(frequency).Add(shift).Add(seedShift2(seed))) * amplitude
		},
	)
}

type Perlin3DNode struct {
	Time      nodes.LiftedPort[vector3.Float64] `description:"The points to sample the noise at, despite the name - pass positions here."`
	Amplitude nodes.LiftedPort[float64]         `description:"Scales the noise's output range." default:"1"`
	Frequency nodes.LiftedPort[vector3.Float64] `description:"Scales each point before sampling, so higher means finer noise." default:"1,1,1"`
	Shift     nodes.LiftedPort[vector3.Float64] `description:"Offset added to each point before sampling, for a different slice of the same noise."`
	Seed      nodes.LiftedPort[int]             `description:"Picks a different region of noise. The same seed always gives the same noise."`
}

func (cn Perlin3DNode) Description() string {
	return "Perlin noise sampled at 3D points."
}

func (cn Perlin3DNode) Out(out *nodes.Lifted[float64]) {
	nodes.Zip5(
		out,
		cn.Time,
		cn.Frequency,
		cn.Shift,
		cn.Amplitude,
		nodes.LiftedOr(cn.Seed, 0),
		func(time, frequency, shift vector3.Float64, amplitude float64, seed int) float64 {
			return Perlin3D(time.MultByVector(frequency).Add(shift).Add(seedShift3(seed))) * amplitude
		},
	)
}

type Perlin3DFieldNode struct {
	Amplitude nodes.Output[float64]         `description:"Scales the noise's output range." default:"1"`
	Frequency nodes.Output[vector3.Float64] `description:"Scales input position before sampling." default:"1,1,1"`
	Shift     nodes.Output[vector3.Float64] `description:"Offset added to position before sampling."`
	Seed      nodes.Output[int]             `description:"Picks a different region of noise. The same seed always gives the same noise."`
}

func (cn Perlin3DFieldNode) Description() string {
	return "Perlin noise as a field that can be sampled anywhere in 3D space."
}

func (cn Perlin3DFieldNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	amplitude := nodes.GetOutputValue(out, cn.Amplitude)
	frequency := nodes.GetOutputValue(out, cn.Frequency)
	shift := nodes.GetOutputValue(out, cn.Shift).Add(seedShift3(nodes.GetOutputValue(out, cn.Seed)))

	out.Set(func(v vector3.Float64) float64 {
		return Perlin3D(v.MultByVector(frequency).Add(shift)) * amplitude
	})
}
