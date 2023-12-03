package main

import (
	"image/color"
	"math"
	"os"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/extrude"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector3"
)

func AbductionRing(radius, baseThickness, magnitude float64) modeling.Mesh {
	pathSize := 120
	path := make([]vector3.Float64, pathSize)
	thickness := make([]float64, pathSize)

	angleIncrement := (1.0 / float64(pathSize)) * 2.0 * math.Pi
	for i := 0; i < pathSize; i++ {
		angle := angleIncrement * float64(i)
		path[i] = vector3.New(math.Cos(angle)*radius, math.Sin(angle*5)*magnitude, math.Sin(angle)*radius)
		thickness[i] = (math.Sin(angle*8) * magnitude * 0.25) + baseThickness
	}

	mat := modeling.Material{
		Name:              "Abduction Ring",
		DiffuseColor:      color.RGBA{0, 255, 0, 255},
		AmbientColor:      color.RGBA{0, 255, 0, 255},
		SpecularColor:     color.Black,
		SpecularHighlight: 0,
		OpticalDensity:    5,
	}
	return extrude.
		ClosedCircleWithThickness(20, thickness, path).
		SetMaterial(mat)
}

func contour(positions []vector3.Float64, times int) modeling.Mesh {
	return repeat.Circle(extrude.CircleWithConstantThickness(7, .3, positions), times, 0)
}

func sideLights(numberOfLights int, radius float64) modeling.Mesh {
	sides := 8
	light := primitives.Cylinder{Sides: sides, Height: 0.5, Radius: 0.5}.ToMesh().
		Append(primitives.Cylinder{Sides: sides, Height: 0.25, Radius: 0.25}.ToMesh().Transform(
			meshops.TranslateAttribute3DTransformer{
				Amount: vector3.New(0., .35, 0.),
			},
		)).
		Rotate(quaternion.FromTheta(-math.Pi/2, vector3.Forward[float64]()))

	return repeat.Circle(light, numberOfLights, radius)
}

func UfoBody(outerRadius float64, portalRadius float64, frameSections int) modeling.Mesh {
	path := []vector3.Float64{
		vector3.Up[float64]().Scale(-1),
		vector3.Up[float64]().Scale(2),

		vector3.Up[float64]().Scale(0.5),
		vector3.Up[float64]().Scale(3),
		vector3.Up[float64]().Scale(4),
		vector3.Up[float64]().Scale(5),

		vector3.Up[float64]().Scale(5.5),
		vector3.Up[float64]().Scale(5.5),
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
	domePath := make([]vector3.Float64, 0)
	domePath = append(domePath, path[len(path)-1])
	domeThickness := make([]float64, 0)
	domeThickness = append(domeThickness, thickness[len(thickness)-1])
	for i := 0; i < domeResolution; i++ {
		percent := float64(i+1) / float64(domeResolution)

		height := math.Sin(percent*halfPi) * float64(domeHight)
		domePath = append(domePath, vector3.Up[float64]().Scale(height+domeStartHeight))

		cosResult := math.Cos(percent * halfPi)
		domeThickness = append(domeThickness, (cosResult * domeStartWidth))
	}

	mat := modeling.Material{
		Name:              "UFO Body",
		DiffuseColor:      color.RGBA{128, 128, 128, 255},
		AmbientColor:      color.RGBA{128, 128, 128, 255},
		SpecularColor:     color.RGBA{128, 128, 128, 255},
		SpecularHighlight: 100,
		OpticalDensity:    1,
	}

	domeMat := modeling.Material{
		Name:              "UFO Dome",
		DiffuseColor:      color.RGBA{0, 0, 255, 255},
		AmbientColor:      color.RGBA{0, 0, 255, 255},
		SpecularColor:     color.RGBA{0, 0, 255, 255},
		SpecularHighlight: 100,
		Transparency:      0.2,
		OpticalDensity:    2,
	}

	return extrude.CircleWithThickness(20, thickness, path).
		Append(contour([]vector3.Float64{
			vector3.New(thickness[2], path[2].Y(), 0),
			vector3.New(thickness[3], path[3].Y(), 0),
			vector3.New(thickness[4], path[4].Y(), 0),
			vector3.New(thickness[5], path[5].Y(), 0),
		}, frameSections)).
		Append(primitives.
			Cylinder{
			Sides:  20,
			Height: 1,
			Radius: outerRadius + 1,
		}.ToMesh().
			Translate(vector3.New(0., 3.5, 0.))).
		Append(extrude.ClosedCircleWithConstantThickness(8, .25, repeat.CirclePoints(frameSections, portalRadius)).
			Translate(vector3.Up[float64]().Scale(0.5))).
		Append(sideLights(frameSections, outerRadius+1).Translate(vector3.New(0., 3.5, 0.))).
		SetMaterial(mat).
		Append(extrude.CircleWithThickness(20, domeThickness, domePath).SetMaterial(domeMat))
}

func main() {
	ufoOuterRadius := 10.
	ufoportalRadius := 4.
	ring := AbductionRing(ufoportalRadius, 0.5, 0.5)
	ringSpacing := vector3.New(0., 3., 0.)
	final := ring.
		Append(ring.
			Scale(vector3.Fill(.75)).
			Translate(ringSpacing.Scale(1)).
			Rotate(quaternion.FromTheta(0.3, vector3.Down[float64]()))).
		Append(ring.
			Scale(vector3.Fill(.5)).
			Translate(ringSpacing.Scale(2)).
			Rotate(quaternion.FromTheta(0.5, vector3.Down[float64]()))).
		Append(UfoBody(ufoOuterRadius, ufoportalRadius, 8).Translate(ringSpacing.Scale(2.5)))

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

	obj.WriteMesh(final, "ufo.mtl", objFile)
	obj.WriteMaterials(final, mtlFile)
}
