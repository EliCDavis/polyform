package main

import (
	"math/rand"
	"os"

	"github.com/EliCDavis/mesh/obj"
	"github.com/EliCDavis/mesh/triangulation"
	"github.com/EliCDavis/vector"
)

func main() {
	n := 1000
	pointRange := vector.NewVector2(100, 100)
	points := make([]vector.Vector2, n)
	for i := 0; i < n; i++ {
		points[i] = vector.NewVector2(
			rand.Float64()*pointRange.X(),
			rand.Float64()*pointRange.Y(),
		)
	}

	final := triangulation.BowyerWatson(points)

	objFile, err := os.Create("points.obj")
	if err != nil {
		panic(err)
	}
	defer objFile.Close()

	obj.WriteMesh(&final, "", objFile)
}
