package main

import (
	"log"
	"math"
	"time"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/vector"
)

func main() {
	canvas := marching.NewMarchingCanvas(10)

	start := time.Now()
	sphereCube := marching.
		Sphere(vector.Vector3Zero(), 1.2, 1).
		Modify(
			modeling.PositionAttribute,
			marching.Box(vector.Vector3Zero(), vector.Vector3One().MultByConstant(2), 1),
			func(a, b sample.Vec3ToFloat) sample.Vec3ToFloat {
				return func(v vector.Vector3) float64 {
					return math.Max(a(v), b(v))
				}
			},
		)

	pipeRadius := 0.5
	pipeStrength := 1.
	pipeLength := .6
	pipes := marching.CombineFields(
		marching.Line(
			vector.Vector3Right().MultByConstant(pipeLength),
			vector.Vector3Left().MultByConstant(pipeLength),
			pipeRadius,
			pipeStrength,
		),
		marching.Line(
			vector.Vector3Up().MultByConstant(pipeLength),
			vector.Vector3Down().MultByConstant(pipeLength),
			pipeRadius,
			pipeStrength,
		),
		marching.Line(
			vector.Vector3Forward().MultByConstant(pipeLength),
			vector.Vector3Backwards().MultByConstant(pipeLength),
			pipeRadius,
			pipeStrength,
		),
	)

	canvas.AddFieldParallel(sphereCube.Modify(
		modeling.PositionAttribute,
		pipes,
		func(a, b sample.Vec3ToFloat) sample.Vec3ToFloat {
			return func(v vector.Vector3) float64 {
				return a(v) * b(v)
			}
		},
	))

	mesh := canvas.MarchParallel(-.0).
		WeldByFloat3Attribute(modeling.PositionAttribute, 3).
		SmoothLaplacian(10, .2).
		CalculateSmoothNormals()

	log.Printf("time to compute: %s", time.Now().Sub(start))

	err := obj.Save("tmp/csg/csg.obj", mesh)
	if err != nil {
		panic(err)
	}
}
