package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/polyform/rendering/materials"
	"github.com/EliCDavis/polyform/rendering/textures"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func randomBallsScene() []rendering.Hittable {
	world := make([]rendering.Hittable, 0)

	checkerPattern := textures.NewCheckerColorPattern(
		vector3.New(0.2, 0.3, 0.1),
		vector3.New(0.9, 0.9, 0.9),
	)
	ground_material := materials.NewLambertian(checkerPattern)
	world = append(world, rendering.NewSphere(vector3.New(0., -1000., 0.), 1000, ground_material))

	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			choose_mat := rand.Float64()
			center := vector3.New(
				float64(a)+(0.9*rand.Float64()),
				0.2,
				float64(b)+(0.9*rand.Float64()),
			)

			if center.Sub(vector3.New(4., 0.2, 0.)).Length() > 0.9 {
				var sphere_material rendering.Material = nil

				if choose_mat < 0.8 {
					// diffuse
					albedo := vector3.Rand().MultByVector(vector3.Rand())
					sphere_material = materials.NewLambertian(textures.NewSolidColorTexture(albedo))
					dir := vector3.RandNormal().Scale(rand.Float64())
					world = append(
						world,
						rendering.NewAnimatedSphere(
							0.2,
							sphere_material,
							func(t float64) vector3.Float64 {
								return center.Add(dir.Scale(t))
							},
						))
				} else if choose_mat < 0.95 {
					// metal
					albedo := vector3.RandRange(0.4, 1.)
					fuzz := rand.Float64() * 0.5
					sphere_material = materials.NewFuzzyMetal(albedo, fuzz)
					world = append(world, rendering.NewSphere(center, 0.2, sphere_material))

				} else {
					// glass
					sphere_material := materials.NewDielectric(1.5)
					world = append(world, rendering.NewSphere(center, 0.2, sphere_material))
				}
			}
		}
	}

	material1 := materials.NewDielectric(1.5)
	world = append(world, rendering.NewSphere(vector3.New(0., 1., 0.), 1.0, material1))

	material2 := materials.NewLambertian(textures.NewSolidColorTexture(vector3.New(0.4, 0.2, 0.1)))
	world = append(world, rendering.NewSphere(vector3.New(-4., 1., 0.), 1.0, material2))

	material3 := materials.NewFuzzyMetal(vector3.New(0.7, 0.6, 0.5), 0.0)
	world = append(world, rendering.NewSphere(vector3.New(4., 1., 0.), 1.0, material3))

	return world
}

func SimpleLightScene() []rendering.Hittable {
	world := make([]rendering.Hittable, 0)

	checkerPattern := textures.NewCheckerColorPattern(
		vector3.New(0.2, 0.3, 0.1),
		vector3.New(0.9, 0.9, 0.9),
	)
	ground_material := materials.NewLambertian(checkerPattern)
	world = append(world, rendering.NewSphere(vector3.New(0., -1000., 0.), 1000, ground_material))
	world = append(world, rendering.NewSphere(vector3.New(0., 2., 0.), 2, ground_material))

	diffuseLight := materials.NewDiffuseLightWithColor(vector3.New(4., 4., 4.))
	world = append(world, rendering.NewXYRectangle(vector2.New(3., 1.), vector2.New(4., 3.), -2., diffuseLight))
	return world
}

func videoScene(spheres int, radius float64) []rendering.Hittable {
	world := make([]rendering.Hittable, 0)

	checkerPattern := textures.NewCheckerColorPattern(
		vector3.New(0.2, 0.3, 0.1),
		vector3.New(0.9, 0.9, 0.9),
	)
	ground_material := materials.NewLambertian(checkerPattern)
	world = append(world, rendering.NewSphere(vector3.New(0., -1000., 0.), 1000, ground_material))

	bigSphereCenter := vector3.New(0., 2., 0.)
	world = append(world, rendering.NewSphere(bigSphereCenter, 2, materials.NewDielectric(1.5)))
	world = append(world, rendering.NewSphere(bigSphereCenter.Scale(0.9), -2, materials.NewDielectric(1.5)))

	angleInc := (math.Pi * 2.) / float64(spheres)
	for i := 0; i < spheres; i++ {
		matType := rand.Float64()

		angle := angleInc * float64(i)

		animationFunction := func(t float64) vector3.Float64 {
			adjustedAngle := angle + (t * 0.5 * math.Pi * 2.)
			return bigSphereCenter.Add(vector3.New(
				math.Sin(adjustedAngle)*radius,
				math.Sin(angle+(t*2*math.Pi*2.))*0.2,
				math.Cos(adjustedAngle)*radius,
			))
		}

		var sphereMaterial rendering.Material
		if matType < 0.4 {
			sphereMaterial = materials.NewLambertian(textures.NewCheckerPatternWithTilingRate(
				textures.NewSolidColorTexture(vector3.Rand().MultByVector(vector3.Rand())),
				textures.NewSolidColorTexture(vector3.Rand().MultByVector(vector3.Rand())),
				30,
			))
		} else if matType < 0.8 {
			albedo := vector3.Rand().Scale(0.5).Add(vector3.New(0.5, 0.5, 0.5))
			sphereMaterial = materials.NewFuzzyMetal(albedo, (rand.Float64()*0.5)+0.3)
		} else {
			albedo := vector3.
				Rand().
				Scale(0.2).
				Add(vector3.New(0.7, 0.7, 0.7)).
				Scale(4)
			sphereMaterial = materials.NewDiffuseLightWithColor(albedo)
		}

		world = append(world, rendering.NewAnimatedSphere(0.4, sphereMaterial, animationFunction))
	}

	// diffuseLight := materials.NewDiffuseLightWithColor(vector3.New(4., 4., 4.))
	// world = append(world, rendering.NewXYRectangle(vector2.New(3., 1.), vector2.New(4., 3.), -2., diffuseLight))
	return world
}

