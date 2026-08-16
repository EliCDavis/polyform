package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

// func displacementExample(p vector3.Float64) float64 {
// 	return math.Sin(20*p.X()) * math.Sin(20*p.Y()) * math.Sin(20*p.Z())
// }

func Displace(primitive, displacement sample.Vec3ToFloat, p vector3.Float64) float64 {
	d1 := primitive(p)
	d2 := displacement(p)
	return d1 + d2
}

type DisplaceNode struct {
	Primitive    nodes.Output[sample.Vec3ToFloat] `description:"The base field to perturb."`
	Displacement nodes.Output[sample.Vec3ToFloat] `description:"A field whose value is added to Primitive's distance at every point — positive values push the surface out, negative values pull it in."`
}

func (n DisplaceNode) Description() string {
	return "Perturbs Primitive's surface by adding Displacement's value at every point."
}

func (n DisplaceNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Primitive == nil {
		return
	}

	primitive := nodes.GetOutputValue(out, n.Primitive)

	if n.Displacement == nil {
		out.Set(primitive)
		return
	}

	displacement := nodes.GetOutputValue(out, n.Displacement)
	out.Set(func(v vector3.Float64) float64 {
		return Displace(primitive, displacement, v)
	})
}

// func opTwist(primitive sample.Vec3ToFloat, p vector3.Float64) float64 {
// 	const k = 10.0 // or some other amount
// 	c := math.Cos(k * p.Y())
// 	s := math.Sin(k * p.Y())
// 	m := math.mat2(c, -s, s, c)
// 	q := vector3.New(m*p.xz, p.Y())
// 	return primitive(q)
// }

// func opCheapBend(primitive sample.Vec3ToFloat, p vector3.Float64) float64 {
// 	const k = 10.0 // or some other amount
// 	c := math.Cos(k * p.X())
// 	s := math.Sin(k * p.X())
// 	m := math.mat2(c, -s, s, c)
// 	q := vector3.New(m*p.xy, p.Z())
// 	return primitive(q)
// }
