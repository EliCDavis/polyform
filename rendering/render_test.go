package rendering_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/polyform/rendering/materials"
	"github.com/EliCDavis/polyform/rendering/textures"
	"github.com/EliCDavis/vector/vector3"
)

func randomScene() []rendering.Hittable {
	world := make([]rendering.Hittable, 0)

	ground_material := materials.NewLambertian(textures.NewSolidColorTexture(vector3.New(0.5, 0.5, 0.5)))
	world = append(world, rendering.NewSphere(vector3.New(0., -1000., 0.), 1000, ground_material))
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

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
					albedo := vector3.Rand(r).MultByVector(vector3.Rand(r))
					sphere_material = materials.NewLambertian(textures.NewSolidColorTexture(albedo))
					dir := vector3.RandNormal(r).Scale(rand.Float64())
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
					albedo := vector3.RandRange(r, 0.4, 1.)
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

func bunnyScene() []rendering.Hittable {
	world := make([]rendering.Hittable, 0)

	var jewelMat rendering.Material = materials.NewFuzzyMetal(vector3.New(0., 0.9, 0.4), 0.1)
	jewelMat = materials.NewBarycentric()
	// jewelMat = materials.NewLambertian(textures.NewSolidColorTexture(vector3.New(0.7, 0.7, 0.7)))
	// jewelMat = materials.NewDielectric(1.5)

	bunny, err := obj.Load("../test-models/stanford-bunny.obj")
	if err != nil {
		panic(err)
	}

	world = append(world,
		rendering.NewMesh(
			bunny.Transform(
				meshops.CenterAttribute3DTransformer{},
				meshops.ScaleAttribute3DTransformer{
					Amount: vector3.Fill(20.),
				},
				meshops.TranslateAttribute3DTransformer{
					Amount: vector3.Up[float64]().Scale(2),
				},
				meshops.SmoothNormalsTransformer{},
			),
			jewelMat,
		),
	)

	// diffuseLight := materials.NewDiffuseLightWithColor(vector3.New(4., 4., 4.))
	// world = append(world, rendering.NewXYRectangle(vector2.New(3., 1.), vector2.New(4., 3.), -2., diffuseLight))

	checkerPattern := textures.NewCheckerColorPattern(
		vector3.New(0.2, 0.3, 0.1),
		vector3.New(0.9, 0.9, 0.9),
	)
	ground_material := materials.NewLambertian(checkerPattern)
	world = append(world, rendering.NewSphere(vector3.New(0., -1000., 0.), 1000, ground_material))
	// world = append(world, rendering.NewSphere(vector3.New(0., 2., 4.), 2, jewelMat))

	return world
}

func BenchmarkRender(b *testing.B) {
	origin := vector3.New(13., 2., 3.)
	lookat := vector3.New(0., 0., 0.)
	aperatre := 0.1

	aspectRatio := 3. / 2.

	camera := rendering.NewCamera(20., aspectRatio, aperatre, 10., origin, lookat, vector3.Up[float64](), 0, 1, func(v vector3.Float64) vector3.Float64 {
		return vector3.New(0., 0.5, 0.9)
	})

	for n := 0; n < b.N; n++ {
		// always record the result of Fib to prevent
		// the compiler eliminating the function call.
		err := rendering.RenderToFile(50, 50, 50, randomScene(), camera, "example2.png", nil)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkBunnyRender(b *testing.B) {
	origin := vector3.New(26., 6., 6.)
	lookat := vector3.New(0., 2., 0.)
	aperatre := 0.1

	aspectRatio := 3. / 2.

	camera := rendering.NewCamera(20., aspectRatio, aperatre, origin.Distance(lookat), origin, lookat, vector3.Up[float64](), 0, 1, func(v vector3.Float64) vector3.Float64 {
		return vector3.New(0., 0.5, 0.9)
	})

	scene := bunnyScene()

	for n := 0; n < b.N; n++ {
		// always record the result of Fib to prevent
		// the compiler eliminating the function call.
		err := rendering.RenderToFile(10, 20, 100, scene, camera, "example2.png", nil)
		if err != nil {
			panic(err)
		}
	}
}
