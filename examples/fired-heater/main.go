package main

import (
	"image/color"
	"math"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Segment struct {
	mesh   modeling.Mesh
	height float64
}

func Chimney(funnelWidth, funnelHeight, taperHeight, shootWidth, shootHeight float64) modeling.Mesh {
	halfTotalHeight := (taperHeight + shootHeight + funnelHeight) / 2.
	path := []vector3.Float64{
		vector3.New(0, -halfTotalHeight, 0),
		vector3.New(0, -halfTotalHeight+funnelHeight, 0),
		vector3.New(.0, -halfTotalHeight+funnelHeight+taperHeight, 0),
		vector3.New(.0, halfTotalHeight, 0),
	}

	rows := 4
	rowSpacing := shootHeight / float64(rows)
	allRows := modeling.EmptyMesh()
	for i := 0; i < rows; i++ {
		pos := vector3.New(0, rowSpacing*float64(i)-halfTotalHeight+funnelHeight+taperHeight, 0)
		allRows = allRows.
			Append(primitives.Cylinder(20, 0.3, shootWidth+.3).Translate(pos))
	}

	widths := []float64{
		funnelWidth,
		funnelWidth,
		shootWidth,
		shootWidth,
	}

	return extrude.CircleWithThickness(20, widths, path).
		Append(allRows).
		Append(primitives.Cylinder(20, 0.3, funnelWidth+.3).
			Translate(vector3.New(0, -halfTotalHeight+funnelHeight, 0)))
}

func Chasis(height, width float64) modeling.Mesh {
	chasis := primitives.Cylinder(20, height, width)

	rows := 4
	rowSpacing := height / float64(rows+1)
	for i := 1; i <= rows; i++ {
		pos := vector3.New(0, rowSpacing*float64(i)-(height/2.), 0)
		chasis = chasis.
			Append(primitives.Cylinder(20, 0.5, width+.3).Translate(pos))
	}

	column := primitives.Cube().Scale(vector3.Zero[float64](), vector3.New(.2, height, .2))
	columns := repeat.Circle(column, 8, width)
	chasis = chasis.Append(columns)

	return chasis
}

func Legs(height, width float64, numLegs int) modeling.Mesh {
	columnHeight := 1.
	legHeight := height - columnHeight

	leg := primitives.Cube().
		Scale(vector3.Zero[float64](), vector3.New(1, legHeight, 1)).
		Translate(vector3.New(0, -(columnHeight / 2.), 0))

	return primitives.
		Cylinder(20, columnHeight, width).
		Translate(vector3.New(0, (height/2.)-(columnHeight/2.), 0)).
		Append(repeat.Circle(leg, numLegs, width-2.))
}

func Floor(floorHeight, radius, walkWidth float64) modeling.Mesh {
	numLegs := int(math.Round(2*math.Pi*radius) / 4)
	legHeight := 2.
	post := primitives.Cube().
		Scale(vector3.Zero[float64](), vector3.New(.1, legHeight, .1)).
		Translate(vector3.New(0, legHeight/2., 0))

	pathPointCount := numLegs * 2
	angleIncrement := (1.0 / float64(pathPointCount)) * 2.0 * math.Pi
	path := make([]vector3.Float64, pathPointCount)
	postRadius := radius + walkWidth - .1
	for i := 0; i < pathPointCount; i++ {
		angle := angleIncrement * float64(i)
		path[i] = vector3.New(math.Cos(angle)*postRadius, 0, math.Sin(angle)*postRadius)
	}
	railing := extrude.ClosedCircleWithConstantThickness(12, .05, path)

	sides := 20
	angleIncrement = (1.0 / float64(sides)) * 2.0 * math.Pi
	shapePath := make([]vector3.Float64, sides)
	offset := radius + (walkWidth / 2)
	for i := 0; i < sides; i++ {
		angle := angleIncrement * float64(i)
		shapePath[i] = vector3.New(math.Cos(angle)*offset, 0, math.Sin(angle)*offset)
	}

	return extrude.ClosedShape(PiShape(floorHeight, walkWidth), shapePath).
		Append(repeat.Circle(post, numLegs, postRadius-.2)).
		Append(railing.Translate(vector3.Up[float64]().MultByConstant(legHeight))).
		Append(railing.Translate(vector3.Up[float64]().MultByConstant(legHeight / 2)))
}

func PiShape(height, width float64) []vector2.Float64 {
	halfWidth := (width / 2.)
	topHeight := height / 2.
	bottomHeight := -topHeight
	nubHeight := bottomHeight - topHeight
	nubSize := halfWidth / 3.

	return []vector2.Float64{
		vector2.New(-halfWidth, topHeight),
		vector2.New(halfWidth, topHeight),
		vector2.New(halfWidth, bottomHeight),

		vector2.New(halfWidth-nubSize, bottomHeight),
		vector2.New(halfWidth-nubSize, nubHeight),
		vector2.New(halfWidth-nubSize-nubSize, nubHeight),
		vector2.New(halfWidth-nubSize-nubSize, bottomHeight),

		vector2.New(-halfWidth+nubSize+nubSize, bottomHeight),
		vector2.New(-halfWidth+nubSize+nubSize, nubHeight),
		vector2.New(-halfWidth+nubSize, nubHeight),
		vector2.New(-halfWidth+nubSize, bottomHeight),

		vector2.New(-halfWidth, bottomHeight),
	}
}

func PutTogetherSegments(segments ...Segment) modeling.Mesh {
	offset := 0.
	finalMesh := modeling.EmptyMesh()
	for _, segment := range segments {
		offset += segment.height / 2
		finalMesh = finalMesh.Append(segment.mesh.Translate(vector3.New(0, offset, 0)))
		offset += segment.height / 2
	}
	return finalMesh
}

func main() {
	chasisHeight := 20.
	chasisWidth := 7.
	floorHeight := 0.5
	legsHeight := 5.

	mat := modeling.Material{
		Name:              "Fired Heater Body",
		DiffuseColor:      color.RGBA{128, 128, 128, 255},
		AmbientColor:      color.RGBA{128, 128, 128, 255},
		SpecularColor:     color.RGBA{128, 128, 128, 255},
		SpecularHighlight: 100,
		OpticalDensity:    1,
	}

	final := PutTogetherSegments(
		Segment{
			mesh:   Legs(legsHeight, 8., 8),
			height: legsHeight,
		},
		Segment{
			mesh:   Floor(floorHeight, chasisWidth, 4),
			height: floorHeight,
		},
		Segment{
			mesh:   Chasis(chasisHeight, chasisWidth),
			height: chasisHeight,
		},
		Segment{
			mesh:   Floor(floorHeight, chasisWidth, 3),
			height: floorHeight,
		},
		Segment{
			mesh:   Chimney(chasisWidth, 4, 5, chasisWidth/6, 10),
			height: 19,
		},
	).SetMaterial(mat)

	err := obj.Save("fired-heater.obj", final)
	if err != nil {
		panic(err)
	}
}
