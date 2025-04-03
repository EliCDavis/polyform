package main

import (
	"math"
	"math/rand"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/vector/vector2"
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

	terrain := triangulation.
		BowyerWatson(points).
		Transform(
			meshops.CenterAttribute3DTransformer{},
			meshops.CustomTransformer{
				Func: func(m modeling.Mesh) (results modeling.Mesh, err error) {
					pos := m.Float3Attribute(modeling.PositionAttribute)
					uvs := make([]vector2.Float64, pos.Len())
					for i := 0; i < pos.Len(); i++ {
						uvs[i] = pos.At(i).XZ().Scale(1. / mapSize)
					}
					return m.SetFloat2Attribute(modeling.TexCoordAttribute, uvs), nil
				},
			},
		)

	obj.SaveMesh("tmp/background/background.obj", terrain)
}
