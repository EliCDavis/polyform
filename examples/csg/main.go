package main

import (
	"log"
	"math"
	"time"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
)

func main() {
	canvas := marching.NewMarchingCanvas(10)

	start := time.Now()
	sphereCube := marching.
		Sphere(vector3.Zero[float64](), 1.2, 1).
		Modify(
			modeling.PositionAttribute,
			marching.Box(vector3.Zero[float64](), vector3.One[float64]().Scale(2), 1),
			func(a, b sample.Vec3ToFloat) sample.Vec3ToFloat {
				return func(v vector3.Float64) float64 {
					return math.Max(a(v), b(v))
				}
			},
		)

	pipeRadius := 0.5
	pipeStrength := 1.
	pipeLength := .6
	pipes := marching.CombineFields(
		marching.Line(
			vector3.Right[float64]().Scale(pipeLength),
			vector3.Left[float64]().Scale(pipeLength),
			pipeRadius,
			pipeStrength,
		),
		marching.Line(
			vector3.Up[float64]().Scale(pipeLength),
			vector3.Down[float64]().Scale(pipeLength),
			pipeRadius,
			pipeStrength,
		),
		marching.Line(
			vector3.Forward[float64]().Scale(pipeLength),
			vector3.Backwards[float64]().Scale(pipeLength),
			pipeRadius,
			pipeStrength,
		),
	)

	canvas.AddFieldParallel(sphereCube.Modify(
		modeling.PositionAttribute,
		pipes,
		func(a, b sample.Vec3ToFloat) sample.Vec3ToFloat {
			return func(v vector3.Float64) float64 {
				return a(v) * b(v)
			}
		},
	))

	mesh := canvas.MarchParallel(-.0).
		WeldByFloat3Attribute(modeling.PositionAttribute, 3).
		Transform(
			meshops.LaplacianSmoothTransformer{
				Attribute:       modeling.PositionAttribute,
				Iterations:      10,
				SmoothingFactor: .2,
			},
			meshops.SmoothNormalsTransformer{},
		)

	log.Printf("time to compute: %s", time.Now().Sub(start))

	err := obj.Save("tmp/csg/csg.obj", mesh)
	if err != nil {
		panic(err)
	}
}
