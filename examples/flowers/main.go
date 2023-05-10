package main

import (
	"math"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector3"
)

func pedal(baseWidth, midWidth, pedalLength, tipLength float64, arcVertCount int) modeling.Mesh {
	halfBaseWidth := baseWidth / 2.
	halfMidWidth := midWidth / 2.

	verts := []vector3.Float64{
		vector3.New(-halfBaseWidth, 0., 0.),
		vector3.New(-halfMidWidth, 0., pedalLength),
	}

	xInx := midWidth / (float64(arcVertCount) + 2.)
	for i := 0; i < arcVertCount; i++ {
		x := -halfMidWidth + (xInx * float64(i+1))
		z := pedalLength + (math.Sin((float64(i+1)/float64(arcVertCount+2))*math.Pi) * tipLength)
		verts = append(verts, vector3.New(x, 0., z))
	}

	verts = append(
		verts,
		vector3.New(halfMidWidth, 0., pedalLength),
		vector3.New(halfBaseWidth, 0., 0.),
	)

	faces := make([]int, 0)
	for i := 1; i < arcVertCount+4; i++ {
		faces = append(faces, 0, i, i+1)
	}

	return modeling.NewMesh(faces).SetFloat3Attribute(modeling.PositionAttribute, verts)
}

func flower(numPedals int) modeling.Mesh {
	// return pedal(0.1, 0.4, 0.4, 0.1, 10)
	return repeat.Circle(pedal(0.1, 0.4, 0.4, 0.1, 10), numPedals, 0.1)
}

func main() {
	gltf.SaveText("tmp/flowers/flowers.gltf", flower(5))
}
