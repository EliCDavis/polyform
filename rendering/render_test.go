package rendering_test

import (
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/polyform/rendering/materials"
	"github.com/EliCDavis/polyform/rendering/textures"
	"github.com/EliCDavis/vector/vector3"
)

func randomScene() []rendering.Hittable {
	world := make([]rendering.Hittable, 0)

	ground_material := materials.NewLambertian(textures.NewSolidColorTexture(vector3.New(0.5, 0.5, 0.5)))
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
		err := rendering.Render(50, 50, 50, aspectRatio, randomScene(), camera, "example2.png", nil)
		if err != nil {
			panic(err)
		}
	}
}
