package main

import (
	"image/color"
	"math"
	"os"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/mesh/extrude"
	"github.com/EliCDavis/mesh/obj"
	"github.com/EliCDavis/mesh/primitives"
	"github.com/EliCDavis/mesh/repeat"
	"github.com/EliCDavis/vector"
)

func AbductionRing(radius, baseThickness, magnitude float64) mesh.Mesh {
	pathSize := 120
	path := make([]vector.Vector3, pathSize)
	thickness := make([]float64, pathSize)

	angleIncrement := (1.0 / float64(pathSize)) * 2.0 * math.Pi
	for i := 0; i < pathSize; i++ {
		angle := angleIncrement * float64(i)
		path[i] = vector.NewVector3(math.Cos(angle)*radius, math.Sin(angle*5)*magnitude, math.Sin(angle)*radius)
		thickness[i] = (math.Sin(angle*8) * magnitude * 0.25) + baseThickness
	}

	mat := mesh.Material{
		Name:              "Abduction Ring",
		DiffuseColor:      color.RGBA{0, 255, 0, 255},
		AmbientColor:      color.RGBA{0, 255, 0, 255},
		SpecularColor:     color.Black,
		SpecularHighlight: 0,
		Dissolve:          1,
		OpticalDensity:    5,
	}
	return extrude.
		ClosedCircleWithThickness(20, thickness, path).
		SetMaterial(mat)
}

func contour(positions []vector.Vector3, times int) mesh.Mesh {
	return repeat.Circle(extrude.Circle(7, .3, positions), times, 0)
}

func sideLights(numberOfLights int, radius float64) mesh.Mesh {
	sides := 8
	light := primitives.Cylinder(sides, 0.5, 0.5).
		Append(primitives.Cylinder(sides, 0.25, 0.25).Translate(vector.NewVector3(0, .35, 0))).
		Rotate(mesh.UnitQuaternionFromTheta(-math.Pi/2, vector.Vector3Forward()))

	return repeat.Circle(light, numberOfLights, radius)
}

func UfoBody(outerRadius float64, portalRadius float64, frameSections int) mesh.Mesh {
	path := []vector.Vector3{
		vector.Vector3Up().MultByConstant(-1),
		vector.Vector3Up().MultByConstant(2),

		vector.Vector3Up().MultByConstant(0.5),
		vector.Vector3Up().MultByConstant(3),
		vector.Vector3Up().MultByConstant(4),
		vector.Vector3Up().MultByConstant(5),

		vector.Vector3Up().MultByConstant(5.5),
		vector.Vector3Up().MultByConstant(5.5),
	}
	thickness := []float64{
		0,
		portalRadius - 1,

		portalRadius,
		outerRadius,
		outerRadius,
		portalRadius,

		portalRadius,
		portalRadius - 1,
	}

	domeResolution := 10
	domeHight := 2
	domeStartHeight := 5.5
	domeStartWidth := portalRadius - 1
	halfPi := math.Pi / 2.
	domePath := make([]vector.Vector3, 0)
	domePath = append(domePath, path[len(path)-1])
	domeThickness := make([]float64, 0)
	domeThickness = append(domeThickness, thickness[len(thickness)-1])
	for i := 0; i < domeResolution; i++ {
		percent := float64(i+1) / float64(domeResolution)

		height := math.Sin(percent*halfPi) * float64(domeHight)
		domePath = append(domePath, vector.Vector3Up().MultByConstant(height+domeStartHeight))

		cosResult := math.Cos(percent * halfPi)
		domeThickness = append(domeThickness, (cosResult * domeStartWidth))
	}

	mat := mesh.Material{
		Name:              "UFO Body",
		DiffuseColor:      color.RGBA{128, 128, 128, 255},
		AmbientColor:      color.RGBA{128, 128, 128, 255},
		SpecularColor:     color.RGBA{128, 128, 128, 255},
		SpecularHighlight: 100,
		Dissolve:          1,
		OpticalDensity:    1,
	}

	domeMat := mesh.Material{
		Name:              "UFO Dome",
		DiffuseColor:      color.RGBA{0, 0, 255, 255},
		AmbientColor:      color.RGBA{0, 0, 255, 255},
		SpecularColor:     color.RGBA{0, 0, 255, 255},
		SpecularHighlight: 100,
		Dissolve:          0.8,
		OpticalDensity:    2,
	}

	return extrude.CircleWithThickness(20, thickness, path).
		Append(contour([]vector.Vector3{
			vector.NewVector3(thickness[2], path[2].Y(), 0),
			vector.NewVector3(thickness[3], path[3].Y(), 0),
			vector.NewVector3(thickness[4], path[4].Y(), 0),
			vector.NewVector3(thickness[5], path[5].Y(), 0),
		}, frameSections)).
		Append(primitives.
			Cylinder(20, 1, outerRadius+1).
			Translate(vector.NewVector3(0, 3.5, 0))).
		Append(extrude.ClosedCircle(8, .25, repeat.Point(frameSections, portalRadius)).
			Translate(vector.Vector3Up().MultByConstant(0.5))).
		Append(sideLights(frameSections, outerRadius+1).Translate(vector.NewVector3(0, 3.5, 0))).
		SetMaterial(mat).
		Append(extrude.CircleWithThickness(20, domeThickness, domePath).SetMaterial(domeMat))
}

func main() {
	ufoOuterRadius := 10.
	ufoportalRadius := 4.
	ring := AbductionRing(ufoportalRadius, 0.5, 0.5)
	ringSpacing := vector.NewVector3(0, 3., 0)
	final := ring.
		Append(ring.
			Scale(vector.Vector3Zero(), vector.Vector3One().MultByConstant(.75)).
			Translate(ringSpacing.MultByConstant(1)).
			Rotate(mesh.UnitQuaternionFromTheta(0.3, vector.Vector3Down()))).
		Append(ring.
			Scale(vector.Vector3Zero(), vector.Vector3One().MultByConstant(.5)).
			Translate(ringSpacing.MultByConstant(2)).
			Rotate(mesh.UnitQuaternionFromTheta(0.5, vector.Vector3Down()))).
		Append(UfoBody(ufoOuterRadius, ufoportalRadius, 8).Translate(ringSpacing.MultByConstant(2.5)))

	mtlFile, err := os.Create("ufo.mtl")
	if err != nil {
		panic(err)
	}
	defer mtlFile.Close()

	objFile, err := os.Create("ufo.obj")
	if err != nil {
		panic(err)
	}
	defer objFile.Close()

	obj.WriteMesh(&final, "ufo.mtl", objFile)

	obj.WriteMaterials(&final, mtlFile)
}
