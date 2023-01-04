package main

import (
	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/vector"
)

func main() {

	resolution := 100
	cubesPerUnit := 10.
	canvas := marching.NewMarchingCanvas(resolution, resolution, resolution, cubesPerUnit)

	center := vector.Vector3One().MultByConstant((float64(resolution) / cubesPerUnit) / 2.)

	sphereCube := marching.Sphere(center, 1.4, 1).
		Multiply(marching.BoundingBox(modeling.NewAABB(center, vector.Vector3One().MultByConstant(2)), 1))

	pipeStrength := 5.
	pipeRadius := 0.5
	pipes := marching.
		Line(
			center.Sub(vector.Vector3Right().MultByConstant(2)),
			center.Add(vector.Vector3Right().MultByConstant(2)),
			pipeRadius,
			pipeStrength,
		).
		Add(marching.Line(
			center.Sub(vector.Vector3Up().MultByConstant(2)),
			center.Add(vector.Vector3Up().MultByConstant(2)),
			pipeRadius,
			pipeStrength,
		)).
		Add(marching.Line(
			center.Sub(vector.Vector3Forward().MultByConstant(2)),
			center.Add(vector.Vector3Forward().MultByConstant(2)),
			pipeRadius,
			pipeStrength,
		))

	canvas.AddFieldParallel(sphereCube.Sub(pipes))

	mesh := canvas.March(.2).
		CenterFloat3Attribute(modeling.PositionAttribute).
		SmoothLaplacian(10, .2).
		CalculateSmoothNormals()

	err := obj.Save("tmp/csg/csg.obj", mesh)
	if err != nil {
		panic(err)
	}

}
