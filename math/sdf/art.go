package sdf

import (
	"github.com/EliCDavis/polyform/math/sample"
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
