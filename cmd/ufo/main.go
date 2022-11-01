package main

import (
	"log"
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
	return extrude.ClosedCircleWithThickness(20, thickness, path)
}

func contour(positions []vector.Vector3) mesh.Mesh {
	return repeat.Circle(extrude.Circle(7, .3, positions), 8, 0)
}

func sideLights(numberOfLights int, radius float64) mesh.Mesh {
	light := primitives.Cylinder(16, 0.5, 0.5).
		Append(primitives.Cylinder(16, 0.25, 0.25).Translate(vector.NewVector3(0, .35, 0))).
		Rotate(mesh.UnitQuaternionFromTheta(-math.Pi/2, vector.Vector3Forward()))

	return repeat.Circle(light, numberOfLights, radius)
}

func UfoBody(outerRadius float64, portalRadius float64) mesh.Mesh {
	path := []vector.Vector3{
		vector.Vector3Up().MultByConstant(0),
		vector.Vector3Up().MultByConstant(1),

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
	for i := 0; i < domeResolution; i++ {
		percent := float64(i+1) / float64(domeResolution)

		height := math.Sin(percent*halfPi) * float64(domeHight)
		path = append(path, vector.Vector3Up().MultByConstant(height+domeStartHeight))

		cosResult := math.Cos(percent * halfPi)
		log.Println(cosResult)
		thickness = append(thickness, (cosResult * domeStartWidth))
	}

	return extrude.CircleWithThickness(20, thickness, path).
		Append(contour([]vector.Vector3{
			vector.NewVector3(thickness[2], path[2].Y(), 0),
			vector.NewVector3(thickness[3], path[3].Y(), 0),
			vector.NewVector3(thickness[4], path[4].Y(), 0),
			vector.NewVector3(thickness[5], path[5].Y(), 0),
		})).
		Append(primitives.
			Cylinder(20, 1, outerRadius+1).
			Translate(vector.NewVector3(0, 3.5, 0))).
		Append(sideLights(8, outerRadius+1).Translate(vector.NewVector3(0, 3.5, 0)))
}

func main() {
	ufoOuterRadius := 10.
	ufoportalRadius := 4.
	ring := AbductionRing(ufoportalRadius, 0.5, 0.5)
	ringSpacing := vector.NewVector3(0, 3., 0)
	final := ring.
		Append(ring.Translate(ringSpacing.MultByConstant(1)).Rotate(mesh.UnitQuaternionFromTheta(0.3, vector.Vector3Down()))).
		Append(ring.Translate(ringSpacing.MultByConstant(2)).Rotate(mesh.UnitQuaternionFromTheta(0.5, vector.Vector3Down()))).
		Append(UfoBody(ufoOuterRadius, ufoportalRadius).Translate(ringSpacing.MultByConstant(3)))

	f, err := os.Create("ufo.obj")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	obj.Write(&final, f)
}