func simsScene(time float64) []rendering.Hittable {
	world := make([]rendering.Hittable, 0)

	jewelColor := vector3.New(0., 0.9, 0.4)
	var jewelMat rendering.Material = materials.NewFuzzyMetal(jewelColor, 0.1)
	// jewelMat = materials.NewBarycenstric()
	// jewelMat = materials.NewLambertian(textures.NewSolidColorTexture(vector3.New(0., 0.7, 0.3)))
	jewelMat = materials.NewDielectricWithColor(1.5, jewelColor)

	jewelGeometry := primitives.
		UVSphere(1, 2, 8).
		Scale(vector3.Zero[float64](), vector3.New(1., 2., 1.)).
		Translate(vector3.Up[float64]().Scale(2)).
		Rotate(modeling.UnitQuaternionFromTheta(time*0.5, vector3.Down[float64]())).
		Unweld().
		CalculateFlatNormals()

	world = append(world,
		rendering.NewMesh(
			jewelGeometry,
			jewelMat,
		),
	)

	world = append(world,
		rendering.NewMesh(
			jewelGeometry.Scale(vector3.Up[float64]().Scale(2), vector3.One[float64]().Scale(0.9)).FlipTriWinding(),
			jewelMat,
		),
	)

	checkerPattern := textures.NewCheckerColorPattern(
		vector3.New(0.2, 0.3, 0.1),
		vector3.New(0.9, 0.9, 0.9),
	)
	ground_material := materials.NewLambertian(checkerPattern)
	world = append(world, rendering.NewSphere(vector3.New(0., -1000., 0.), 1000, ground_material))

	return world
}

func bunnyScene() []rendering.Hittable {
	world := make([]rendering.Hittable, 0)

	jewelColor := vector3.New(0., 0.9, 0.4)
	var jewelMat rendering.Material = materials.NewFuzzyMetal(jewelColor, 0.1)
	jewelMat = materials.NewBarycentric()
	jewelMat = materials.NewLambertian(textures.NewSolidColorTexture(vector3.New(0.7, 0.7, 0.7)))
	jewelMat = materials.NewDielectricWithColor(1.5, jewelColor)

	bunny, err := obj.Load("test-models/stanford-bunny.obj")
	if err != nil {
		panic(err)
	}

	b1 := bunny.
		CenterFloat3Attribute(modeling.PositionAttribute).
		Scale(vector3.Zero[float64](), vector3.One[float64]().Scale(20)).
		Translate(vector3.Up[float64]().Scale(2)).
		CalculateSmoothNormals()

	world = append(world,
		rendering.NewMesh(
			b1,
			jewelMat,
		),
	)
	// world = append(world,
	// 	rendering.NewMesh(
	// 		b1.Scale(vector3.Zero[float64](), vector3.One[float64]().Scale(.95)),
	// 		jewelMat,
	// 	),
	// )

	checkerPattern := textures.NewCheckerColorPattern(
		vector3.New(0.2, 0.3, 0.1),
		vector3.New(0.9, 0.9, 0.9),
	)
	ground_material := materials.NewLambertian(checkerPattern)
	world = append(world, rendering.NewSphere(vector3.New(0., -1000., 0.), 1000, ground_material))

	return world
}

func main() {
	// origin := vector3.New(13., 2., 3.)
	origin := vector3.New(6., 6., 12.)
	lookat := vector3.New(0., 2., 0.)
	aperatre := 0.1

	aspectRatio := 3. / 2.

	background := func(v vector3.Float64) vector3.Float64 {
		return vector3.New(151./255., 178./255., 222./255.)
	}

	t1 := time.Now()

	fps := 24.
	duration := 2.
	timeInc := 1. / fps

	totalFrames := int(fps * duration)

	for i := 0; i < totalFrames; i++ {
		start := float64(i) * timeInc
		end := start + (timeInc * .25)

		camera := rendering.NewCamera(20., aspectRatio, aperatre, origin.Distance(lookat), origin, lookat, vector3.Up[float64](), start, end, background)

		completion := make(chan float64, 1)

		scene := simsScene(float64(i) / (math.Pi * 2.))

		go func() {
			err := rendering.RenderToFile(50, 200, 800, scene, camera, fmt.Sprintf("tmp/sims/frame_%06d.png", i), completion)
			if err != nil {
				log.Print(err)
				panic(err)
			}
		}()

		lastProgress := -1.
		for progress := range completion {
			if progress-lastProgress > .001 {
				lastProgress = progress
				log.Printf("Image %d Progress: %.2f%%\n", i, progress*100.)
			}
		}

	}

	totalTime := time.Since(t1)
	timePerFrame := time.Duration(int(totalTime) / totalFrames)

	fmt.Printf("Total time: %s; With an average frame time of: %s", totalTime, timePerFrame)
}
