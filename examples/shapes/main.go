package main

import (
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"os"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/obj"
	"github.com/EliCDavis/mesh/triangulation"
	"github.com/EliCDavis/vector"
)

func outline() []triangulation.Constraint {
	return []triangulation.Constraint{
		triangulation.NewConstraint([]vector.Vector2{
			vector.NewVector2(-1000, -1000),
			vector.NewVector2(-1000, 1000),
			vector.NewVector2(1000, 1000),
			vector.NewVector2(1000, -1000),
		}),
	}
}

func main() {
	n := 5000
	mapSize := 3000.
	mapRadius := mapSize / 2
	points := make([]vector.Vector2, n)
	for i := 0; i < n; i++ {
		theta := rand.Float64() * 2 * math.Pi
		points[i] = vector.
			NewVector2(math.Cos(theta), math.Sin(theta)).
			MultByConstant(mapRadius * math.Sqrt(rand.Float64()))
	}

	mat := mesh.Material{
		Name: "Shape",
	}

	terrain := triangulation.
		ConstrainedBowyerWatson(points, outline()).
		SetMaterial(mat)

	objFile, err := os.Create("shape.obj")
	if err != nil {
		panic(err)
	}
	defer objFile.Close()

	mtlFile, err := os.Create("shape.mtl")
	if err != nil {
		panic(err)
	}
	defer mtlFile.Close()

	obj.WriteMesh(&terrain, "shape.mtl", objFile)
	obj.WriteMaterials(&terrain, mtlFile)
}
