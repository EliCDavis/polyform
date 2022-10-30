package main

import (
	"math"
	"os"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/extrude"
	"github.com/EliCDavis/mesh/obj"
	"github.com/EliCDavis/mesh/primitives"
	"github.com/EliCDavis/mesh/repeat"
	"github.com/EliCDavis/vector"
)

func Chasis(height float64) mesh.Mesh {
	width := 7.
	chasis := primitives.Cylinder(20, height, width)

	rows := 4
	rowSpacing := height / float64(rows+1)
	for i := 1; i <= rows; i++ {
		pos := vector.NewVector3(0, rowSpacing*float64(i)-(height/2.), 0)
		chasis = chasis.
			Append(primitives.Cylinder(20, 0.5, width+.3).Translate(pos))
	}

	column := primitives.Cube().Scale(vector.Vector3Zero(), vector.NewVector3(.2, height, .2))
	columns := repeat.Circle(column, 8, width)
	chasis = chasis.Append(columns)

	return chasis
}

func Legs(height, width float64, numLegs int) mesh.Mesh {
	columnHeight := 1.
	legHeight := height - columnHeight

	leg := primitives.Cube().
		Scale(vector.Vector3Zero(), vector.NewVector3(1, legHeight, 1)).
		Translate(vector.NewVector3(0, -(columnHeight / 2.), 0))

	return primitives.
		Cylinder(20, columnHeight, width).
		Translate(vector.NewVector3(0, (height/2.)-(columnHeight/2.), 0)).
		Append(repeat.Circle(leg, numLegs, width-2.))
}

func Floor(floorHeight, radius float64) mesh.Mesh {
	numLegs := int(math.Round(2*math.Pi*radius) / 4)
	legHeight := 2.
	post := primitives.Cube().
		Scale(vector.Vector3Zero(), vector.NewVector3(.1, legHeight, .1)).
		Translate(vector.NewVector3(0, legHeight/2., 0))

	pathPointCount := numLegs * 2
	angleIncrement := (1.0 / float64(pathPointCount)) * 2.0 * math.Pi
	path := make([]vector.Vector3, pathPointCount)
	for i := 0; i < pathPointCount; i++ {
		angle := angleIncrement * float64(i)
		path[i] = vector.NewVector3(math.Cos(angle)*radius, 0, math.Sin(angle)*radius)
	}
	railing := extrude.ClosedCircle(12, .1, path)

	return primitives.Cylinder(20, floorHeight, radius).
		Append(repeat.Circle(post, numLegs, radius-.2)).
		Append(railing.Translate(vector.Vector3Up().MultByConstant(legHeight))).
		Append(railing.Translate(vector.Vector3Up().MultByConstant(legHeight / 2)))
}

func main() {
	chasisHeight := 20.
	floorHeight := 0.5
	legsHeight := 5.

	offset := legsHeight / 2.
	legs := Legs(legsHeight, 9., 8).
		Translate(vector.NewVector3(0, offset, 0))

	offset += (legsHeight / 2.) + floorHeight/2.
	floor := Floor(floorHeight, 10).
		Translate(vector.NewVector3(0, offset, 0))

	offset += (floorHeight / 2.) + chasisHeight/2.
	chasis := Chasis(chasisHeight).
		Translate(vector.NewVector3(0, offset, 0))

	final := chasis.
		Append(floor).
		Append(legs)

	// sides := 20
	// angleIncrement := (1.0 / float64(sides)) * 2.0 * math.Pi
	// path := make([]vector.Vector3, sides)
	// for i := 0; i < sides; i++ {
	// 	angle := angleIncrement * float64(i)
	// 	path[i] = vector.NewVector3(math.Cos(angle)*10, 0, math.Sin(angle)*10)
	// }
	// width := 2.
	// halfWidth := (width / 2.)
	// topHeight := .5
	// bottomHeight := topHeight / 2.
	// nubSize := halfWidth / 3.
	// final = extrude.Shape(
	// 	[]vector.Vector2{
	// 		vector.NewVector2(-halfWidth, topHeight),
	// 		vector.NewVector2(halfWidth, topHeight),
	// 		vector.NewVector2(halfWidth, 0),

	// 		vector.NewVector2(halfWidth-nubSize, 0),
	// 		vector.NewVector2(halfWidth-nubSize, -bottomHeight),
	// 		vector.NewVector2(halfWidth-nubSize-nubSize, -bottomHeight),
	// 		vector.NewVector2(halfWidth-nubSize-nubSize, 0),

	// 		vector.NewVector2(-halfWidth+nubSize+nubSize, 0),
	// 		vector.NewVector2(-halfWidth+nubSize+nubSize, -bottomHeight),
	// 		vector.NewVector2(-halfWidth+nubSize, -bottomHeight),
	// 		vector.NewVector2(-halfWidth+nubSize, 0),

	// 		vector.NewVector2(-halfWidth, 0),
	// 	}, path)

	f, err := os.Create("fired-heater.obj")
	if err != nil {
		panic(err)
	}

	obj.Write(&final, f)
}
