package main

import (
	"math"
	"math/rand"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func main() {
	n := 1000
	mapSize := 100.
	mapRadius := mapSize / 2
	mapOffset := vector2.New(mapRadius, mapRadius)
	points := make([]vector2.Float64, n)
	for i := 0; i < n; i++ {
		theta := rand.Float64() * 2 * math.Pi
		points[i] = vector2.
			New(math.Cos(theta), math.Sin(theta)).
			Scale(mapRadius * math.Sqrt(rand.Float64())).
			Add(mapOffset)
	}

	uvs := make([]vector2.Float64, 0)

	terrain := triangulation.
		BowyerWatson(points).
		CenterFloat3Attribute(modeling.PositionAttribute).
		ScanFloat3Attribute(modeling.PositionAttribute, func(i int, v vector3.Float64) {
			uvs = append(uvs, v.XZ().Scale(1./mapSize))
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)

	obj.Save("tmp/background/background.obj", terrain)
}
