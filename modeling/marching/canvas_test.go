package marching_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/vector/vector3"
)

var meshResult modeling.Mesh

func BenchmarkSphere(b *testing.B) {
	cubesPerUnit := 100.
	var mesh modeling.Mesh
	for n := 0; n < b.N; n++ {
		canvas := marching.NewMarchingCanvas(cubesPerUnit)

		canvas.AddField(
			marching.Sphere(vector3.Zero[float64](), 2., 1.),
		)

		mesh = canvas.March(0)
	}

	meshResult = mesh
}

func BenchmarkMarchSphereParallel(b *testing.B) {
	cubesPerUnit := 10.
	canvas := marching.NewMarchingCanvas(cubesPerUnit)

	canvas.AddField(
		marching.Sphere(vector3.Zero[float64](), 2., 1.),
	)

	var mesh modeling.Mesh
	for n := 0; n < b.N; n++ {
		mesh = canvas.MarchParallel(0)
	}

	meshResult = mesh
}

func BenchmarkAddField_Sphere(b *testing.B) {
	cubesPerUnit := 100.
	canvas := marching.NewMarchingCanvas(cubesPerUnit)
	field := marching.Sphere(vector3.Zero[float64](), 2., 1.)
	for n := 0; n < b.N; n++ {
		canvas.AddField(field)
	}

	meshResult = canvas.March(0)
}

func BenchmarkAddFieldParallel_Sphere(b *testing.B) {
	cubesPerUnit := 100.
	canvas := marching.NewMarchingCanvas(cubesPerUnit)
	field := marching.Sphere(vector3.Zero[float64](), 2., 1.)
	for n := 0; n < b.N; n++ {
		canvas.AddFieldParallel(field)
	}

	meshResult = canvas.March(0)
}

func BenchmarkAddFieldParallel2_Sphere(b *testing.B) {
	cubesPerUnit := 100.
	canvas := marching.NewMarchingCanvas(cubesPerUnit)
	field := marching.Sphere(vector3.Zero[float64](), 2., 1.)
	for n := 0; n < b.N; n++ {
		canvas.AddFieldParallel2(field)
	}

	meshResult = canvas.March(0)
}
